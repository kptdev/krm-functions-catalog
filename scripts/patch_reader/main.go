// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kptdev/krm-functions-catalog/scripts/patch_reader/pkg/latestpatch"
)

func exitWithErr(err error) {
	fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(1)
}

type arguments struct {
	FunctionName string
	MinorVersion string
}

// validate command line arguments
func (a arguments) validate() error {
	if a.FunctionName == "" {
		return fmt.Errorf("function name not set")
	}
	if a.MinorVersion == "" {
		return fmt.Errorf("minor version not set")
	}
	return nil
}

// parse command line arguments
func parseArgs() (arguments, error) {
	args := arguments{}
	flag.StringVar(&args.FunctionName, "function", os.Getenv("FUNCTION_NAME"),
		"function name (can also use FUNCTION_NAME environment variable)")
	flag.StringVar(&args.MinorVersion, "minor", os.Getenv("MINOR_VERSION"),
		"minor version (can also use MINOR_VERSION environment variable)")

	flag.Parse()

	err := args.validate()
	if err != nil {
		flag.Usage()
	}
	return args, err
}

func main() {
	var err error
	args, err := parseArgs()
	if err != nil {
		exitWithErr(err)
	}
	patch, err := latestpatch.GetLatestPatch(args.FunctionName, args.MinorVersion)
	if err != nil {
		exitWithErr(err)
	}
	output, err := json.Marshal(patch)
	if err != nil {
		exitWithErr(err)
	}
	fmt.Printf("%s", output)
}
