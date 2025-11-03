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

package terraformgenerator

import (
	"embed"
	"strings"

	sdk "github.com/kptdev/krm-functions-catalog/archived/functions/go/export-terraform/thirdparty/kyaml/fnsdk"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:embed templates
var templates embed.FS

const (
	kccAPI         = "cnrm.cloud.google.com"
	skipAnnotation = "cnrm.cloud.google.com/ignore-clusterless"
)

func Processor(rl *sdk.ResourceList) error {
	var resources terraformResources
	supportedKinds := map[string]bool{
		"Folder":                      true,
		"Organization":                true,
		"IAMPolicyMember":             true,
		"IAMPartialPolicy":            true,
		"IAMPolicy":                   true,
		"Project":                     true,
		"LoggingLogSink":              true,
		"BigQueryDataset":             true,
		"PubSubTopic":                 true,
		"StorageBucket":               true,
		"IAMAuditConfig":              true,
		"ComputeNetwork":              true,
		"ComputeSubnetwork":           true,
		"ComputeRoute":                true,
		"ComputeRouter":               true,
		"ComputeRouterNAT":            true,
		"ComputeAddress":              true,
		"ServiceNetworkingConnection": true,
		"ComputeFirewall":             true,
		"LoggingLogBucket":            true,
		"ComputeSharedVPCHostProject": true,
	}

	for _, item := range rl.Items {
		if !strings.Contains(item.APIVersion(), kccAPI) {
			continue
		}

		shouldSkip := item.Annotation(skipAnnotation)
		if shouldSkip == "true" {
			continue
		}

		if _, ok := supportedKinds[item.Kind()]; !ok {
			continue
		}

		_, err := resources.getResourceRef(item.Kind(), strings.TrimSpace(item.Name()), item)
		if err != nil {
			return err
		}
	}

	resources.makeVariables()

	data, err := resources.getHCL()
	if err != nil {
		return err
	}

	configMap := makeConfigMap(data)

	return rl.UpsertObjectToItems(configMap, nil, false)
}

func makeConfigMap(data map[string]string) interface{} {
	configMap := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "terraform",
			Annotations: map[string]string{
				"config.kubernetes.io/local-config":  "true",
				"blueprints.cloud.google.com/syntax": "hcl",
				"blueprints.cloud.google.com/flavor": "terraform",
				"internal.config.kubernetes.io/path": "terraform.yaml",
			},
		},
		Data: data,
	}
	return configMap
}
