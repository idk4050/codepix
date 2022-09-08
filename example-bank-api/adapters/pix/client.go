package pix

import (
	"bytes"
	"codepix/example-bank-api/config"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"
	"github.com/golang-jwt/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	Conn        *grpc.ClientConn
	logger      logr.Logger
	credentials *Credentials
}

func Open(config config.Config, logger logr.Logger) (*Client, error) {
	cfg := config.PixAPI
	logger = logger.WithName("pix.api")

	credentials := &Credentials{
		APIKey:        cfg.APIKey,
		TokenEndpoint: cfg.TokenEndpoint,
		token:         &atomic.Value{},
	}
	conn, err := grpc.Dial(cfg.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(credentials),
	)
	if err != nil {
		return nil, err
	}
	logger.Info("Pix API client opened")

	client := &Client{
		Conn:        conn,
		logger:      logger,
		credentials: credentials,
	}
	return client, nil
}

func (c *Client) Close() error {
	err := c.Conn.Close()
	if err != nil {
		c.logger.Error(err, "Pix API failed to close")
		return err
	}
	c.logger.Info("Pix API connection closed")
	return nil
}

func (c *Client) RefreshCredentials(ctx context.Context) {
	for {
		token, expiry, err := fetchNewToken(c.credentials.APIKey, c.credentials.TokenEndpoint)
		if err != nil {
			c.logger.Error(err, "failed to refresh token")
		} else {
			c.credentials.token.Store(token)
			c.logger.Info("token refreshed")
		}
		time.Sleep(time.Until(expiry) - time.Minute)
	}
}

func fetchNewToken(APIKey, endpoint string) (string, time.Time, error) {
	reply, err := http.Post(endpoint, "text/plain", bytes.NewReader([]byte(APIKey)))
	if err != nil {
		return "", time.Time{}, err
	}
	body, err := io.ReadAll(reply.Body)
	if err != nil {
		return "", time.Time{}, err
	}
	if reply.StatusCode != http.StatusOK {
		return "", time.Time{}, fmt.Errorf("%s: %s", reply.Status, string(body))
	}
	claims := jwt.MapClaims{}
	_, _, err = (&jwt.Parser{}).ParseUnverified(string(body), claims)
	if err != nil {
		return "", time.Time{}, err
	}
	var expiry time.Time
	switch exp := claims["exp"].(type) {
	case float64:
		expiry = time.Unix(int64(exp), 0)
	case json.Number:
		v, _ := exp.Int64()
		expiry = time.Unix(v, 0)
	default:
		return "", time.Time{}, errors.New("no expiry")
	}
	return string(body), expiry, nil
}

type Credentials struct {
	APIKey        string
	TokenEndpoint string
	token         *atomic.Value
}

func (c *Credentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	token, ok := c.token.Load().(string)
	if !ok {
		return nil, errors.New("no token set")
	}
	return map[string]string{
		"authorization": token,
	}, nil
}

func (c *Credentials) RequireTransportSecurity() bool {
	return false
}
