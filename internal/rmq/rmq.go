package rmq

import (
	"context"
	"fmt"
	"net"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
}

type Exchange struct {
	Name string
	Kind string
}

type Queue struct {
	Name       string
	Durable    bool
	AutoDelete bool
}

type RMQ struct {
	config    *Config
	exchanges []Exchange
	queues    []Queue

	connMU sync.RWMutex
	conn   *amqp.Connection

	channelMU sync.RWMutex
	channel   *amqp.Channel
}

func New(config *Config, exchanges []Exchange, queues []Queue) *RMQ {
	return &RMQ{
		config:    config,
		exchanges: exchanges,
		queues:    queues,
	}
}

func (rmq *RMQ) Connect(ctx context.Context) error {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s/", rmq.config.Username, rmq.config.Password, net.JoinHostPort(rmq.config.Host, rmq.config.Port)))
	if err != nil {
		return fmt.Errorf("rmq failed to connect to %s:%s: %w", rmq.config.Host, rmq.config.Port, err)
	}

	channel, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("rmq failed to open channel: %w", err)
	}

	rmq.setConnection(conn)
	rmq.setChannel(channel)

	if err := rmq.Ping(ctx); err != nil {
		return err
	}

	if err := rmq.declareExchangesAndQueues(); err != nil {
		return fmt.Errorf("rmq failed to decalare echanges and queues: %w", err)
	}

	return nil
}

func (rmq *RMQ) setConnection(conn *amqp.Connection) {
	rmq.connMU.Lock()
	defer rmq.connMU.Unlock()

	rmq.conn = conn
}

func (rmq *RMQ) getConnection() *amqp.Connection {
	rmq.connMU.RLock()
	defer rmq.connMU.RUnlock()

	return rmq.conn
}

func (rmq *RMQ) setChannel(channel *amqp.Channel) {
	rmq.channelMU.Lock()
	defer rmq.channelMU.Unlock()

	rmq.channel = channel
}

func (rmq *RMQ) getChannel() *amqp.Channel {
	rmq.channelMU.RLock()
	defer rmq.channelMU.RUnlock()

	return rmq.channel
}

func (rmq *RMQ) declareExchangesAndQueues() error {
	for _, exchange := range rmq.exchanges {
		err := rmq.DeclareExchange(exchange)
		if err != nil {
			return err
		}
	}

	for _, queue := range rmq.queues {
		_, err := rmq.DeclareQueue(queue)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rmq *RMQ) DeclareExchange(exchange Exchange) error {
	err := rmq.getChannel().ExchangeDeclare(
		exchange.Name,
		exchange.Kind,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("rmq failed do declare exchange %q: %w", exchange.Name, err)
	}

	return nil
}

func (rmq *RMQ) DeclareQueue(queue Queue) (amqp.Queue, error) {
	var amqpQueue amqp.Queue

	amqpQueue, err := rmq.getChannel().QueueDeclare(
		queue.Name,
		queue.Durable,
		queue.AutoDelete,
		false,
		false,
		nil,
	)
	if err != nil {
		return amqpQueue, fmt.Errorf("rmq failed to declare queue: %w", err)
	}

	return amqpQueue, nil
}

func (rmq *RMQ) DeleteQueue(queueName string) error {
	_, err := rmq.getChannel().QueueDelete(
		queueName,
		false,
		false,
		false,
	)
	if err != nil {
		return fmt.Errorf("rmq failed to delete queue: %w", err)
	}

	return nil
}

func (rmq *RMQ) Publish(ctx context.Context, exchangeName, routingKey string, message []byte) error {
	err := rmq.getChannel().PublishWithContext(ctx, exchangeName, routingKey, false, false, amqp.Publishing{
		ContentType: "text/json",
		Body:        message,
	})
	if err != nil {
		return fmt.Errorf("rmq failed to publish message to exchange %q: %w", exchangeName, err)
	}

	return nil
}

func (rmq *RMQ) BindQueueToExchange(queueName, exchangeName, routingKey string) error {
	err := rmq.getChannel().QueueBind(
		queueName,
		routingKey,
		exchangeName,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("rmq failed to bind queue %q to exchange %q: %w", queueName, exchangeName, err)
	}

	return nil
}

func (rmq *RMQ) Consume(_ context.Context, queueName string) (<-chan amqp.Delivery, error) {
	msgs, err := rmq.getChannel().Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("rmq failed to consume messages from queue %q: %w", queueName, err)
	}

	return msgs, nil
}

func (rmq *RMQ) Disconnect() error {
	if err := rmq.getChannel().Close(); err != nil {
		return fmt.Errorf("rmq failed to close channel: %w", err)
	}

	if err := rmq.getConnection().Close(); err != nil {
		return fmt.Errorf("rmq failed to close connection: %w", err)
	}

	return nil
}

func (rmq *RMQ) Ping(ctx context.Context) error {
	err := rmq.getChannel().PublishWithContext(ctx, "", "", false, false,
		amqp.Publishing{
			DeliveryMode: amqp.Transient,
			Body:         []byte("test"),
		})
	if err != nil {
		return fmt.Errorf("rmq failed to ping: %w", err)
	}

	return nil
}

func (rmq *RMQ) Reconnect(ctx context.Context) error {
	_ = rmq.Disconnect()

	err := rmq.Connect(ctx)
	if err != nil {
		return fmt.Errorf("rmq failed to reconnect: %w", err)
	}

	return nil
}
