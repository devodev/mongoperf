package query

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InsertOneMeta .
type InsertOneMeta struct {
	Data    map[string]interface{}
	Options *options.InsertOneOptions
}

// InsertOneQuery .
type InsertOneQuery struct {
	config *Definition
	meta   *InsertOneMeta
}

// Run implements the Querier interface.
func (q *InsertOneQuery) Run(ctx context.Context, col *mongo.Collection) *Result {
	result := NewQueryResult(q.config)
	_, err := col.InsertOne(ctx, q.meta.Data, q.meta.Options)
	if err != nil {
		return result.WithError(err)
	}
	return result.WithResult(1)
}

// NewInsertOneQuery .
func NewInsertOneQuery(config *Definition) (Querier, error) {
	var meta InsertOneMeta
	if err := mapstructure.Decode(config.Meta, &meta); err != nil {
		return nil, err
	}
	if len(meta.Data) == 0 {
		return nil, fmt.Errorf("Data is empty")
	}
	if meta.Options == nil {
		meta.Options = options.InsertOne()
	}
	return &InsertOneQuery{config: config, meta: &meta}, nil
}
