package mongodb

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Action .
type Action string

// Action enum .
const (
	InsertOneAction  Action = "InsertOne"
	InsertManyAction        = "InsertMany"
	UpdateOneAction         = "UpdateOneAction"
	FindOneAction           = "FindOneAction"
	FindAction              = "Find"
)

// UnmarshalYAML implements the yaml.Unmarshaller interface.
func (a *Action) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshal(a); err != nil {
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

// Config .
type Config struct {
	Database   *string               `yaml:"Database"`
	Collection *string               `yaml:"Collection"`
	Parallel   *int                  `yaml:"Parallel,omitempty"`
	BufferSize *int                  `yaml:"BufferSize,omitempty"`
	Queries    []ScenarioQueryConfig `yaml:"Queries"`
}

// UnmarshalYAML implements the yaml.Unmarshaller interface.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshal(c); err != nil {
		return err
	}
	if c.Database == nil {
		return fmt.Errorf("Database must not be empty")
	}
	if c.Collection == nil {
		return fmt.Errorf("Collection must not be empty")
	}
	switch p := c.Parallel; {
	case p == nil:
		c.Parallel = Int(1)
	case *p < 1:
		return fmt.Errorf("Parallel must be greater or equal to 1")
	default:
	}
	switch s := c.BufferSize; {
	case s == nil:
		c.BufferSize = Int(1000)
	case *s < 1:
		return fmt.Errorf("BufferSize must be greater or equal to 1")
	default:
	}
	if len(c.Queries) == 0 {
		return fmt.Errorf("Queries must not be empty")
	}
	return nil
}

// ConfigQuery .
type ConfigQuery struct {
	Name   *string                `yaml:"Name"`
	Action *Action                `yaml:"Action"`
	Repeat *int                   `yaml:"Repeat,omitempty"`
	Meta   map[string]interface{} `yaml:"Meta"`
}

// UnmarshalYAML implements the yaml.Unmarshaller interface.
func (c *ConfigQuery) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshal(c); err != nil {
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

// ParseConfigFile returns a Config from parsing the file
// using the provided filepath.
func ParseConfigFile(fp string) (*Config, error) {
	filename, err := filepath.Abs(fp)
	if err != nil {
		return nil, err
	}
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ParseConfig(f)
}

// ParseConfig returns a Config from bytes.
func ParseConfig(b []byte) (*Config, error) {
	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// Int .
func Int(i int) *int {
	return &i
}
