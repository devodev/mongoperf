package mongodb

import (
	"context"
	"fmt"

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
	Meta   *ScenarioMeta
}

// ScenarioMeta .
type ScenarioMeta struct {
	Payload     bson.M
	PayloadList bson.A
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
		c.logger.Printf("running query #%d: %q", idx+1, *query.Name)
		c.runQuery(ctx, collection, query)
	}
	return nil
}

func (c *Client) runQuery(ctx context.Context, collection *mongo.Collection, q *ScenarioQuery) error {
	switch a := q.Action; {
	default:
		return fmt.Errorf("scenario action not supported: %v", a)
	case *a == "InsertOne":
		doc := q.Meta.Payload

		c.logger.Printf("inserting: %+v", doc)
		insertResult, err := collection.InsertOne(ctx, doc)
		if err != nil {
			return err
		}
		c.logger.Println("Inserted a single document: ", insertResult.InsertedID)
	case *a == "InsertMany":
		docs := q.Meta.PayloadList

		c.logger.Printf("inserting: %+v", docs)
		insertManyResult, err := collection.InsertMany(ctx, docs)
		if err != nil {
			return err
		}
		c.logger.Println("Inserted multiple documents: ", insertManyResult.InsertedIDs)
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
