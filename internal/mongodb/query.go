package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// QueryResult .
type QueryResult struct {
	Query *ScenarioQueryConfig

	Start       time.Time
	End         time.Time
	TotalChange int
	Error       error
}

// NewResult .
func NewResult(q *ScenarioQueryConfig) *QueryResult {
	return &QueryResult{
		Query: q,
		Start: time.Now(),
		End:   time.Time{},
	}
}

// WithError .
func (r *QueryResult) WithError(err error) *QueryResult {
	r.setEnd()
	r.Error = err
	return r
}

// WithResult .
func (r *QueryResult) WithResult(total int) *QueryResult {
	r.setEnd()
	r.TotalChange = total
	return r
}

func (r *QueryResult) setEnd() {
	r.End = time.Now()
}

func (c *Client) runQuery(ctx context.Context, collection *mongo.Collection, q *ScenarioQueryConfig) *QueryResult {
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
		c.logger.Debugf("query meta: %+v", meta)

		insertResult, err := collection.InsertOne(ctx, meta.Data, meta.Options)
		if err != nil {
			return result.WithError(err)
		}
		c.logger.Debugf("Inserted a single document: %v", insertResult.InsertedID)
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
		c.logger.Debugf("query meta: %+v", meta)

		insertManyResult, err := collection.InsertMany(ctx, meta.Data, meta.Options)
		if err != nil {
			return result.WithError(err)
		}
		c.logger.Debugf("Inserted multiple documents: %v", insertManyResult.InsertedIDs)
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
		c.logger.Debugf("query meta: %+v", meta)

		updateResult, err := collection.UpdateOne(ctx, meta.Filter, meta.Data, meta.Options)
		if err != nil {
			return result.WithError(err)
		}
		c.logger.Debugf("Matched %v documents and updated %v documents", updateResult.MatchedCount, updateResult.ModifiedCount)
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
		c.logger.Debugf("query meta: %+v", meta)

		var findResult map[string]interface{}
		err := collection.FindOne(ctx, meta.Filter, meta.Options).Decode(&findResult)
		if err != nil {
			return result.WithError(err)
		}
		c.logger.Debugf("Found a single document: %+v", findResult)
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
		c.logger.Debugf("query meta: %+v", meta)

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
		c.logger.Debugf("Found multiple documents (array of pointers): %+v", results)
		return result.WithResult(len(results))
	}
}
