package connectionwatcher

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type Pinger interface {
	Ping(ctx context.Context) error
	Reconnect(ctx context.Context) error
}

type Config struct {
	PingInterval     time.Duration
	PingTimeout      time.Duration
	ReconnectTimeout time.Duration
}

type Watcher struct {
	config        *Config
	services      map[string]Pinger
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	mu            sync.Mutex
	errorMap      map[string]error
	lastCheckedAt map[string]time.Time
}

type Status struct {
	Status        bool
	Error         error
	LastCheckedAt time.Time
}

func New(config *Config) *Watcher {
	return &Watcher{
		services:      make(map[string]Pinger),
		errorMap:      make(map[string]error),
		lastCheckedAt: make(map[string]time.Time),
		config:        config,
	}
}

func (w *Watcher) AddService(name string, service Pinger) {
	w.services[name] = service
	w.errorMap[name] = nil
}

func (w *Watcher) Start(ctx context.Context) map[string]<-chan struct{} {
	slog.Info("Starting watcher...")

	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	w.cancel = cancel

	heartbeats := make(map[string]<-chan struct{})
	for serviceName, service := range w.services {
		heartbeats[serviceName] = w.doWatch(ctx, serviceName, service)
	}

	slog.Info("Watcher started successfully")

	return heartbeats
}

func (w *Watcher) doWatch(ctx context.Context, serviceName string, service Pinger) <-chan struct{} {
	heartbeat := make(chan struct{})

	w.wg.Add(1)
	go w.watch(ctx, serviceName, service, heartbeat)

	return heartbeat
}

func (w *Watcher) watch(ctx context.Context, serviceName string, service Pinger, heartbeat chan struct{}) {
	defer w.wg.Done()
	defer close(heartbeat)

	ticker := time.NewTicker(w.config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			select {
			case heartbeat <- struct{}{}:
			default: // nobody listen
			}

			w.checkAndReconnect(ctx, serviceName, service)
		}
	}
}

func (w *Watcher) checkAndReconnect(ctx context.Context, serviceName string, service Pinger) {
	pingCtx, cancel := context.WithTimeout(ctx, w.config.PingTimeout)
	defer cancel()

	err := service.Ping(pingCtx)
	if err == nil {
		w.updateConnectionStatus(serviceName, err)

		return
	}

	reconnectCtx, cancel := context.WithTimeout(ctx, w.config.ReconnectTimeout)
	defer cancel()

	err = service.Reconnect(reconnectCtx)
	if err != nil {
		slog.Error(fmt.Sprintf("%s failed to reconnect: %s", serviceName, err))
	} else {
		slog.Info(fmt.Sprintf("%s reconnected", serviceName))
	}

	w.updateConnectionStatus(serviceName, err)
}

func (w *Watcher) updateConnectionStatus(serviceName string, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.errorMap[serviceName] = err
	w.lastCheckedAt[serviceName] = time.Now()
}

func (w *Watcher) GetConnectionStatuses() map[string]Status {
	w.mu.Lock()
	defer w.mu.Unlock()

	statuses := make(map[string]Status)

	for serviceName, lastCheckedAt := range w.lastCheckedAt {
		statuses[serviceName] = Status{
			Status:        w.errorMap[serviceName] == nil,
			Error:         w.errorMap[serviceName],
			LastCheckedAt: lastCheckedAt,
		}
	}

	return statuses
}

func (w *Watcher) Stop() {
	slog.Info("Stopping watcher...")
	w.cancel()
	w.wg.Wait()
	slog.Info("Watcher stopped successfully")
}
