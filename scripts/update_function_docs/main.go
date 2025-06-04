// Copyright 2021 Google LLC
# Modifications Copyright (C) 2025 OpenInfra Foundation Europe.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Usage: update_function_docs -branch <RELEASE_BRANCH>
//
// e.g. update_function_docs -branch origin/apply-setters/v0.2
//
// The command will checkout the release branch and update the function/example
// docs with the latest patch version for the release. If the docs are updated
// then a commit is created with the changes. The manual steps left to the user
// are to push the commit to a branch and create a pull request.
package main

import (
	"flag"
	"fmt"
	"os"
)

func exitWithErr(err error) {
	fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(1)
}

type arguments struct {
	ReleaseBranch string
}

// validate command line arguments
func (a arguments) validate() error {
	if a.ReleaseBranch == "" {
		return fmt.Errorf("release branch not set")
	}
	return nil
}

// parse command line arguments
func parseArgs() (arguments, error) {
	args := arguments{}
	flag.StringVar(&args.ReleaseBranch, "branch", os.Getenv("RELEASE_BRANCH"),
		"release branch (can also use RELEASE_BRANCH environment variable)")

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
	if !isCleanRepo() {
		exitWithErr(fmt.Errorf("dirty repo"))
	}
	if err = gitFetch(); err != nil {
		exitWithErr(err)
	}
	if err = gitCheckout(args.ReleaseBranch); err != nil {
		exitWithErr(err)
	}
	fr, err := newFunctionRelease(args.ReleaseBranch)
	if err != nil {
		exitWithErr(err)
	}
	if err = fr.updateDocs(); err != nil {
		exitWithErr(err)
	}
	if isCleanRepo() {
		exitWithErr(fmt.Errorf("docs up to date"))
	}
	if err = gitAdd(); err != nil {
		exitWithErr(err)
	}
	msg := fmt.Sprintf("docs: Update tags for %s/%s/%s",
		fr.Language, fr.FunctionName, fr.LatestPatchVersion)
	if err = gitCommit(msg); err != nil {
		exitWithErr(err)
	}
	if err = gitShow(); err != nil {
		exitWithErr(err)
	}
}
