// Copyright 2021 Google LLC
// Modifications Copyright (C) 2025 OpenInfra Foundation Europe.
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
package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

func runCmd(name string, arg ...string) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(name, arg...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	fmt.Printf("%s\n", cmd.String())
	err := cmd.Run()
	if err != nil {
		return stdout.String(), fmt.Errorf("%s\n%s", stderr.String(), err)
	}
	return stdout.String(), err
}

func isCleanRepo() bool {
	_, err := runCmd("git", "diff-index", "--quiet", "HEAD", "--")
	if err != nil {
		return false
	}
	return true
}

func gitFetch() error {
	_, err := runCmd("git", "fetch", "--tags")
	return err
}

func gitCheckout(branch string) error {
	_, err := runCmd("git", "checkout", branch)
	return err
}

func gitAdd() error {
	_, err := runCmd("git", "add", "-u")
	return err
}

func gitCommit(msg string) error {
	stdout, err := runCmd("git", "commit", "-m", msg)
	fmt.Printf("%v\n", stdout)
	return err
}

func gitShow() error {
	stdout, err := runCmd("git", "show")
	fmt.Printf("%v\n", stdout)
	return err
}
