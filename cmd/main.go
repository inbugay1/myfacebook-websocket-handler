package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/inbugay1/httprouter"
	"myfacebook-websocket-handler/internal/apiclient"
	"myfacebook-websocket-handler/internal/apiv1/handler"
	apiv1middleware "myfacebook-websocket-handler/internal/apiv1/middleware"
	"myfacebook-websocket-handler/internal/config"
	"myfacebook-websocket-handler/internal/connectionwatcher"
	"myfacebook-websocket-handler/internal/httpclient"
	httproutermiddleware "myfacebook-websocket-handler/internal/httprouter/middleware"
	"myfacebook-websocket-handler/internal/httpserver"
	"myfacebook-websocket-handler/internal/myfacebookapiclient"
	"myfacebook-websocket-handler/internal/repository/rest"
	"myfacebook-websocket-handler/internal/rmq"
	"myfacebook-websocket-handler/internal/ws"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application error: %s", err)
	}
}

func run() error {
	envConfig := config.GetConfigFromEnv()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel(envConfig.LogLevel),
	}))
	slog.SetDefault(logger)

	rabbitMQ := rmq.New(&rmq.Config{
		Host:     envConfig.RMQHost,
		Port:     envConfig.RMQPort,
		Username: envConfig.RMQUsername,
		Password: envConfig.RMQPassword,
	}, []rmq.Exchange{
		{
			Name: "/post/feed/posted",
			Kind: "direct",
		},
	}, nil)

	ctx := context.Background()

	if err := rabbitMQ.Connect(ctx); err != nil {
		return fmt.Errorf("cannot connect to rmq: %w", err)
	}

	slog.Info(fmt.Sprintf("Successfully connected to rmq on %s", net.JoinHostPort(envConfig.RMQHost, envConfig.RMQPort)))

	defer func() {
		if err := rabbitMQ.Disconnect(); err != nil {
			log.Fatalf("Failed to disconnect from rmq: %s", err)
		}
	}()

	connectionWatcher := connectionwatcher.New(&connectionwatcher.Config{
		PingInterval:     time.Duration(envConfig.ConnectionWatcherPingIntervalSeconds) * time.Second,
		PingTimeout:      time.Duration(envConfig.ConnectionWatcherPingTimeoutSeconds) * time.Second,
		ReconnectTimeout: time.Duration(envConfig.ConnectionWatcherReconnectTimeoutSeconds) * time.Second,
	})

	connectionWatcher.AddService("rmq", rabbitMQ)

	connectionWatcher.Start(ctx)
	defer connectionWatcher.Stop()

	httpClient := httpclient.New(&httpclient.Config{
		InsecureSkipVerify: true,
	})

	apiClient := apiclient.New(envConfig.MyfacebookAPIBaseURL, httpClient)

	myfacebookAPIClient := myfacebookapiclient.New(apiClient)

	userRepository := rest.NewUserRepository(myfacebookAPIClient)

	router := httprouter.New()

	requestResponseMiddleware := httproutermiddleware.NewRequestResponseLog()

	apiV1ErrorResponseMiddleware := apiv1middleware.NewErrorResponse()
	apiV1ErrorLogMiddleware := apiv1middleware.NewErrorLog()
	apiV1AuthMiddleware := apiv1middleware.NewAuth(userRepository)

	webSocketUpgrader := &websocket.Upgrader{
		ReadBufferSize:  envConfig.WebSocketUpgraderReadBufferSize,
		WriteBufferSize: envConfig.WebSocketUpgraderWriteBufferSize,
	}
	if !envConfig.WebSocketUpgraderCheckOrigin {
		webSocketUpgrader.CheckOrigin = func(r *http.Request) bool { return true }
	}

	wsClientFactory := ws.NewClientFactory(&ws.ClientConfig{
		MessageMaxSize:      int64(envConfig.WebSocketMaxMessageSize),
		ReadTimeoutSeconds:  envConfig.WebSocketReadTimeoutSeconds,
		WriteTimeoutSeconds: envConfig.WebSocketWriteTimeoutSeconds,
		PingIntervalSeconds: envConfig.WebSocketPingIntervalSeconds,
	}, rabbitMQ, envConfig)

	router.Group(func(router httprouter.Router) {
		router.Use(
			apiV1ErrorResponseMiddleware,
			apiV1ErrorLogMiddleware,
			apiV1AuthMiddleware,
		)

		router.Get("/ws", &handler.WebsocketHandler{
			Upgrader:        webSocketUpgrader,
			EnvConfig:       envConfig,
			WSClientFactory: wsClientFactory,
			UserRepository:  userRepository,
		}, "/ws")
	})

	router.Use(requestResponseMiddleware)

	httpServer := httpserver.New(httpserver.Config{
		Port:                          envConfig.HTTPPort,
		RequestMaxHeaderBytes:         envConfig.RequestHeaderMaxSize,
		ReadHeaderTimeoutMilliseconds: envConfig.RequestReadHeaderTimeoutMilliseconds,
	}, router)

	httpServerErrCh := httpServer.Start()
	defer httpServer.Shutdown()

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	select {
	case osSignal := <-osSignals:
		slog.Info(fmt.Sprintf("got signal from OS: %v. Exit...", osSignal))
	case err := <-httpServerErrCh:
		return fmt.Errorf("http server error: %w", err)
	}

	return nil
}

func logLevel(lvl string) slog.Level {
	switch lvl {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	}

	return slog.LevelInfo
}
