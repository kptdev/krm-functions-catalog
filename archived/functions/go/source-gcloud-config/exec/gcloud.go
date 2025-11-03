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
package exec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"

	"gopkg.in/ini.v1"
)

var domainRegex = regexp.MustCompile(`\S+@(\S+)`)

// For testing
var GetGcloudContextFn = GetGcloudContext

func NewGcloudErr(ServerMsg string) *GcloudErr {
	return &GcloudErr{ServerMsg}
}

type GcloudErr struct {
	Msg string
}

func (g *GcloudErr) Error() string {
	return g.Msg
}

// GetGcloudContext executes `gcloud` commands to get default gcloud config value.
func GetGcloudContext() (map[string]string, error) {
	var projectID, zone, region, account, domain, orgID string
	gcloudCfg, err := ListGcloudConfig()
	if err != nil || gcloudCfg == "" {
		return nil, err
	}

	// Read the INI formatted gcloud config.
	// DO NOT USE github.com/go-gcfg/gcfg which has a nicer function UI but does not support INI with underscore.
	// gcloud config `disable_usage_reporting` breaks.
	iniCfg, err := ini.LoadSources(ini.LoadOptions{Loose: true}, []byte(gcloudCfg))
	if err != nil {
		return nil, err
	}
	core := iniCfg.Section("core")
	if core != nil {
		projectID = core.Key("project").String()
		account = core.Key("account").String()
	}
	compute := iniCfg.Section("compute")
	if compute != nil {
		zone = compute.Key("zone").String()
		region = compute.Key("region").String()
	}

	// Query cloud server to get the OrganizationID which the project ID belongs to.
	var gcloudErr error
	if projectID != "" {
		orgID, gcloudErr = getGcloudOrgID(projectID)
	}
	if account != "" {
		// e.g. account `NAME@COMPANY.com` has matching domain `COMPANY.com`
		matches := domainRegex.FindStringSubmatch(account)
		if len(matches) < 2 {
			// Skip if cannot parse `domain` from gcloud `core/account`
		} else {
			domain = matches[1]
		}
	}
	data := map[string]string{
		"namespace": projectID,
		"projectID": projectID,
		"zone":      zone,
		"region":    region,
		"domain":    domain,
		"orgID":     orgID,
	}
	if gcloudErr != nil {
		// A common error in getting gcloud config is forgetting to run `gcloud auth login`. We surface the error and its
		// instructions to users. This won't cause function failure so the function can concatenate with other functions
		// and run as a pipeline.
		// Users can rerun this function to update the OrgID once they get the authentication.
		return data, gcloudErr
	}
	return data, nil
}

// ListGcloudConfig runs `gcloud config list` to read local default gcloud configuration.
func ListGcloudConfig() (string, error) {
	var cmdOut, cmdErr bytes.Buffer
	cmd := exec.Command("gcloud", "config", "list")
	cmd.Stdout = &cmdOut
	cmd.Stderr = &cmdErr
	err := cmd.Run()
	if err != nil {
		return "", NewGcloudErr(fmt.Sprintf("unable to run `gcloud config list` %v", cmdErr.String()))
	}
	return cmdOut.String(), nil
}

type Ancestor struct {
	ID           string `json:"id,omitempty" yaml:"id,omitempty"`
	ResourceType string `json:"type,omitempty" yaml:"type,omitempty"`
}

// getGcloudOrgID queries cloud server to get the `Organization ID`.
func getGcloudOrgID(projectID string) (string, error) {
	var err, out bytes.Buffer
	cmdListAncestors := exec.Command("gcloud", "projects", "get-ancestors",
		projectID, "--format=json")
	cmdListAncestors.Stdout = &out
	cmdListAncestors.Stderr = &err
	if e := cmdListAncestors.Run(); e != nil {
		return "", NewGcloudErr(fmt.Sprintf("`orgID` is not set: %v", err.String()))
	}
	if err.Len() > 0 {
		return "", NewGcloudErr(fmt.Sprintf("`orgID` is not set: %v", err.String()))
	}

	var rh []Ancestor
	if err := json.Unmarshal(out.Bytes(), &rh); err != nil {
		return "", err
	}
	// A Google Resource Hierarchy can only have one Org and may have multiple folders.
	for _, a := range rh {
		if a.ResourceType == "organization" {
			return a.ID, nil
		}
	}
	return "", nil
}
