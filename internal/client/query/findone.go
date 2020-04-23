package query

import (
	"context"

	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FindOneMeta .
type FindOneMeta struct {
	Filter  map[string]interface{}
	Options *options.FindOneOptions
}

// FindOneQuery .
type FindOneQuery struct {
	config *Definition
	meta   *FindOneMeta
}

// Run implements the Querier interface.
func (q *FindOneQuery) Run(ctx context.Context, col *mongo.Collection) *Result {
	result := NewQueryResult(q.config)
	findoneResult := col.FindOne(ctx, q.meta.Filter, q.meta.Options)
	if findoneResult.Err() != nil {
		if findoneResult.Err() != mongo.ErrNoDocuments {
			return result.WithError(findoneResult.Err())
		}
		return result.WithResult(0)
	}
	return result.WithResult(1)
}

// NewFindOneQuery .
func NewFindOneQuery(config *Definition) (Querier, error) {
	var meta FindOneMeta
	if err := mapstructure.Decode(config.Meta, &meta); err != nil {
		return nil, err
	}
	if meta.Options == nil {
		meta.Options = options.FindOne()
	}
	return &FindOneQuery{config: config, meta: &meta}, nil
}
