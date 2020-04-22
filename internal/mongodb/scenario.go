package mongodb

import (
	"context"
	"sync"
	"time"
)

// ScenarioConfig .
type ScenarioConfig struct {
	Database   *string
	Collection *string
	Parallel   int
	BufferSize int
	Queries    []*ScenarioQueryConfig
}

// ScenarioQueryConfig .
type ScenarioQueryConfig struct {
	Name   *string
	Action *string
	Repeat *int
	Meta   map[string]interface{}
}

// RunScenario .
func (c *Client) RunScenario(ctx context.Context, s *ScenarioConfig, resultCh chan *QueryResult) error {
	client, cleanFn, err := c.connect(ctx)
	if err != nil {
		return err
	}
	defer cleanFn()

	collection := client.Database(*s.Database).Collection(*s.Collection)
	c.logger.Infof("using database: %v", *s.Database)
	c.logger.Infof("using collection: %v", *s.Collection)

	// Query Producer
	queryCh := make(chan *ScenarioQueryConfig, s.BufferSize)
	go func() {
		wg := &sync.WaitGroup{}
		wg.Add(len(s.Queries))

		for _, query := range s.Queries {
			go func(q *ScenarioQueryConfig) {
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
			}(query)
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
	wg.Add(s.Parallel)

	for i := 0; i < s.Parallel; i++ {
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
