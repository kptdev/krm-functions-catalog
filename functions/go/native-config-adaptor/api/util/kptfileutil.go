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

package util

import (
	"fmt"
	"strings"

	kptfilev1 "github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/native-config-adaptor/api/kptfile/v1"
	"gopkg.in/yaml.v3"
)

// DecodeKptfile decodes a KptFile from a yaml string.
func DecodeKptfile(kf string) (*kptfilev1.KptFile, error) {
	kptfile := &kptfilev1.KptFile{}
	f := strings.NewReader(kf)
	d := yaml.NewDecoder(f)
	d.KnownFields(true)
	if err := d.Decode(&kptfile); err != nil {
		return &kptfilev1.KptFile{}, fmt.Errorf("invalid 'v1' Kptfile: %w", err)
	}
	return kptfile, nil
}
