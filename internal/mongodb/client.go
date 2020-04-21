package mongodb

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Scenario .
type Scenario struct {
	Database   *string
	Collection *string
	Queries    []*ScenarioQuery
}

// ScenarioQuery .
type ScenarioQuery struct {
	Name   *string
	Action *string
	Meta   map[string]interface{}
}

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

// New returns a new Client using the provided URI.
func New(uri string, options ...Option) *Client {
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
			c.logger.Println(err)
			return err
		}
		c.logger.Println("Connection to MongoDB closed.")
		return nil
	}
	return client, close, nil
}

// RunScenario .
func (c *Client) RunScenario(ctx context.Context, s *Scenario) error {
	client, close, err := c.connect(ctx)
	if err != nil {
		return err
	}
	defer close()

	collection := client.Database(*s.Database).Collection(*s.Collection)
	c.logger.Printf("using database: %v", *s.Database)
	c.logger.Printf("using collection: %v", *s.Collection)

	for idx, query := range s.Queries {
		c.logger.Printf("running query #%d", idx+1)
		err := c.runQuery(ctx, collection, query)
		if err != nil {
			c.logger.Error(err)
		}
	}
	return nil
}

func (c *Client) runQuery(ctx context.Context, collection *mongo.Collection, q *ScenarioQuery) error {
	c.logger.Printf("scenario name: %v", *q.Name)
	c.logger.Printf("scenario action: %v", *q.Action)

	switch a := q.Action; {
	default:
		return fmt.Errorf("scenario action not supported: %v", a)
	case *a == "InsertOne":
		payload := q.Meta
		if payload == nil {
			return fmt.Errorf("Meta is nil")
		}

		data, ok := payload["Data"]
		if !ok {
			return fmt.Errorf("Data attribute is mandatory in Payload")
		}

		opts := options.InsertOne()
		dataOpts, ok := payload["Options"]
		if ok {
			if err := mapstructure.Decode(dataOpts, opts); err != nil {
				return fmt.Errorf("Options attribute is of wrong type")
			}
		}
		c.logger.Printf("%v data: %+v", *a, data)
		c.logger.Printf("%v options: %+v", *a, opts)

		insertResult, err := collection.InsertOne(ctx, data, opts)
		if err != nil {
			return err
		}
		c.logger.Println("Inserted a single document: ", insertResult.InsertedID)
	case *a == "InsertMany":
		payload := q.Meta
		if payload == nil {
			return fmt.Errorf("Meta is nil")
		}

		data, ok := payload["Data"]
		if !ok {
			return fmt.Errorf("Data attribute is mandatory in Payload")
		}
		dataSlice, ok := data.([]interface{})
		if !ok {
			return fmt.Errorf("Data attribute is of wrong type")
		}

		opts := options.InsertMany()
		dataOpts, ok := payload["Options"]
		if ok {
			if err := mapstructure.Decode(dataOpts, opts); err != nil {
				return fmt.Errorf("Options attribute is of wrong type")
			}
		}
		c.logger.Printf("%v data: %+v", *a, data)
		c.logger.Printf("%v options: %+v", *a, opts)

		insertManyResult, err := collection.InsertMany(ctx, dataSlice, opts)
		if err != nil {
			return err
		}
		c.logger.Println("Inserted multiple documents: ", insertManyResult.InsertedIDs)
	case *a == "UpdateOne":
		payload := q.Meta
		if payload == nil {
			return fmt.Errorf("Meta is nil")
		}

		data, ok := payload["Data"]
		if !ok {
			return fmt.Errorf("Update attribute is mandatory in Payload")
		}
		filter, ok := payload["Filter"]
		if !ok {
			return fmt.Errorf("Filter attribute is mandatory in Payload")
		}

		opts := options.Update()
		dataOpts, ok := payload["Options"]
		if ok {
			if err := mapstructure.Decode(dataOpts, opts); err != nil {
				return fmt.Errorf("Options attribute is of wrong type")
			}
		}
		c.logger.Printf("%v data: %+v", *a, data)
		c.logger.Printf("%v filter: %+v", *a, filter)
		c.logger.Printf("%v options: %+v", *a, opts)

		updateResult, err := collection.UpdateOne(ctx, filter, data, opts)
		if err != nil {
			return err
		}
		c.logger.Printf("Matched %v documents and updated %v documents.\n", updateResult.MatchedCount, updateResult.ModifiedCount)
	case *a == "FindOne":
		payload := q.Meta
		if payload == nil {
			return fmt.Errorf("Meta is nil")
		}

		filter, ok := payload["Filter"]
		if !ok {
			return fmt.Errorf("Filter attribute is mandatory in Payload")
		}

		opts := options.FindOne()
		dataOpts, ok := payload["Options"]
		if ok {
			if err := mapstructure.Decode(dataOpts, opts); err != nil {
				return fmt.Errorf("Options attribute is of wrong type")
			}
		}
		c.logger.Printf("%v filter: %+v", *a, filter)
		c.logger.Printf("%v options: %+v", *a, opts)

		var result map[string]interface{}
		err := collection.FindOne(ctx, filter, opts).Decode(&result)
		if err != nil {
			return err
		}
		c.logger.Printf("Found a single document: %+v\n", result)
	case *a == "Find":
		payload := q.Meta
		if payload == nil {
			return fmt.Errorf("Meta is nil")
		}

		filter, ok := payload["Filter"]
		if !ok {
			return fmt.Errorf("Filter attribute is mandatory in Payload")
		}

		opts := options.Find()
		dataOpts, ok := payload["Options"]
		if ok {
			if err := mapstructure.Decode(dataOpts, opts); err != nil {
				return fmt.Errorf("Options attribute is of wrong type")
			}
		}
		c.logger.Printf("%v filter: %+v", *a, filter)
		c.logger.Printf("%v options: %+v", *a, opts)

		cur, err := collection.Find(ctx, filter, opts)
		if err != nil {
			return err
		}
		defer cur.Close(ctx)

		var results []interface{}
		for cur.Next(ctx) {
			var elem interface{}
			err := cur.Decode(&elem)
			if err != nil {
				return err
			}
			results = append(results, &elem)
		}
		if err := cur.Err(); err != nil {
			return err
		}
		c.logger.Printf("Found multiple documents (array of pointers): %+v\n", results)
	}
	return nil
}

