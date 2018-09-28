package usermode

import (
	"github.com/scylladb/scylla-bench/random"

	yaml "gopkg.in/yaml.v2"
)

// Profile represents a benchmarking schema that is used to drive
// a database benchmark.
//
// For more details on the profile see the following document:
//
//   https://cassandra.apache.org/doc/latest/tools/cassandra_stress.html#profile
//
type Profile struct {
	Keyspace           string
	KeyspaceDefinition string
	Table              string
	TableDefinition    string
	ColumnSpec         []ColumnSpec
	Insert             *Insert
}

// ColumnSpec describes a metadata of a column.
type ColumnSpec struct {
	Name       string
	Size       random.Distribution
	Population random.Distribution
	Cluster    random.Distribution
}

// DefaultColumnSpec describes default metadata for a column.
var DefaultColumnSpec = &ColumnSpec{
	Cluster: &random.Fixed{Value: 1},
	Population: &random.Uniform{
		Min: 1,
		Max: 1000000000,
	},
	Size: &random.Uniform{
		Min: 4,
		Max: 8,
	},
}

// Insert describes the way data is populated across columns.
type Insert struct {
	BatchType  string
	Partitions random.Distribution
	Visits     random.Distribution
	Select     *random.Ratio
}

// see https://git.io/fACz6
func newDefaultInsert() *Insert {
	return &Insert{
		BatchType:  "UNLOGGED",
		Partitions: &random.Fixed{Value: 1},
		Visits:     &random.Fixed{Value: 1},
		Select: &random.Ratio{
			Distribution: &random.Fixed{Value: 1},
			Value:        1,
		},
	}
}

// ParseProfile parses an YAML profile.
func ParseProfile(p []byte) (*Profile, error) {
	var raw struct {
		Keyspace           string `yaml:"keyspace"`
		KeyspaceDefinition string `yaml:"keyspace_definition"`
		Table              string `yaml:"table"`
		TableDefinition    string `yaml:"table_definition"`
		ColumnSpec         []struct {
			Name       string `yaml:"name"`
			Size       string `yaml:"size"`
			Population string `yaml:"population"`
			Cluster    string `yaml:"cluster"`
		} `yaml:"columnspec"`
		Insert *struct {
			Partitions string `yaml:"partitions"`
			Select     string `yaml:"select"`
			Visits     string `yaml:"visits"`
			BatchType  string `yaml:"batchtype"`
		} `yaml:"insert"`
	}

	if err := yaml.Unmarshal(p, &raw); err != nil {
		return nil, err
	}

	q := &Profile{
		Keyspace:           raw.Keyspace,
		KeyspaceDefinition: raw.KeyspaceDefinition,
		Table:              raw.Table,
		TableDefinition:    raw.TableDefinition,
		Insert:             newDefaultInsert(),
	}

	for _, spec := range raw.ColumnSpec {
		var err error
		var col = ColumnSpec{Name: spec.Name}

		if spec.Size != "" {
			if col.Size, err = random.ParseDistribution(spec.Size); err != nil {
				return nil, err
			}
		} else {
			col.Size = DefaultColumnSpec.Size
		}
		if spec.Population != "" {
			if col.Population, err = random.ParseDistribution(spec.Population); err != nil {
				return nil, err
			}
		} else {
			col.Population = DefaultColumnSpec.Population
		}
		if spec.Cluster != "" {
			if col.Cluster, err = random.ParseDistribution(spec.Cluster); err != nil {
				return nil, err
			}
		}

		q.ColumnSpec = append(q.ColumnSpec, col)
	}

	if i := raw.Insert; i != nil {
		var err error
		q.Insert.BatchType = i.BatchType

		if i.Partitions != "" {
			if q.Insert.Partitions, err = random.ParseDistribution(i.Partitions); err != nil {
				return nil, err
			}
		}
		if i.Visits != "" {
			if q.Insert.Visits, err = random.ParseDistribution(i.Visits); err != nil {
				return nil, err
			}
		}
		if i.Select != "" {
			if q.Insert.Select, err = random.ParseRatio(i.Select); err != nil {
				return nil, err
			}
		}
	}

	return q, nil
}
