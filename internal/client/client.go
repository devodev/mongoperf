package client

import (
	"context"
	"mongoperf/internal/client/query"
	"sync"

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

// New returns a new Client using the provided URI.
func New(ctx context.Context, uri string, options ...Option) (*Client, error) {
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

// RunScenario .
func (c *Client) RunScenario(ctx context.Context, scenario *Scenario) (map[string]*ReportAggregator, error) {
	collection := c.client.Database(*scenario.Database).Collection(*scenario.Collection)
	c.logger.Infof("using database: %v", *scenario.Database)
	c.logger.Infof("using collection: %v", *scenario.Collection)

	bufferSize := *scenario.BufferSize
	numConsumers := *scenario.Parallel
	numIteration := *scenario.Repeat

	wg := &sync.WaitGroup{}
	wg.Add(numConsumers)

	closing := make(chan struct{})
	closed := make(chan struct{})
	dataCh := make(chan query.Querier, bufferSize)
	resultCh := make(chan *query.Result, 0)

	results := make(map[string]*ReportAggregator)

	// stop signals the producer to stop sending
	// it can be called multiple times
	stop := func() {
		select {
		case closing <- struct{}{}:
			<-closed
		case <-closed:
		}
	}

	// context cancellation handler
	go func() {
		select {
		case <-ctx.Done():
			stop()
		}
	}()

	// 1 Producer
	go func() {
		defer func() {
			close(closed)
			close(dataCh)
		}()
		var queriers []query.Querier
		for _, def := range scenario.Queries {
			querier, err := query.NewQuerier(&def)
			if err != nil {
				c.logger.Error(err)
				continue
			}
			queriers = append(queriers, querier)
			c.logger.Debugf("registered query %v with action %v", *def.Name, *def.Action)
		}

		loops := 0
		for {
			for _, q := range queriers {
				select {
				default:
				case <-closing:
					return
				}

				select {
				case dataCh <- q:
				case <-closing:
					return
				}
			}
			loops++
			if loops == numIteration {
				return
			}
		}
	}()

	// N Consumers
	for i := 0; i < numConsumers; i++ {
		go func() {
			defer wg.Done()
			for querier := range dataCh {
				resultCh <- querier.Run(context.TODO(), collection)
			}
		}()
	}

	// Add/Update ReportQuery
	wgResults := &sync.WaitGroup{}
	wgResults.Add(1)
	go func(r map[string]*ReportAggregator) {
		defer wgResults.Done()
		for result := range resultCh {
			rq, ok := r[*result.Definition.Name]
			if !ok {
				rq = NewReportQuery(result.Definition, numConsumers)
			}
			delta := result.End.Sub(result.Start)
			rq.Update(delta, result.TotalChange, result.Error)
			r[*result.Definition.Name] = rq
		}
	}(results)

	wg.Wait()
	close(resultCh)
	wgResults.Wait()

	return results, nil
}
