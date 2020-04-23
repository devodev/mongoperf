package query

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InsertManyMeta .
type InsertManyMeta struct {
	Data    []interface{}
	Options *options.InsertManyOptions
}

// InsertManyQuery .
type InsertManyQuery struct {
	config *Definition
	meta   *InsertManyMeta
}

// Run implements the Querier interface.
func (q *InsertManyQuery) Run(ctx context.Context, col *mongo.Collection) *Result {
	if q == nil {
		panic("InsertManyQuery is nil")
	}
	result := NewQueryResult(q.config)
	insertManyResult, err := col.InsertMany(ctx, q.meta.Data, q.meta.Options)
	if err != nil {
		return result.WithError(err)
	}
	return result.WithResult(len(insertManyResult.InsertedIDs))
}

// NewInsertManyQuery .
func NewInsertManyQuery(config *Definition) (Querier, error) {
	var meta InsertManyMeta
	if err := mapstructure.Decode(config.Meta, &meta); err != nil {
		return nil, err
	}
	if len(meta.Data) == 0 {
		return nil, fmt.Errorf("Data is empty")
	}
	if meta.Options == nil {
		meta.Options = options.InsertMany()
	}
	return &InsertManyQuery{config: config, meta: &meta}, nil
}
