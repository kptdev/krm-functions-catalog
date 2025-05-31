// Copyright 2021 Google LLC
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
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/gatekeeper/generated"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	k8syaml "sigs.k8s.io/yaml"
)

const (
	stdin    = "/dev/stdin"
	stdout   = "/dev/stdout"
	nullByte = "\x00"
)

type GatekeeperProcessor struct {
	input   string
	output  string
	useJSON bool

	inputBuf  *bytes.Buffer
	outputBuf *bytes.Buffer
}

func (gkp *GatekeeperProcessor) Process(resourceList *framework.ResourceList) error {
	resourceList.Result = &framework.Result{
		Name: "gatekeeper",
	}
	var objects []*unstructured.Unstructured
	for _, item := range resourceList.Items {
		s, err := item.String()
		if err != nil {
			return err
		}

		un := &unstructured.Unstructured{}
		err = k8syaml.Unmarshal([]byte(s), un)
		if err != nil {
			return err
		}

		// add the filepath to the objects name in sanitized form. this is done
		// to get unique identifier in case the same resource is defined in
		// different files. Usually this happens when running the function
		// across packages
		if !isTemplate(un) && !isConstraint(un) {
			un.SetName(fmt.Sprintf("%s%s%s", un.GetName(), nullByte,
				strings.ReplaceAll(item.GetAnnotations()[kioutil.PathAnnotation], "/", nullByte)))
		}

		objects = append(objects, un)
	}

	result, err := Validate(objects)
	// When err is not nil, result should be nil.
	if err != nil {
		result = &framework.Result{
			Items: []framework.ResultItem{
				{
					Message:  err.Error(),
					Severity: framework.Error,
				},
			},
		}
	}

	// unwrap the null-byte filename hack again
	if result != nil {
		for i, item := range result.Items {
			parts := strings.SplitN(item.ResourceRef.Name, nullByte, 2)
			item.ResourceRef.Name = parts[0]
			if len(parts) == 2 {
				item.Field.Path = strings.ReplaceAll(parts[1], nullByte, "/")
			}
			result.Items[i] = item
		}
	}

	resourceList.Result = result
	if resultContainsError(result) {
		return result
	}
	return nil
}

func (gkp *GatekeeperProcessor) ProcessInput() error {
	content, err := os.ReadFile(gkp.input)
	if err != nil {
		return fmt.Errorf("unable to process input: %w", err)
	}
	if gkp.useJSON {
		content, err = k8syaml.JSONToYAML(content)
		if err != nil {
			return fmt.Errorf("unable to process input: %w", err)
		}
	}

	if len(content) > 0 && content[0] == '{' {
		// yaml.Unmarshal doesn't fail on trying to parse JSON, and will happily
		// return something. This safeguards against that.
		return fmt.Errorf("tried to parse JSON as YAML. Use --json flag.")
	}

	gkp.inputBuf = bytes.NewBuffer(content)
	return nil
}

func (gkp *GatekeeperProcessor) ProcessOutput() error {
	var err error
	if gkp.outputBuf == nil {
		return fmt.Errorf("the output buffer must not be nil")
	}
	content := gkp.outputBuf.Bytes()
	if gkp.useJSON {
		content, err = k8syaml.YAMLToJSON(content)
		if err != nil {
			return fmt.Errorf("unable to process output: %w", err)
		}
	}

	err = os.WriteFile(gkp.output, content, 0644)
	if err != nil {
		return fmt.Errorf("unable to process output: %w", err)
	}
	return nil
}

func (gkp *GatekeeperProcessor) Read(p []byte) (n int, err error) {
	if gkp.inputBuf == nil {
		return 0, nil
	}
	return gkp.inputBuf.Read(p)
}

func (gkp *GatekeeperProcessor) Write(p []byte) (n int, err error) {
	if gkp.outputBuf == nil {
		gkp.outputBuf = &bytes.Buffer{}
	}
	return gkp.outputBuf.Write(p)
}

func (gkp *GatekeeperProcessor) addFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&gkp.input, "input", "i", stdin,
		`path to the input file`)
	cmd.Flags().StringVarP(&gkp.output, "output", "o", stdout,
		`path to the output file`)
	cmd.Flags().BoolVar(&gkp.useJSON, "json", false,
		`input and output is JSON instead of YAML`)
}

func (gkp *GatekeeperProcessor) runGatekeeper() error {
	if err := gkp.ProcessInput(); err != nil {
		return err
	}
	// Process the error only after processing the output.
	err := framework.Execute(gkp, &kio.ByteReadWriter{
		Reader: gkp,
		Writer: gkp,
		// We should not set the id annotation in the function, since we should not
		// overwrite what the orchestrator set.
		OmitReaderAnnotations: true,
		// We should not remove the id annotations in the function, since the
		// orchestrator (e.g. kpt) may need them.
		KeepReaderAnnotations: true,
	})
	err2 := gkp.ProcessOutput()
	if err != nil {
		return err
	}
	return err2
}

func main() {
	gkp := &GatekeeperProcessor{}
	cmd := &cobra.Command{
		Short: generated.GatekeeperShort,
		Long:  generated.GatekeeperLong,
		Run: func(cmd *cobra.Command, args []string) {
			if err := gkp.runGatekeeper(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}
	gkp.addFlags(cmd)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func resultContainsError(result *framework.Result) bool {
	if result == nil {
		return false
	}
	for _, item := range result.Items {
		if item.Severity == framework.Error {
			return true
		}
	}
	return false
}

func isTemplate(u *unstructured.Unstructured) bool {
	gvk := u.GroupVersionKind()
	return gvk.Kind == "ConstraintTemplate"
}

func isConstraint(u *unstructured.Unstructured) bool {
	gvk := u.GroupVersionKind()
	return gvk.Group == "constraints.gatekeeper.sh"
}
