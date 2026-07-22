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

package main

import (
	"fmt"
	"os"

	"github.com/kptdev/krm-functions-catalog/functions/go/ensure-name-substring/generated"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	ensp := EnsureNameSubstringProcessor{}
	cmd := command.Build(&ensp, command.StandaloneEnabled, false)

	cmd.Short = generated.EnsureNameSubstringShort
	cmd.Long = generated.EnsureNameSubstringLong
	return cmd.Execute()
}

type EnsureNameSubstringProcessor struct{}

func (ensp *EnsureNameSubstringProcessor) Process(*framework.ResourceList) error {
	return fmt.Errorf("this version of ensure-name-substring is broken and has been retracted," +
		" please use <= v0.2.0 or >= v0.2.5")
}
