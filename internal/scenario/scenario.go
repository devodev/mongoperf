package scenario

import (
	"context"
	"mongoperf/internal/scenario/query"
	"sync"
	"time"
)

// RunScenario .
func (c *Client) RunScenario(ctx context.Context, config *Scenario, resultCh chan *query.Result) error {
	collection := c.client.Database(*config.Database).Collection(*config.Collection)
	c.logger.Infof("using database: %v", *config.Database)
	c.logger.Infof("using collection: %v", *config.Collection)

	// Query Producer
	queryCh := make(chan *query.Query, *config.BufferSize)
	go func() {
		wg := &sync.WaitGroup{}
		wg.Add(len(config.Queries))

		for _, qq := range config.Queries {
			go func(q *query.Query) {
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
			}(&qq)
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

			for config := range queryCh {
				c.logger.Debugf("received query name: %v action: %v", *config.Name, *config.Action)

				querier, err := query.NewQuerier(config)
				if err != nil {
					c.logger.Error(err)
					continue
				}
				resultCh <- querier.Run(context.TODO(), collection)
			}
		}()
	}

	wg.Wait()
	close(resultCh)
	return nil
}
