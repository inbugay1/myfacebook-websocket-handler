package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	amqp "github.com/rabbitmq/amqp091-go"
	"myfacebook-websocket-handler/internal/config"
	"myfacebook-websocket-handler/internal/rmq"
)

type ClientConfig struct {
	MessageMaxSize      int64
	ReadTimeoutSeconds  int
	WriteTimeoutSeconds int
	PingIntervalSeconds int
}

type Client struct {
	config    *ClientConfig
	userID    string
	envConfig *config.EnvConfig
	cancel    context.CancelFunc
	rmq       *rmq.RMQ
	mu        sync.Mutex
	conn      *websocket.Conn
}

type ClientFactory struct {
	config    *ClientConfig
	envConfig *config.EnvConfig
	rmq       *rmq.RMQ
}

func NewClientFactory(config *ClientConfig, rmq *rmq.RMQ,
	envConfig *config.EnvConfig) *ClientFactory {
	return &ClientFactory{
		config:    config,
		envConfig: envConfig,
		rmq:       rmq,
	}
}

func (f *ClientFactory) NewClient(conn *websocket.Conn, userID string, cancel context.CancelFunc) *Client {
	return &Client{
		config:    f.config,
		envConfig: f.envConfig,
		rmq:       f.rmq,
		userID:    userID,
		cancel:    cancel,
		conn:      conn,
	}
}

func (c *Client) ReadPump(_ context.Context) {
	defer func() {
		_ = c.conn.Close()
		c.cancel()
	}()

	c.conn.SetReadLimit(c.config.MessageMaxSize)

	err := c.conn.SetReadDeadline(time.Now().Add(time.Duration(c.config.ReadTimeoutSeconds) * time.Second))
	if err != nil {
		slog.Error(fmt.Sprintf("WS client failed to set read deadline: %s", err))

		return
	}

	c.conn.SetPongHandler(func(string) error {
		err = c.conn.SetReadDeadline(time.Now().Add(time.Duration(c.config.ReadTimeoutSeconds) * time.Second))
		if err != nil {
			slog.Error(fmt.Sprintf("WS client failed to set read deadline: %s", err))
		}

		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			slog.Debug(fmt.Sprintf("WS client failed to read message from ws: %s", err))

			return
		}
	}
}

func (c *Client) WritePump(ctx context.Context) {
	pingTicker := time.NewTicker(time.Duration(c.config.PingIntervalSeconds) * time.Second)

	defer pingTicker.Stop()

	msgs, err := c.consumeRMQMessages(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("WS client failed to consume rmq messages: %s", err))

		return
	}

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				slog.Error("WS client RMQ channel is closed")

				select {
				case <-time.After(1 * time.Second):
					newMsgs, err := c.consumeRMQMessages(ctx)
					if err != nil {
						slog.Error(fmt.Sprintf("WS client failed to consume rmq messages: %s", err))

						continue
					}

					msgs = newMsgs
				case <-ctx.Done():
					return
				}

				continue
			}

			serverMsg, err := c.processRMQMessage(ctx, msg.Body)
			if err != nil {
				slog.Error(fmt.Sprintf("WS client failed to process rmq message: %s", err))

				if err := msg.Reject(false); err != nil {
					slog.Error(fmt.Sprintf("WS client failed to reject rmq message: %s", err))
				}

				continue
			}

			if err := msg.Ack(false); err != nil {
				slog.Error(fmt.Sprintf("WS client failed to ack rmq message: %s", err))

				continue
			}

			err = c.WriteMessage(serverMsg)
			if err != nil {
				slog.Error(fmt.Sprintf("WS client failed to write server message to ws: %s", err))
			}
		case <-ctx.Done():
			slog.Debug("WS client exit from WritePump loop: context canceled")

			return
		case <-pingTicker.C:
			err := c.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Duration(c.config.WriteTimeoutSeconds)*time.Second))
			if err != nil {
				slog.Info(fmt.Sprintf("WS client failed to write PingMessage to ws: %s", err))

				return
			}
		}
	}
}

func (c *Client) consumeRMQMessages(ctx context.Context) (<-chan amqp.Delivery, error) {
	queue, err := c.rmq.DeclareQueue(rmq.Queue{
		Name:       "user_" + c.userID,
		AutoDelete: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to declare rmq queue: %w", err)
	}

	err = c.rmq.BindQueueToExchange(queue.Name, "/post/feed/posted", c.userID)
	if err != nil {
		return nil, fmt.Errorf("failed to bind queue to exchange: %w", err)
	}

	msgs, err := c.rmq.Consume(ctx, queue.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to consume rmq messages: %w", err)
	}

	return msgs, nil
}

func (c *Client) WriteMessage(serverMessage serverMessage) error {
	messageBytes, err := json.Marshal(serverMessage)
	if err != nil {
		return fmt.Errorf("unable to marshal server message %+v, err: %w", serverMessage, err)
	}

	err = c.writeMessage(messageBytes)
	if err != nil {
		return fmt.Errorf("failed to write message to ws: %w", err)
	}

	return nil
}

func (c *Client) writeMessage(message []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.conn.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(c.config.WriteTimeoutSeconds))); err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}

	if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
		return fmt.Errorf("failed to write TextMessage to ws: %w", err)
	}

	return nil
}

func (c *Client) processRMQMessage(_ context.Context, msg []byte) (serverMessage, error) {
	var serverMsg serverMessage

	err := json.Unmarshal(msg, &serverMsg)
	if err != nil {
		return serverMsg, fmt.Errorf("failed to unmarshal rmq message: %w", err)
	}

	return serverMsg, nil
}
