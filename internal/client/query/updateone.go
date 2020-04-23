package query

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UpdateOneMeta .
type UpdateOneMeta struct {
	Data    map[string]interface{}
	Filter  map[string]interface{}
	Options *options.UpdateOptions
}

// UpdateOneQuery .
type UpdateOneQuery struct {
	config *Definition
	meta   *UpdateOneMeta
}

// Run implements the Querier interface.
func (q *UpdateOneQuery) Run(ctx context.Context, col *mongo.Collection) *Result {
	result := NewQueryResult(q.config)
	updateResult, err := col.UpdateOne(ctx, q.meta.Filter, q.meta.Data, q.meta.Options)
	if err != nil {
		return result.WithError(err)
	}
	return result.WithResult(int(updateResult.ModifiedCount))
}

// NewUpdateOneQuery .
func NewUpdateOneQuery(config *Definition) (Querier, error) {
	var meta UpdateOneMeta
	if err := mapstructure.Decode(config.Meta, &meta); err != nil {
		return nil, err
	}
	if len(meta.Data) == 0 {
		return nil, fmt.Errorf("Data is empty")
	}
	if meta.Options == nil {
		meta.Options = options.Update()
	}
	return &UpdateOneQuery{config: config, meta: &meta}, nil
}
