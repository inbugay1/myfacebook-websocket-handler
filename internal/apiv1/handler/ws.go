package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"myfacebook-websocket-handler/internal/apiv1"
	"myfacebook-websocket-handler/internal/config"
	"myfacebook-websocket-handler/internal/repository"
	"myfacebook-websocket-handler/internal/ws"
)

var errUserIDTypeAssertionFailed = errors.New("failed to assert user_id to string")

type WebsocketHandler struct {
	Upgrader        *websocket.Upgrader
	EnvConfig       *config.EnvConfig
	WSClientFactory *ws.ClientFactory
	UserRepository  repository.UserRepository
}

func (h *WebsocketHandler) Handle(responseWriter http.ResponseWriter, request *http.Request) error {
	ctx := request.Context()

	conn, err := h.Upgrader.Upgrade(responseWriter, request, nil)
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("websocket handler failed to upgrade http to ws: %w", err))
	}

	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return apiv1.NewServerError(errUserIDTypeAssertionFailed)
	}

	ctx, cancel := context.WithCancel(context.Background())
	wsClient := h.WSClientFactory.NewClient(conn, userID, cancel)

	go wsClient.ReadPump(ctx)  //nolint:contextcheck
	go wsClient.WritePump(ctx) //nolint:contextcheck

	return nil
}
