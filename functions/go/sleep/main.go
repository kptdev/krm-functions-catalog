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

	"github.com/kptdev/krm-functions-catalog/functions/go/sleep/generated"
	"github.com/kptdev/krm-functions-catalog/functions/go/sleep/processor"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
)

func main() {
	cmd := command.Build(&processor.SleepProcessor{}, command.StandaloneEnabled, false)
	cmd.Short = generated.SleepShort
	cmd.Long = generated.SleepLong
	cmd.Example = generated.SleepExamples

	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
}
