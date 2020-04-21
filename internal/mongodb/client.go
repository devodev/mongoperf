package mongodb

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Trainer .
type Trainer struct {
	Name string
	Age  int
	City string
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

// RunDemo .
func (c *Client) RunDemo(ctx context.Context, db, col string) error {
	// Set client options
	clientOptions := options.Client().ApplyURI(c.uri)

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return err
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return err
	}

	// clean when stopping
	defer func() {
		err = client.Disconnect(context.TODO())
		if err != nil {
			c.logger.Println(err)
		}
		c.logger.Println("Connection to MongoDB closed.")
	}()

	c.logger.Println("Connected to MongoDB!")

	// Get a collection
	collection := client.Database(db).Collection(col)

	// Declare test models
	ash := Trainer{"Ash", 10, "Pallet Town"}
	misty := Trainer{"Misty", 10, "Cerulean City"}
	brock := Trainer{"Brock", 15, "Pewter City"}

	// Insert one
	insertResult, err := collection.InsertOne(context.TODO(), ash)
	if err != nil {
		return err
	}
	c.logger.Println("Inserted a single document: ", insertResult.InsertedID)

	// Insert multiple
	trainers := []interface{}{misty, brock}
	insertManyResult, err := collection.InsertMany(context.TODO(), trainers)
	if err != nil {
		return err
	}
	c.logger.Println("Inserted multiple documents: ", insertManyResult.InsertedIDs)

	// Update one
	filter := bson.D{{Key: "name", Value: "Ash"}}
	update := bson.D{{Key: "$inc", Value: bson.D{{Key: "age", Value: 1}}}}
	updateResult, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	c.logger.Printf("Matched %v documents and updated %v documents.\n", updateResult.MatchedCount, updateResult.ModifiedCount)

	// Find one
	var result Trainer
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
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
	cur, err := collection.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		return err
	}

	// Finding multiple documents returns a cursor
	// Iterating through the cursor allows us to decode documents one at a time
	for cur.Next(context.TODO()) {

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
	cur.Close(context.TODO())

	c.logger.Printf("Found multiple documents (array of pointers): %+v\n", results)

	// Delete all
	deleteResult, err := collection.DeleteMany(context.TODO(), bson.D{{}})
	if err != nil {
		return err
	}
	c.logger.Printf("Deleted %v documents in the trainers collection\n", deleteResult.DeletedCount)
	return nil
}
