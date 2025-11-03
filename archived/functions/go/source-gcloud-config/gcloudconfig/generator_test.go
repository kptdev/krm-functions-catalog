// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package gcloudconfig

import (
	"testing"

	"github.com/kptdev/krm-functions-catalog/archived/functions/go/source-gcloud-generator/exec"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	yaml2 "sigs.k8s.io/yaml"
)

func TestGenerate(t *testing.T) {
	expected := `apiVersion: v1
data:
  domain: test-domain
  namespace: test-namespace
  orgID: test-orgID
  projectID: test-projectID
  region: test-region
  zone: test-zone
kind: ConfigMap
metadata:
  annotations:
    config.kubernetes.io/local-config: "true"
  name: gcloud-config.kpt.dev
`
	exec.GetGcloudContextFn = func() (map[string]string, error) {
		return map[string]string{
			"domain":    "test-domain",
			"namespace": "test-namespace",
			"orgID":     "test-orgID",
			"projectID": "test-projectID",
			"region":    "test-region",
			"zone":      "test-zone",
		}, nil
	}
	gen := GcloudConfigGenerator{}

	outputs, err := gen.Generate([]*yaml.RNode{})
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if len(outputs) != 1 {
		t.Fatalf("expect to generate 1 rnode, got %v", len(outputs))
	}
	a, _ := outputs[0].MarshalJSON()
	actual, _ := yaml2.JSONToYAML(a)
	if string(actual) != expected {
		t.Fatalf("expect %v, got %v", expected, string(actual))
	}

}
