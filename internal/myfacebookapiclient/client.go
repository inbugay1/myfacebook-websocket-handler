package myfacebookapiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	endpointFindUserByToken = "/int/user/findByToken" //nolint:gosec
	endpointGetUserByID     = "/int/user/%s"
)

type HTTPAPIClient interface {
	Get(ctx context.Context, path string) (*http.Response, error)
	Post(ctx context.Context, path string, body interface{}) (*http.Response, error)
}

type Client struct {
	apiClient HTTPAPIClient
}

func New(apiClient HTTPAPIClient) *Client {
	return &Client{
		apiClient: apiClient,
	}
}

func (c *Client) GetUserByToken(ctx context.Context, token string) (*User, error) {
	response, err := c.apiClient.Get(ctx, fmt.Sprintf("%s/%s", endpointFindUserByToken, token))
	if err != nil {
		return nil, fmt.Errorf("myfacebookapiclient failed to get user by token: %w", err)
	}

	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusNotFound:
		return nil, ErrNotFound
	case http.StatusOK:
		var user User

		err = json.NewDecoder(response.Body).Decode(&user)
		if err != nil {
			return nil, fmt.Errorf("myfacebookapiclient failed to decode api client response: %w", err)
		}

		return &user, nil
	}

	return nil, ErrUnexpectedStatusCode
}

func (c *Client) GetUserByID(ctx context.Context, userID string) (*User, error) {
	response, err := c.apiClient.Get(ctx, fmt.Sprintf(endpointGetUserByID, userID))
	if err != nil {
		return nil, fmt.Errorf("myfacebookapiclient failed to get user by id: %w", err)
	}

	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusNotFound:
		return nil, ErrNotFound
	case http.StatusOK:
		var user User

		err = json.NewDecoder(response.Body).Decode(&user)
		if err != nil {
			return nil, fmt.Errorf("myfacebookapiclient failed to decode api client response: %w", err)
		}

		return &user, nil
	}

	return nil, ErrUnexpectedStatusCode
}
