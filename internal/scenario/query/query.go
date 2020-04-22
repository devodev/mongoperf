package query

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

// Query .
type Query struct {
	Name   *string                `yaml:"Name"`
	Action *Action                `yaml:"Action"`
	Repeat *int                   `yaml:"Repeat,omitempty"`
	Meta   map[string]interface{} `yaml:"Meta"`
}

// UnmarshalYAML implements the yaml.Unmarshaller interface.
func (c *Query) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type C Query
	newConfig := (*C)(c)
	if err := unmarshal(&newConfig); err != nil {
		return err
	}
	if c.Name == nil {
		return fmt.Errorf("Name must not be empty")
	}
	switch r := c.Repeat; {
	case r == nil:
		c.Repeat = Int(1)
	case *r < 0:
		return fmt.Errorf("Repeat must be greater or equal to 0")
	default:
	}
	return nil
}

// Action .
type Action string

// UnmarshalYAML implements the yaml.Unmarshaller interface.
func (a *Action) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type A Action
	newAction := (*A)(a)
	if err := unmarshal(&newAction); err != nil {
		return err
	}
	if a == nil {
		return fmt.Errorf("Action must not be empty")
	}
	switch *a {
	case InsertOneAction, InsertManyAction, UpdateOneAction, FindOneAction, FindAction:
		return nil
	}
	return fmt.Errorf("Action not supported")
}

// Action enum .
const (
	InsertOneAction  Action = "InsertOne"
	InsertManyAction        = "InsertMany"
	UpdateOneAction         = "UpdateOneAction"
	FindOneAction           = "FindOneAction"
	FindAction              = "Find"
)

// Querier .
type Querier interface {
	Run(context.Context, *mongo.Collection) *Result
}

// NewQuerier .
func NewQuerier(config *Query) (Querier, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if config.Action == nil {
		return nil, fmt.Errorf("config.Action is nil")
	}
	if config.Meta == nil {
		return nil, fmt.Errorf("config.Meta is nil")
	}
	var (
		querier Querier
		err     error
	)
	switch *config.Action {
	default:
		return nil, fmt.Errorf("action not supported")
	case InsertOneAction:
		querier, err = NewInsertOneQuery(config)
	case InsertManyAction:
		querier, err = NewInsertManyQuery(config)
	case UpdateOneAction:
		querier, err = NewUpdateOneQuery(config)
	case FindOneAction:
		querier, err = NewFindOneQuery(config)
	case FindAction:
		querier, err = NewFindQuery(config)
	}
	return querier, err
}

// Result .
type Result struct {
	Query *Query

	Start       time.Time
	End         time.Time
	TotalChange int
	Error       error
}

// NewQueryResult .
func NewQueryResult(q *Query) *Result {
	return &Result{
		Query: q,
		Start: time.Now(),
		End:   time.Time{},
	}
}

// WithError .
func (r *Result) WithError(err error) *Result {
	r.setEnd()
	r.Error = err
	return r
}

// WithResult .
func (r *Result) WithResult(total int) *Result {
	r.setEnd()
	r.TotalChange = total
	return r
}

func (r *Result) setEnd() {
	r.End = time.Now()
}

// Int .
func Int(i int) *int {
	return &i
}
