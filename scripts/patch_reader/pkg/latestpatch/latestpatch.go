// Package latestpatch Copyright 2022 Google LLC
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
package latestpatch

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"golang.org/x/mod/semver"
)

var (
	// pattern of release tags, e.g. functions/go/apply-setters/v1.0.1
	releaseTagPattern = regexp.MustCompile(`.*(go|ts)/[-\w]*/(v\d*\.\d*\.\d*)`)
)

func runCmd(name string, arg ...string) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(name, arg...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return stdout.String(), fmt.Errorf("%s\n%s", stderr.String(), err)
	}
	return stdout.String(), err
}

type PatchVersion struct {
	LatestPatch string `json:"latest_patch"`
	Lang        string `json:"lang"`
}

func GetLatestPatch(functionName string, minorVersion string) (PatchVersion, error) {
	if err := gitFetch(); err != nil {
		return PatchVersion{}, err
	}
	tags, err := gitTag()
	if err != nil {
		return PatchVersion{}, err
	}
	funcPattern := fmt.Sprintf("%s/%s", functionName, minorVersion)
	var lang, latestPatchVersion string
	for _, tag := range strings.Split(tags, "\n") {
		if !strings.Contains(tag, funcPattern) || !releaseTagPattern.MatchString(tag) {
			continue
		}
		segments := strings.Split(tag, "/")
		patchVersion := segments[len(segments)-1]
		if latestPatchVersion == "" ||
				semver.Compare(patchVersion, latestPatchVersion) == 1 {
			latestPatchVersion = patchVersion
			lang = segments[len(segments)-3]
		}
	}
	if latestPatchVersion == "" || lang == "" {
		return PatchVersion{}, fmt.Errorf("could not find matching tag for release branch")
	}
	return PatchVersion{
		LatestPatch: latestPatchVersion,
		Lang:        lang,
	}, nil
}

func gitFetch() error {
	_, err := runCmd("git", "fetch", "--tags")
	return err
}

func gitTag() (string, error) {
	return runCmd("git", "tag")
}
