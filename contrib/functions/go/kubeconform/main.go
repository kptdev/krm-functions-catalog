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
	"github.com/kptdev/krm-functions-catalog/contrib/functions/go/kubeconform/generated"
	"github.com/kptdev/krm-functions-sdk/go/fn"
	"github.com/spf13/cobra"
	"os"
)

// nolint
func main() {

	cmd := &cobra.Command{
		Short: generated.KubeconformShort,
		Long:  generated.KubeconformLong,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fn.AsMain(fn.ResourceListProcessorFunc(Run))
		},
		SilenceUsage:  true, // don't print usage on error
		SilenceErrors: true, // suppress default error printing to stderr
	}

	// Optionally add flags here if you want CLI users to override FunctionConfig
	//cmd.Flags().BoolP("help", "h", false, "Show help")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
