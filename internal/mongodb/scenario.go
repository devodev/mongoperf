package mongodb

import (
	"context"
	"sync"
	"time"
)

// RunScenario .
func (c *Client) RunScenario(ctx context.Context, config *Config, resultCh chan *QueryResult) error {
	collection := c.client.Database(*config.Database).Collection(*config.Collection)
	c.logger.Infof("using database: %v", *config.Database)
	c.logger.Infof("using collection: %v", *config.Collection)

	// Query Producer
	queryCh := make(chan *ConfigQuery, *config.BufferSize)
	go func() {
		wg := &sync.WaitGroup{}
		wg.Add(len(config.Queries))

		for _, query := range config.Queries {
			go func(q *ConfigQuery) {
				defer wg.Done()

				switch r := *q.Repeat; {
				default:
					for i := 0; i < r; i++ {
						select {
						case queryCh <- q:
						case <-ctx.Done():
							return
						}
					}
				case r < 1:
					for {
						select {
						case queryCh <- q:
						case <-ctx.Done():
							return
						}
					}
				}
			}(&query)
		}
		wg.Wait()
		// wait until all query are consumed
		for {
			if len(queryCh) == 0 {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		close(queryCh)
	}()

	// Query Consumer
	wg := &sync.WaitGroup{}
	wg.Add(*config.Parallel)

	for i := 0; i < *config.Parallel; i++ {
		go func() {
			defer wg.Done()

			for query := range queryCh {
				c.logger.Infof("received query name: %v action: %v", *query.Name, *query.Action)

				res := c.runQuery(context.TODO(), collection, query)
				if res.Error != nil {
					c.logger.Error(res.Error)
				}
				resultCh <- res
			}
		}()
	}

	wg.Wait()
	close(resultCh)
	return nil
}
