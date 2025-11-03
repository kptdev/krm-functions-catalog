// Copyright 2022 Google LLC
// Modifications Copyright (C) 2025 OpenInfra Foundation Europe.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package gcloudconfig

import (
	"errors"
	"fmt"

	"github.com/kptdev/krm-functions-catalog/archived/functions/go/source-gcloud-generator/exec"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
)

const ResultFile = "gcloud-config.yaml"

func NewProcessor() *Processor {
	return &Processor{}
}

type Processor struct{}

func (p *Processor) Process(resourceList *framework.ResourceList) error {
	gen := &GcloudConfigGenerator{}
	updated, err := gen.Generate(resourceList.Items)
	var gcloudErr *exec.GcloudErr
	switch {
	case err == nil:
		resourceList.Items = updated
	case errors.As(err, &gcloudErr):
		resourceList.Items = updated
		resourceList.Results = append(resourceList.Results,
			&framework.Result{
				Message:  err.Error(),
				Severity: framework.Warning,
			})
	default:
		resourceList.Results = framework.Results{
			&framework.Result{
				Message:  err.Error(),
				Severity: framework.Error,
			},
		}
		return resourceList.Results
	}
	// Store gcloud context to `gcloud-config.yaml`. Otherwise it will use default file pattern
	// configmap_gcloud-config.kpt.dev.yaml
	if err := resourceList.Filter(&filters.FileSetter{FilenamePattern: ResultFile}); err != nil {
		resourceList.Results = framework.Results{
			&framework.Result{
				Message:  err.Error(),
				Severity: framework.Error,
			},
		}
		return resourceList.Results
	}

	// Notify users the gcloud context is stored in `gcloud-config.yaml`.
	resourceList.Results = append(resourceList.Results, &framework.Result{
		Message:  fmt.Sprintf("store gcloud context in %v", ResultFile),
		Severity: framework.Info,
		File:     &framework.File{Path: ResultFile, Index: 0},
	})
	return nil
}
