// Copyright 2021, 2026 The kpt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"

	"github.com/kptdev/kpt/api/kubeobject"
	"github.com/kptdev/krm-functions-sdk/go/fn"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8syaml "sigs.k8s.io/yaml"
)

const (
	stdin  = "/dev/stdin"
	stdout = "/dev/stdout"
)

//go:embed README.md
var readme []byte

//go:embed metadata.yaml
var metadata []byte

// GatekeeperProcessor validates the resources with the Gatekeeper policies
// found in the same ResourceList. It also holds the values of the command line
// flags for reading the input from and writing the output to files and for
// using JSON instead of YAML.
type GatekeeperProcessor struct {
	input   string
	output  string
	useJSON bool
}

func (gkp *GatekeeperProcessor) Process(resourceList *kubeobject.ResourceList) (bool, error) {
	var objects []*unstructured.Unstructured
	for _, item := range resourceList.Items {
		un := &unstructured.Unstructured{}
		if err := k8syaml.Unmarshal([]byte(item.String()), un); err != nil {
			return false, err
		}

		objects = append(objects, un)
	}

	results, err := Validate(objects)
	// When err is not nil, results should be nil.
	if err != nil {
		results = kubeobject.Results{
			{
				Message:  err.Error(),
				Severity: kubeobject.Error,
			},
		}
	}
	resourceList.Results = results
	if results.ExitCode() != 0 {
		return false, results
	}
	return true, nil
}

// addFlags registers the command line flags of the function.
func (gkp *GatekeeperProcessor) addFlags(fs *flag.FlagSet) {
	const (
		inputUsage  = "path to the input file"
		outputUsage = "path to the output file"
	)
	fs.StringVar(&gkp.input, "input", stdin, inputUsage)
	fs.StringVar(&gkp.input, "i", stdin, inputUsage)
	fs.StringVar(&gkp.output, "output", stdout, outputUsage)
	fs.StringVar(&gkp.output, "o", stdout, outputUsage)
	fs.BoolVar(&gkp.useJSON, "json", false, "input and output is JSON instead of YAML")
}

// readInput reads the function input from the file given with --input.
func (gkp *GatekeeperProcessor) readInput() ([]byte, error) {
	content, err := os.ReadFile(gkp.input)
	if err != nil {
		return nil, fmt.Errorf("unable to process input: %w", err)
	}
	if gkp.useJSON {
		content, err = k8syaml.JSONToYAML(content)
		if err != nil {
			return nil, fmt.Errorf("unable to process input: %w", err)
		}
	}

	if len(content) > 0 && content[0] == '{' {
		// yaml.Unmarshal doesn't fail on trying to parse JSON, and will happily
		// return something. This safeguards against that.
		return nil, fmt.Errorf("tried to parse JSON as YAML. Use --json flag")
	}
	return content, nil
}

// writeOutput writes the function output to the file given with --output.
func (gkp *GatekeeperProcessor) writeOutput(content []byte) error {
	var err error
	if gkp.useJSON {
		content, err = k8syaml.YAMLToJSON(content)
		if err != nil {
			return fmt.Errorf("unable to process output: %w", err)
		}
	}

	if err := os.WriteFile(gkp.output, content, 0644); err != nil {
		return fmt.Errorf("unable to process output: %w", err)
	}
	return nil
}

func main() {
	gkp := &GatekeeperProcessor{}
	err := fn.AsMain(gkp,
		fn.WithDocs(readme, metadata),
		fn.WithFlags(gkp.addFlags),
		fn.WithInput(gkp.readInput),
		fn.WithOutput(gkp.writeOutput),
	)
	if err != nil {
		os.Exit(1)
	}
}
