package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/inbugay1/httprouter"
	"myfacebook-websocket-handler/internal/apiv1"
	"myfacebook-websocket-handler/internal/repository"
)

type Auth struct {
	next           httprouter.Handler
	userRepository repository.UserRepository
}

func (m *Auth) Handle(responseWriter http.ResponseWriter, request *http.Request) error {
	authHeader := request.Header.Get("Authorization")
	authHeaderParts := strings.Split(authHeader, "Bearer")

	if len(authHeaderParts) != 2 {
		return apiv1.NewInvalidTokenError("bearer token is missing", nil)
	}

	token := strings.TrimSpace(authHeaderParts[1])

	ctx := request.Context()

	user, err := m.userRepository.GetUserByToken(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return apiv1.NewInvalidTokenError("invalid token",
				fmt.Errorf("auth middleware, user with token %q not found: %w", token, err))
		}

		return apiv1.NewServerError(fmt.Errorf("auth middleware, failed to get user by token: %w", err))
	}

	ctx = context.WithValue(ctx, "user_id", user.ID) //nolint:revive,staticcheck

	err = m.next.Handle(responseWriter, request.WithContext(ctx))
	if err != nil {
		return err //nolint:wrapcheck
	}

	return nil
}

func NewAuth(userRepository repository.UserRepository) httprouter.MiddlewareFunc {
	return func(next httprouter.Handler) httprouter.Handler {
		return &Auth{
			userRepository: userRepository,
			next:           next,
		}
	}
}
