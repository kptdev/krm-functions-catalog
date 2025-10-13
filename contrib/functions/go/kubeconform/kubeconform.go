// Copyright (C) 2025 OpenInfra Foundation Europe
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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/kptdev/krm-functions-sdk/go/fn"
)

const (
	SchemaLocationStr = "-schema-location"
	// FunctionConfig keys
	SchemaLocationKey            = "schema_location"
	AdditionalSchemaLocationsKey = "additional_schema_locations"
	IgnoreMissingSchemasKey      = "ignore_missing_schemas"
	SkipKindsKey                 = "skip_kinds"
	StrictKey                    = "strict"
)

var DefaultSchemaLocation = "/jsonschema"

type KubeconformConfig struct {
	SchemaLocation            string
	AdditionalSchemaLocations []string
	IgnoreMissingSchemas      bool
	SkipKinds                 []string
	Strict                    bool
}

// Kubeconform represents the top-level structure returned by kubeconform.
type Kubeconform struct {
	Resources []Resource `json:"resources"`
}

// Resource represents the validation result for a single resource.
type Resource struct {
	Filename         string            `json:"filename"`
	Kind             string            `json:"kind"`
	Name             string            `json:"name"`
	Version          string            `json:"version"`
	Status           string            `json:"status"` // e.g. "statusValid" or "statusInvalid"
	Message          string            `json:"msg"`
	ValidationErrors []ValidationError `json:"validationErrors"`
}

// ValidationError provides detailed schema validation issues.
type ValidationError struct {
	Path string `json:"path"`
	Msg  string `json:"msg"`
}

func Run(rl *fn.ResourceList) (bool, error) {
	cfg := extractConfig(rl.FunctionConfig)
	args := buildKubeconformArgs(
		cfg.SchemaLocation,
		cfg.AdditionalSchemaLocations,
		cfg.IgnoreMissingSchemas,
		cfg.SkipKinds,
		cfg.Strict,
	)

	var results fn.Results
	var exit error
	for _, obj := range rl.Items {
		objResults, err := runKubeconformForObject(obj, args)
		if err != nil {
			if exit == nil {
				exit = fmt.Errorf("KRM validation failed")
			}
			results = append(results, objResults...)
			continue
		}
	}

	rl.Results = append(rl.Results, results...)
	if exit != nil {
		return false, nil
	}
	return true, nil
}

func runKubeconformForObject(obj *fn.KubeObject, args []string) (fn.Results, error) {
	cmd := exec.Command("kubeconform", args...)

	data, err := marshalKubeObject(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal object: %w", err)
	}
	cmd.Stdin = bytes.NewReader(data)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		if stdout.Len() == 0 {
			return nil, fmt.Errorf("failed to run kubeconform: %w", err)
		}
	}

	raw := stdout.Bytes()
	if bytes.Contains(raw, []byte("Failed initializing schema file")) {
		return fn.Results{
			fn.ErrorConfigObjectResult(
				fmt.Errorf("%s", fmt.Sprintf("Validating arbitrary CRDs not supported. Consider setting %s or %s in FunctionConfig:\n%s",
					IgnoreMissingSchemasKey, SkipKindsKey, string(raw))),
				obj,
			),
		}, nil
	}

	var kubeConf Kubeconform
	if err := json.Unmarshal(raw, &kubeConf); err != nil {
		return fn.Results{
			fn.ErrorConfigObjectResult(
				fmt.Errorf("%s", fmt.Sprintf("Failed to parse kubeconform output:\n%s\n%s", err, string(raw))),
				obj,
			),
		}, nil
	}

	var results fn.Results
	var hasError bool
	for _, res := range kubeConf.Resources {
		if res.Status != "statusValid" {
			hasError = true
			for _, e := range res.ValidationErrors {
				results = append(results, ConfigObjectResult(e.Msg, e.Path, obj, fn.Error))
			}
		}
	}
	var validationError error
	if hasError {
		validationError = fmt.Errorf("validation failed for one or more resources")
	}

	return results, validationError
}

func ConfigObjectResult(msg string, path string, obj *fn.KubeObject, severity fn.Severity) *fn.Result {
	return &fn.Result{
		Message:  msg,
		Severity: severity,
		ResourceRef: &fn.ResourceRef{
			APIVersion: obj.GetAPIVersion(),
			Kind:       obj.GetKind(),
			Name:       obj.GetName(),
			Namespace:  obj.GetNamespace(),
		},
		Field: &fn.Field{
			Path: strings.TrimPrefix(strings.ReplaceAll(path, "/", "."), "."),
		},
		File: &fn.File{
			Path:  obj.PathAnnotation(),
			Index: obj.IndexAnnotation(),
		},
	}
}

func extractConfig(fc *fn.KubeObject) KubeconformConfig {
	getString := func(key string) string {
		val := fc.GetString(key)
		if val == "" {
			val, _, _ = fc.NestedString("data", key)
		}
		return val
	}

	getStringList := func(key string) []string {
		val := getString(key)
		if val == "" {
			return nil
		}
		parts := strings.Split(val, ",")
		if len(parts) == 1 && parts[0] == "" {
			return nil
		}
		return parts
	}

	getBool := func(key string) bool {
		if val, found, _ := fc.NestedBool(key); found {
			return val
		}
		strVal := getString(key)
		return strings.EqualFold(strVal, "true") || strVal == "1"
	}

	return KubeconformConfig{
		SchemaLocation:            getString(SchemaLocationKey),
		AdditionalSchemaLocations: getStringList(AdditionalSchemaLocationsKey),
		IgnoreMissingSchemas:      getBool(IgnoreMissingSchemasKey),
		SkipKinds:                 getStringList(SkipKindsKey),
		Strict:                    getBool(StrictKey),
	}
}

func buildKubeconformArgs(schemaLocation string, additional []string, ignoreMissing bool, skipKinds []string, strict bool) []string {
	args := []string{"-output", "json"}

	// Use specified schema locations or fallback to /jsonschema/* directories
	if schemaLocation != "" {
		args = append(args, SchemaLocationStr, schemaLocation)
	}

	if len(additional) > 0 {
		for _, add := range additional {
			args = append(args, SchemaLocationStr, add)
		}
	}

	if schemaLocation == "" && len(additional) == 0 {
		args = append(args, SchemaLocationStr, "file://"+DefaultSchemaLocation)
	}

	if ignoreMissing {
		args = append(args, "-ignore-missing-schemas")
	}
	if len(skipKinds) > 0 {
		args = append(args, "-skip", strings.Join(skipKinds, ","))
	}
	if strict {
		args = append(args, "-strict")
	}
	return args
}

func marshalKubeObject(obj *fn.KubeObject) ([]byte, error) {
	// Convert KubeObject to a typed object (e.g., map[string]interface{})
	var typed map[string]interface{}
	if err := obj.As(&typed); err != nil {
		return nil, fmt.Errorf("failed to convert KubeObject to typed map: %w", err)
	}
	// Marshal the typed map to JSON
	data, err := json.Marshal(typed)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal typed object to JSON: %w", err)
	}
	return data, nil
}
