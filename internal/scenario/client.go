package scenario

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client .
type Client struct {
	client *mongo.Client
	logger *logrus.Logger
}

// Option .
type Option func(*Client)

// WithLogger sets the provided logger on the client.
func WithLogger(l *logrus.Logger) func(c *Client) {
	if l != nil {
		return func(c *Client) {
			c.logger = l
		}
	}
	return func(c *Client) {}
}

// NewClient returns a new Client using the provided URI.
func NewClient(ctx context.Context, uri string, options ...Option) (*Client, error) {
	c := &Client{}
	err := c.connect(ctx, uri)
	if err != nil {
		return nil, err
	}
	for _, opt := range options {
		opt(c)
	}
	return c, nil
}

func (c *Client) connect(ctx context.Context, uri string) error {
	// Set client options
	clientOptions := options.Client().ApplyURI(uri)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return err
	}
	c.client = client
	return nil
}

// Close closes the underlying mongodb client connection.
func (c *Client) Close(ctx context.Context) error {
	err := c.client.Disconnect(ctx)
	if err != nil {
		return err
	}
	return nil
}
