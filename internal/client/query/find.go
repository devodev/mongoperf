package query

import (
	"context"

	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FindMeta .
type FindMeta struct {
	Filter  map[string]interface{}
	Options *options.FindOptions
}

// FindQuery .
type FindQuery struct {
	config *Definition
	meta   *FindMeta
}

// Run implements the Querier interface.
func (q *FindQuery) Run(ctx context.Context, col *mongo.Collection) *Result {
	result := NewQueryResult(q.config)
	cur, err := col.Find(ctx, q.meta.Filter, q.meta.Options)
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
	return result.WithResult(len(results))
}

// NewFindQuery .
func NewFindQuery(config *Definition) (Querier, error) {
	var meta FindMeta
	if err := mapstructure.Decode(config.Meta, &meta); err != nil {
		return nil, err
	}
	if meta.Options == nil {
		meta.Options = options.Find()
	}
	return &FindQuery{config: config, meta: &meta}, nil
}
