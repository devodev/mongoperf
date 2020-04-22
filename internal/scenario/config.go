package scenario

import (
	"fmt"
	"io/ioutil"
	"mongoperf/internal/scenario/query"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Scenario .
type Scenario struct {
	Database   *string       `yaml:"Database"`
	Collection *string       `yaml:"Collection"`
	Parallel   *int          `yaml:"Parallel,omitempty"`
	BufferSize *int          `yaml:"BufferSize,omitempty"`
	Queries    []query.Query `yaml:"Queries"`
}

// UnmarshalYAML implements the yaml.Unmarshaller interface.
func (c *Scenario) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type C Scenario
	newConfig := (*C)(c)
	if err := unmarshal(&newConfig); err != nil {
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

// ParseConfigFile returns a Config from parsing the file
// using the provided filepath.
func ParseConfigFile(fp string) (*Scenario, error) {
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
func ParseConfig(b []byte) (*Scenario, error) {
	var c Scenario
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// Int .
func Int(i int) *int {
	return &i
}
