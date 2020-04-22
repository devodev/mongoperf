package mongodb

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
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

// RunDemo .
func (c *Client) RunDemo(ctx context.Context, db, col string) error {
	// Get a collection
	collection := c.client.Database(db).Collection(col)

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
