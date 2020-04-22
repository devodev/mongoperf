package mongodb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client .
type Client struct {
	uri    string
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
func NewClient(uri string, options ...Option) *Client {
	c := &Client{uri: uri}
	for _, opt := range options {
		opt(c)
	}
	return c
}

func (c *Client) connect(ctx context.Context) (*mongo.Client, func() error, error) {
	// Set client options
	clientOptions := options.Client().ApplyURI(c.uri)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, nil, err
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	// Create closure for cleanup
	close := func() error {
		err = client.Disconnect(ctx)
		if err != nil {
			c.logger.Error(err)
			return err
		}
		c.logger.Info("Connection to MongoDB closed.")
		return nil
	}
	return client, close, nil
}

// Result .
type Result struct {
	Query *ScenarioQuery

	Start       time.Time
	End         time.Time
	TotalChange int
	Error       error
}

// NewResult .
func NewResult(q *ScenarioQuery) *Result {
	return &Result{
		Query: q,
		Start: time.Now(),
		End:   time.Time{},
	}
}

// WithError .
func (r *Result) WithError(err error) *Result {
	r.setEnd()
	r.Error = err
	return r
}

// WithResult .
func (r *Result) WithResult(total int) *Result {
	r.setEnd()
	r.TotalChange = total
	return r
}

func (r *Result) setEnd() {
	r.End = time.Now()
}

// Scenario .
type Scenario struct {
	Database   *string
	Collection *string
	Parallel   int
	Queries    []*ScenarioQuery
}

// ScenarioQuery .
type ScenarioQuery struct {
	Name   *string
	Action *string
	Meta   map[string]interface{}
}

// RunScenario .
func (c *Client) RunScenario(ctx context.Context, s *Scenario, resultCh chan *Result) error {
	client, cleanFn, err := c.connect(ctx)
	if err != nil {
		return err
	}
	defer cleanFn()

	collection := client.Database(*s.Database).Collection(*s.Collection)
	c.logger.Infof("using database: %v", *s.Database)
	c.logger.Infof("using collection: %v", *s.Collection)

	// Query Producer
	queryCh := make(chan *ScenarioQuery)
	go func() {
		for _, query := range s.Queries {
			queryCh <- query
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

				res := c.runQuery(ctx, collection, query)
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

func (c *Client) runQuery(ctx context.Context, collection *mongo.Collection, q *ScenarioQuery) *Result {
	result := NewResult(q)
	switch a := q.Action; {
	default:
		return result.WithError(fmt.Errorf("scenario action not supported: %v", *a))
	case *a == "InsertOne":
		payload := q.Meta
		if payload == nil {
			return result.WithError(fmt.Errorf("Meta is nil"))
		}

		type metaInsertOne struct {
			Data    map[string]interface{}
			Options *options.InsertOneOptions
		}

		var meta metaInsertOne
		if err := mapstructure.Decode(payload, &meta); err != nil {
			return result.WithError(err)
		}
		if len(meta.Data) == 0 {
			return result.WithError(fmt.Errorf("Data is empty"))
		}
		if meta.Options == nil {
			meta.Options = options.InsertOne()
		}
		c.logger.Infof("query meta: %+v", meta)

		insertResult, err := collection.InsertOne(ctx, meta.Data, meta.Options)
		if err != nil {
			return result.WithError(err)
		}
		c.logger.Infof("Inserted a single document: %v", insertResult.InsertedID)
		return result.WithResult(1)
	case *a == "InsertMany":
		payload := q.Meta
		if payload == nil {
			return result.WithError(fmt.Errorf("Meta is nil"))
		}

		type metaInsertMany struct {
			Data    []interface{}
			Options *options.InsertManyOptions
		}

		var meta metaInsertMany
		if err := mapstructure.Decode(payload, &meta); err != nil {
			return result.WithError(err)
		}
		if len(meta.Data) == 0 {
			return result.WithError(fmt.Errorf("Data is empty"))
		}
		if meta.Options == nil {
			meta.Options = options.InsertMany()
		}
		c.logger.Infof("query meta: %+v", meta)

		insertManyResult, err := collection.InsertMany(ctx, meta.Data, meta.Options)
		if err != nil {
			return result.WithError(err)
		}
		c.logger.Infof("Inserted multiple documents: %v", insertManyResult.InsertedIDs)
		return result.WithResult(len(insertManyResult.InsertedIDs))
	case *a == "UpdateOne":
		payload := q.Meta
		if payload == nil {
			return result.WithError(fmt.Errorf("Meta is nil"))
		}

		type metaUpdateOne struct {
			Data    map[string]interface{}
			Filter  map[string]interface{}
			Options *options.UpdateOptions
		}

		var meta metaUpdateOne
		if err := mapstructure.Decode(payload, &meta); err != nil {
			return result.WithError(err)
		}
		if len(meta.Data) == 0 {
			return result.WithError(fmt.Errorf("Data is empty"))
		}
		if meta.Options == nil {
			meta.Options = options.Update()
		}
		c.logger.Infof("query meta: %+v", meta)

		updateResult, err := collection.UpdateOne(ctx, meta.Filter, meta.Data, meta.Options)
		if err != nil {
			return result.WithError(err)
		}
		c.logger.Infof("Matched %v documents and updated %v documents", updateResult.MatchedCount, updateResult.ModifiedCount)
		return result.WithResult(int(updateResult.ModifiedCount))
	case *a == "FindOne":
		payload := q.Meta
		if payload == nil {
			return result.WithError(fmt.Errorf("Meta is nil"))
		}

		type metaFindOne struct {
			Filter  map[string]interface{}
			Options *options.FindOneOptions
		}

		var meta metaFindOne
		if err := mapstructure.Decode(payload, &meta); err != nil {
			return result.WithError(err)
		}
		if meta.Options == nil {
			meta.Options = options.FindOne()
		}
		c.logger.Infof("query meta: %+v", meta)

		var findResult map[string]interface{}
		err := collection.FindOne(ctx, meta.Filter, meta.Options).Decode(&findResult)
		if err != nil {
			return result.WithError(err)
		}
		c.logger.Infof("Found a single document: %+v", findResult)
		return result.WithResult(1)
	case *a == "Find":
		payload := q.Meta
		if payload == nil {
			return result.WithError(fmt.Errorf("Meta is nil"))
		}

		type metaFindMany struct {
			Filter  map[string]interface{}
			Options *options.FindOptions
		}

		var meta metaFindMany
		if err := mapstructure.Decode(payload, &meta); err != nil {
			return result.WithError(err)
		}
		if meta.Options == nil {
			meta.Options = options.Find()
		}
		c.logger.Infof("query meta: %+v", meta)

		cur, err := collection.Find(ctx, meta.Filter, meta.Options)
		if err != nil {
			return result.WithError(err)
		}
		defer cur.Close(ctx)

		var results []interface{}
		for cur.Next(ctx) {
			var elem interface{}
			err := cur.Decode(&elem)
			if err != nil {
				return result.WithError(err)
			}
			results = append(results, &elem)
		}
		if err := cur.Err(); err != nil {
			return result.WithError(err)
		}
		c.logger.Infof("Found multiple documents (array of pointers): %+v", results)
		return result.WithResult(len(results))
	}
}

// RunDemo .
func (c *Client) RunDemo(ctx context.Context, db, col string) error {
	client, close, err := c.connect(ctx)
	if err != nil {
		return err
	}
	defer close()

	c.logger.Infof("Connected to MongoDB!")

	// Get a collection
	collection := client.Database(db).Collection(col)

	type Trainer struct {
		Name string
		Age  int
		City string
	}
	// Declare test models
	ash := Trainer{"Ash", 10, "Pallet Town"}
	misty := Trainer{"Misty", 10, "Cerulean City"}
	brock := Trainer{"Brock", 15, "Pewter City"}

	// Insert one
	insertResult, err := collection.InsertOne(ctx, ash)
	if err != nil {
		return err
	}
	c.logger.Infof("Inserted a single document: ", insertResult.InsertedID)

	// Insert multiple
	trainers := []interface{}{misty, brock}
	insertManyResult, err := collection.InsertMany(ctx, trainers)
	if err != nil {
		return err
	}
	c.logger.Infof("Inserted multiple documents: ", insertManyResult.InsertedIDs)

	// Update one
	filter := bson.D{{Key: "name", Value: "Ash"}}
	update := bson.D{{Key: "$inc", Value: bson.D{{Key: "age", Value: 1}}}}
	updateResult, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	c.logger.Infof("Matched %v documents and updated %v documents.\n", updateResult.MatchedCount, updateResult.ModifiedCount)

	// Find one
	var result Trainer
	err = collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return err
	}
	c.logger.Infof("Found a single document: %+v\n", result)

	// Find multiple
	// Pass these options to the Find method
	findOptions := options.Find()
	findOptions.SetLimit(2)

	// Here's an array in which you can store the decoded documents
	var results []*Trainer

	// Passing bson.D{{}} as the filter matches all documents in the collection
	cur, err := collection.Find(ctx, bson.D{{}}, findOptions)
	if err != nil {
		return err
	}

	// Finding multiple documents returns a cursor
	// Iterating through the cursor allows us to decode documents one at a time
	for cur.Next(ctx) {

		// create a value into which the single document can be decoded
		var elem Trainer
		err := cur.Decode(&elem)
		if err != nil {
			return err
		}

		results = append(results, &elem)
	}

	if err := cur.Err(); err != nil {
		return err
	}

	// Close the cursor once finished
	cur.Close(ctx)

	c.logger.Infof("Found multiple documents (array of pointers): %+v\n", results)

	// Delete all
	deleteResult, err := collection.DeleteMany(ctx, bson.D{{}})
	if err != nil {
		return err
	}
	c.logger.Infof("Deleted %v documents in the trainers collection\n", deleteResult.DeletedCount)
	return nil
}
