// Copyright 2017 Elasticsearch Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aucoalesce

import (
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

//go:generate sh -c "go run mknormalize_data.go normalizationData normalizations.yaml > znormalize_data.go"

var (
	syscallNorms    map[string]*Normalization
	recordTypeNorms map[string]*Normalization
)

func init() {
	data, err := asset("normalizationData")
	if err != nil {
		panic("normalizationData not found in assets")
	}

	syscallNorms, recordTypeNorms, err = LoadNormalizationConfig(data)
	if err != nil {
		panic(errors.Wrap(err, "failed to parse built in normalization mappings"))
	}
}

// Strings is a custom type to enable YAML values that can be either a string
// or a list of strings.
type Strings struct {
	Values []string
}

func (s *Strings) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var singleValue string
	if err := unmarshal(&singleValue); err == nil {
		s.Values = append(s.Values, singleValue)
		return nil
	}

	return unmarshal(&s.Values)
}

type NormalizationConfig struct {
	Default        Normalization `yaml:"default"`
	Normalizations []Normalization
}

type Normalization struct {
	Subject     SubjectMapping `yaml:"subject"`
	Action      string         `yaml:"action"`
	Object      ObjectMapping  `yaml:"object"`
	How         Strings        `yaml:"how"`
	RecordTypes []string       `yaml:"record_types"`
	Syscalls    []string       `yaml:"syscalls"`
}

type SubjectMapping struct {
	PrimaryFieldName   Strings `yaml:"primary"`
	SecondaryFieldName Strings `yaml:"secondary"`
}

type ObjectMapping struct {
	PrimaryFieldName   Strings `yaml:"primary"`
	SecondaryFieldName Strings `yaml:"secondary"`
	What               string  `yaml:"what"`
	PathIndex          int     `yaml:"path_index"`
}

type HowMapping struct {
	FieldName string `yaml:"field"`
}

func LoadNormalizationConfig(b []byte) (syscalls map[string]*Normalization, recordTypes map[string]*Normalization, err error) {
	c := &NormalizationConfig{}
	if err := yaml.Unmarshal(b, c); err != nil {
		return nil, nil, err
	}

	syscalls = map[string]*Normalization{}
	recordTypes = map[string]*Normalization{}

	for i := range c.Normalizations {
		norm := c.Normalizations[i]
		for _, syscall := range norm.Syscalls {
			if _, found := syscalls[syscall]; found {
				return nil, nil, fmt.Errorf("duplication mappings for sycall %v", syscall)
			}
			syscalls[syscall] = &norm
		}
		for _, recordType := range norm.RecordTypes {
			if _, found := recordTypes[recordType]; found {
				return nil, nil, fmt.Errorf("duplication mappings for record_type %v", recordType)
			}
			recordTypes[recordType] = &norm
		}
	}

	return syscalls, recordTypes, nil
}