// RunDemo .
func (c *Client) RunDemo(ctx context.Context, db, col string) error {
	client, close, err := c.connect(ctx)
	if err != nil {
		return err
	}
	defer close()

	c.logger.Println("Connected to MongoDB!")

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
	c.logger.Println("Inserted a single document: ", insertResult.InsertedID)

	// Insert multiple
	trainers := []interface{}{misty, brock}
	insertManyResult, err := collection.InsertMany(ctx, trainers)
	if err != nil {
		return err
	}
	c.logger.Println("Inserted multiple documents: ", insertManyResult.InsertedIDs)

	// Update one
	filter := bson.D{{Key: "name", Value: "Ash"}}
	update := bson.D{{Key: "$inc", Value: bson.D{{Key: "age", Value: 1}}}}
	updateResult, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	c.logger.Printf("Matched %v documents and updated %v documents.\n", updateResult.MatchedCount, updateResult.ModifiedCount)

	// Find one
	var result Trainer
	err = collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return err
	}
	c.logger.Printf("Found a single document: %+v\n", result)

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

	c.logger.Printf("Found multiple documents (array of pointers): %+v\n", results)

	// Delete all
	deleteResult, err := collection.DeleteMany(ctx, bson.D{{}})
	if err != nil {
		return err
	}
	c.logger.Printf("Deleted %v documents in the trainers collection\n", deleteResult.DeletedCount)
	return nil
}
