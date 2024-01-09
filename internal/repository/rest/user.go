package rest

import (
	"context"
	"errors"
	"fmt"

	"myfacebook-websocket-handler/internal/myfacebookapiclient"
	"myfacebook-websocket-handler/internal/repository"
)

type UserRepository struct {
	apiClient *myfacebookapiclient.Client
}

func NewUserRepository(apiClient *myfacebookapiclient.Client) *UserRepository {
	return &UserRepository{
		apiClient: apiClient,
	}
}

func (r *UserRepository) GetUserByToken(ctx context.Context, token string) (*repository.User, error) {
	user, err := r.apiClient.GetUserByToken(ctx, token)
	if err != nil {
		if errors.Is(err, myfacebookapiclient.ErrNotFound) {
			return nil, repository.ErrNotFound
		}

		return nil, fmt.Errorf("userrepository failed to get user by token: %w", err)
	}

	return &repository.User{
		ID: user.ID,
	}, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, userID string) (*repository.User, error) {
	user, err := r.apiClient.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, myfacebookapiclient.ErrNotFound) {
			return nil, repository.ErrNotFound
		}

		return nil, fmt.Errorf("userrepository failed to get user by id: %w", err)
	}

	return &repository.User{
		ID: user.ID,
	}, nil
}
