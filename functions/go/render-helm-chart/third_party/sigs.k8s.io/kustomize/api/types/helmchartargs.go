// Copyright 2026 The kpt Authors
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
// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types

import "sigs.k8s.io/kustomize/kyaml/yaml"

type HelmGlobals struct {
	// ChartHome is a file path to a directory containing a subdirectory for
	// each chart to be included in the output. The default value of this field
	// is "tmp/charts".
	// At runtime, the function will look for the chart under {ChartHome}. If it
	// is there, the function will use it as found. If it is not there, the
	// function will attempt to pull it and put it in {ChartHome}.
	// When run as a container function, local directories must be mounted into
	// the container in order for the function to use them.
	// If the function needs to pull the helm chart while running in a container,
	// ChartHome MUST start with "tmp/".
	ChartHome string `json:"chartHome,omitempty" yaml:"chartHome,omitempty"`

	// ConfigHome defines a value that the function should pass to helm via
	// the HELM_CONFIG_HOME environment variable. If this is set, the function
	// also sets
	//   HELM_CACHE_HOME={ConfigHome}/.cache
	//   HELM_DATA_HOME={ConfigHome}/.data
	// for the helm subprocess.
	ConfigHome string `json:"configHome,omitempty" yaml:"configHome,omitempty"`
}

type HelmChart struct {
	// ChartArgs encapsulates information about the chart being inflated, including
	// the chart's name, version, and repo.
	ChartArgs `json:"chartArgs,omitempty" yaml:"chartArgs,omitempty"`

	// TemplateOptions are fields that become flags to `helm template` when
	// the helm chart is being rendered.
	TemplateOptions `json:"templateOptions,omitempty" yaml:"templateOptions,omitempty"`
}

type ChartArgs struct {
	// Name is the name of the chart, e.g. 'minecraft'.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Version is the version of the chart, e.g. '3.1.3'
	Version string `json:"version,omitempty" yaml:"version,omitempty"`

	// Repo is a URL locating the chart on the internet.
	// This is the argument to helm's  `--repo` flag, e.g.
	// `https://itzg.github.io/minecraft-server-charts`.
	Repo string `json:"repo,omitempty" yaml:"repo,omitempty"`

	// Auth is a reference to the kubernetes resource that
	// contains credentials necessary to access the repository if
	// it is private
	Auth *yaml.ResourceIdentifier `json:"auth,omitempty" yaml:"auth,omitempty"`

	// Registry is the name of the chart registry (only required if
	// the chart comes from an OCI repository)
	Registry string `json:"registry,omitempty" yaml:"registry,omitempty"`
}

type TemplateOptions struct {
	// ApiVersions is the kubernetes apiversions used for Capabilities.APIVersions
	ApiVersions []string `json:"apiVerions,omitempty" yaml:"apiVersions,omitempty"`

	// ReleaseName replaces RELEASE-NAME in chart template output,
	// making a particular inflation of a chart unique with respect to
	// other inflations of the same chart in a cluster. It's the first
	// argument to the helm `install` and `template` commands, i.e.
	//   helm install {RELEASE-NAME} {chartName}
	//   helm template {RELEASE-NAME} {chartName}
	// If omitted, the flag --generate-name is passed to 'helm template'.
	ReleaseName string `json:"releaseName,omitempty" yaml:"releaseName,omitempty"`

	// Namespace set the target namespace for a release. It is .Release.Namespace
	// in the helm template
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`

	// Description is a custom description to add when rendering the helm chart.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// NameTemplate is for specifying the name template used to name the release.
	NameTemplate string `json:"nameTemplate,omitempty" yaml:"nameTemplate,omitempty"`

	// IncludeCRDs specifies if Helm should also generate CustomResourceDefinitions.
	// Defaults to false.
	IncludeCRDs bool `json:"includeCRDs,omitempty" yaml:"includeCRDs,omitempty"`

	// SkipTests skips tests from templated output.
	SkipTests bool `json:"skipTests,omitempty" yaml:"skipTests,omitempty"`

	// Values are values that are specified inline or in a yaml file to use.
	Values `json:"values,omitempty" yaml:"values,omitempty"`
}

type Values struct {
	// ValuesFiles is a list of local file paths to values files to use instead of
	// the default values that accompanied the chart.
	// The default values are in '{ChartHome}/{Name}/values.yaml'.
	ValuesFiles []string `json:"valuesFiles,omitempty" yaml:"valuesFiles,omitempty"`

	// ValuesInline holds value mappings specified directly,
	// rather than in a separate file.
	ValuesInline map[string]interface{} `json:"valuesInline,omitempty" yaml:"valuesInline,omitempty"`

	// ValuesMerge specifies how to treat ValuesInline with respect to Values.
	// Legal values: 'merge', 'override', 'replace'.
	// Defaults to 'override'.
	ValuesMerge string `json:"valuesMerge,omitempty" yaml:"valuesMerge,omitempty"`
}
