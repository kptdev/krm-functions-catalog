// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate HelmChartInflationGeneratorPluginator
package builtins

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"github.com/imdario/mergo"
	"github.com/kptdev/krm-functions-catalog/functions/go/render-helm-chart/third_party/sigs.k8s.io/kustomize/api/types"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/yaml"
)

type HelmArgs struct {
	types.HelmGlobals `json:"helmGlobals,omitempty" yaml:"helmGlobals,omitempty"`
	HelmCharts        []types.HelmChart `json:"helmCharts,omitempty" yaml:"helmCharts,omitempty"`
}

type HelmChartInflationGeneratorPlugin struct {
	types.HelmGlobals `json:",inline,omitempty" yaml:",inline,omitempty"`
	types.HelmChart   `json:",inline,omitempty" yaml:",inline,omitempty"`
	username          string
	password          string
	tmpDir            string
}

const (
	valuesMergeOptionMerge    = "merge"
	valuesMergeOptionOverride = "override"
	valuesMergeOptionReplace  = "replace"
)

var legalMergeOptions = []string{
	valuesMergeOptionMerge,
	valuesMergeOptionOverride,
	valuesMergeOptionReplace,
}

// noinspection GoUnusedGlobalVariable
var KustomizeHelmChartInflationGeneratorPlugin HelmChartInflationGeneratorPlugin

func (p *HelmChartInflationGeneratorPlugin) establishTmpDir() (err error) {
	if p.tmpDir != "" {
		// already done.
		return nil
	}

	p.tmpDir, err = os.MkdirTemp("", "kustomize-helm-")
	return err
}

func (p *HelmChartInflationGeneratorPlugin) ConfigureAuth(items []*fn.KubeObject) (err error) {
	if p.Auth == nil {
		return nil
	}
	if p.Auth.GetKind() != "Secret" {
		return fmt.Errorf("auth `kind` must be `Secret`")
	}

	var targetSecret *fn.KubeObject
	for _, i := range items {
		iNamespace := i.GetNamespace()
		if iNamespace == "" {
			iNamespace = "default"
		}
		authNamespace := p.Auth.Namespace
		if authNamespace == "" {
			authNamespace = "default"
		}
		if i.GetKind() == "Secret" && i.GetName() == p.Auth.Name && iNamespace == authNamespace {
			targetSecret = i
		}
	}
	if targetSecret == nil {
		return fmt.Errorf("could not find Secret %q identified by auth", p.Auth)
	}

	var secret corev1.Secret
	if err := targetSecret.As(&secret); err != nil {
		return fmt.Errorf("could not unmarshal Secret: %s", err.Error())
	}

	user, ok := secret.Data["username"]
	if !ok || len(user) == 0 {
		return fmt.Errorf("could not find username in Secret %s", secret.Name)
	}

	pass, ok := secret.Data["password"]
	if !ok || len(pass) == 0 {
		return fmt.Errorf("could not find password in Secret %s", secret.Name)
	}

	p.username = string(user)
	p.password = string(pass)
	return nil
}

func (p *HelmChartInflationGeneratorPlugin) ValidateArgs() (err error) {
	// ChartHome might be written to by the function in a container,
	// so by default it must be under the `/tmp` directory
	if p.ChartHome == "" {
		p.ChartHome = "tmp/charts"
	}

	if len(p.ValuesFiles) == 0 {
		p.ValuesFiles = append(p.ValuesFiles, filepath.Join(p.ChartHome, p.Name, "values.yaml"))
	}

	if err = p.errIfIllegalValuesMerge(); err != nil {
		return err
	}
	return nil
}

func (p *HelmChartInflationGeneratorPlugin) errIfIllegalValuesMerge() error {
	if p.ValuesMerge == "" {
		// Use the default.
		p.ValuesMerge = valuesMergeOptionOverride
		return nil
	}
	for _, opt := range legalMergeOptions {
		if p.ValuesMerge == opt {
			return nil
		}
	}
	return fmt.Errorf("valuesMerge must be one of %v", legalMergeOptions)
}

func (p *HelmChartInflationGeneratorPlugin) absChartHome() string {
	path, _ := filepath.Abs(p.ChartHome)
	return path
}

func (p *HelmChartInflationGeneratorPlugin) runHelmCommand(
	args []string) ([]byte, error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := exec.Command("helm", args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	var env []string
	if p.ConfigHome != "" {
		env = []string{
			fmt.Sprintf("HELM_CONFIG_HOME=%s", p.ConfigHome),
			fmt.Sprintf("HELM_CACHE_HOME=%s/.cache", p.ConfigHome),
			fmt.Sprintf("HELM_DATA_HOME=%s/.data", p.ConfigHome)}
		cmd.Env = append(os.Environ(), env...)
	}
	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("unable to run: '%s %s' with env=%s; %q: %q",
			"helm", args[0], env, err.Error(), stderr.String())
	}
	return stdout.Bytes(), err
}

// createNewMergedValuesFiles replaces/merges original values file with ValuesInline.
func (p *HelmChartInflationGeneratorPlugin) createNewMergedValuesFiles(path string) (
	string, error) {
	pValues, err := os.ReadFile(path)
	if err != nil {
		if u, urlErr := url.Parse(path); urlErr == nil {
			if u.Scheme == "http" || u.Scheme == "https" {
				resp, err := http.Get(path)
				if err != nil {
					return "", err
				}
				if resp.StatusCode != http.StatusOK {
					return "", fmt.Errorf("non-ok http status: %d returned", resp.StatusCode)
				}
				defer resp.Body.Close()
				pValues, err = io.ReadAll(resp.Body)
				if err != nil {
					return "", err
				}
			} else { // url scheme is not http or https
				schemeErr := fmt.Errorf("unsupported URL scheme: %s", path)
				return "", fmt.Errorf(
					"could not read provided values file %q: when reading as file path, received error %v; when reading as URL, received error %v",
					path, err, schemeErr)
			}
		} else { // invalid path and invalid URL
			return "", fmt.Errorf(
				"could not read provided values file %q: when reading as file path, received error %v; when reading as URL, received error %v",
				path, err, urlErr)
		}
	} else {
		// we want to pass in the absolute path into writeValuesBytes
		path, err = filepath.Abs(path)
		if err != nil {
			return "", err
		}
	}
	if p.ValuesMerge == valuesMergeOptionMerge ||
		p.ValuesMerge == valuesMergeOptionOverride {
		if err = p.replaceValuesInline(pValues); err != nil {
			return "", err
		}
	}
	var b []byte
	b, err = yaml.Marshal(p.ValuesInline)
	if err != nil {
		return "", err
	}
	return p.writeValuesBytes(b, path)
}

func (p *HelmChartInflationGeneratorPlugin) replaceValuesInline(pValues []byte) error {
	var err error
	chValues := make(map[string]interface{})
	if err = yaml.Unmarshal(pValues, &chValues); err != nil {
		return err
	}
	switch p.ValuesMerge {
	case valuesMergeOptionOverride:
		err = mergo.Merge(
			&chValues, p.ValuesInline, mergo.WithOverride)
	case valuesMergeOptionMerge:
		err = mergo.Merge(&chValues, p.ValuesInline)
	}
	p.ValuesInline = chValues
	return err
}

// Write a absolute path file in the tmp file system.
func (p *HelmChartInflationGeneratorPlugin) writeValuesBytes(
	b []byte, path string) (string, error) {
	if err := p.establishTmpDir(); err != nil {
		return "", fmt.Errorf("cannot create tmp dir to write helm values")
	}
	// use a hash of the provided path to generate a unique, valid filename
	hash := md5.Sum([]byte(path))
	newPath := filepath.Join(p.tmpDir, p.Name+"-kustomize-values-"+hex.EncodeToString(hash[:])+".yaml")
	return newPath, os.WriteFile(newPath, b, 0644)
}

func (p *HelmChartInflationGeneratorPlugin) cleanup() {
	if p.tmpDir != "" {
		os.RemoveAll(p.tmpDir)
	}
	if isOciRepo(p.Repo) && p.password != "" {
		// log out of the registry
		p.runHelmCommand([]string{
			"registry",
			"logout",
			p.Registry,
		})
	}
}

func (p *HelmChartInflationGeneratorPlugin) Generate() (objects fn.KubeObjects, err error) {
	defer p.cleanup()
	if err = p.checkHelmVersion(); err != nil {
		return nil, err
	}
	if _, exists := p.chartExistsLocally(); !exists {
		var pullArgs []string
		if isOciRepo(p.Repo) {
			pullArgs, err = p.pullOCIRepo()
			if err != nil {
				return nil, err
			}
		} else {
			pullArgs, err = p.pullNonOCIRepo()
			if err != nil {
				return nil, err
			}
		}
		if _, err := p.runHelmCommand(pullArgs); err != nil {
			return nil, err
		}
	}

	if err := p.processValuesFiles(); err != nil {
		return nil, err
	}
	var stdout []byte
	stdout, err = p.runHelmCommand(p.templateArgs())
	if err != nil {
		return nil, err
	}

	r := &kio.ByteReader{Reader: bytes.NewBufferString(string(stdout)), OmitReaderAnnotations: true}
	nodes, err := r.Read()
	if err != nil {
		return nil, err
	}

	for i := range nodes {
		o, err := fn.ParseKubeObject([]byte(nodes[i].MustString()))
		if err != nil {
			if strings.Contains(err.Error(), "expected exactly one object, got 0") {
				// sometimes helm produces some messages in between resources, we can safely
				// ignore these
				continue
			}
			return nil, fmt.Errorf("failed to parse %s: %s", nodes[i].MustString(), err.Error())
		}
		objects = append(objects, o)
	}

	return objects, nil
}

func (p *HelmChartInflationGeneratorPlugin) processValuesFiles() error {
	var valuesFiles []string
	for _, valuesFile := range p.ValuesFiles {
		file, err := p.createNewMergedValuesFiles(valuesFile)
		if err != nil {
			return err
		}
		valuesFiles = append(valuesFiles, file)
	}
	p.ValuesFiles = valuesFiles
	return nil
}

func (p *HelmChartInflationGeneratorPlugin) pullNonOCIRepo() ([]string, error) {
	repoAddArgs, hash := p.repoAddArgs()
	if repoAddArgs != nil {
		if _, err := p.runHelmCommand(repoAddArgs); err != nil {
			return nil, err
		}
	}
	args := []string{
		"pull",
		"--untar",
		"--untardir", p.absChartHome()}
	if p.Name != "" {
		args = append(args, hash+"/"+p.Name)
	}
	if p.Version != "" {
		args = append(args, "--version", p.Version)
	}
	return args, nil
}

func (p *HelmChartInflationGeneratorPlugin) templateArgs() []string {
	args := []string{"template"}
	if p.ReleaseName != "" {
		args = append(args, p.ReleaseName)
	}
	if p.Namespace != "" {
		args = append(args, "--namespace", p.Namespace)
	}
	if p.NameTemplate != "" {
		args = append(args, "--name-template", p.NameTemplate)
	}
	if p.Name != "" {
		args = append(args, filepath.Join(p.absChartHome(), p.Name))
	}
	for _, valuesFile := range p.ValuesFiles {
		args = append(args, "-f", valuesFile)
	}
	for _, apiVer := range p.ApiVersions {
		args = append(args, "--api-versions", apiVer)
	}
	if p.ReleaseName == "" {
		// AFAICT, this doesn't work as intended due to a bug in helm.
		// See https://github.com/helm/helm/issues/6019
		// I've tried placing the flag before and after the name argument.
		args = append(args, "--generate-name")
	}
	if p.Description != "" {
		args = append(args, "--description", p.Description)
	}
	if p.IncludeCRDs {
		args = append(args, "--include-crds")
	}
	if p.SkipTests {
		args = append(args, "--skip-tests")
	}
	return args
}

func (p *HelmChartInflationGeneratorPlugin) repoAddArgs() ([]string, string) {
	if p.Repo != "" {
		hash := md5.Sum([]byte(p.Repo))
		strHash := hex.EncodeToString(hash[:])
		args := []string{
			"repo",
			"add",
			strHash,
			p.Repo,
		}
		if p.password != "" && p.username != "" {
			args = append(args, []string{
				"--password", p.password,
				"--username", p.username,
			}...)
		}
		return args, strHash
	}
	return nil, ""
}

func (p *HelmChartInflationGeneratorPlugin) registryLoginArgs() []string {
	if p.Repo != "" {
		args := []string{
			"registry",
			"login",
			p.Registry,
		}
		if p.password != "" && p.username != "" {
			args = append(args, []string{
				"--password", p.password,
				"--username", p.username,
			}...)
		}
		return args
	}
	return nil
}

func (p *HelmChartInflationGeneratorPlugin) pullOCIRepo() ([]string, error) {
	if p.password != "" { // credentials provided, so we attempt to login
		repoLoginArgs := p.registryLoginArgs()
		if repoLoginArgs != nil {
			if _, err := p.runHelmCommand(repoLoginArgs); err != nil {
				return nil, err
			}
		}
	}
	args := []string{
		"pull",
		"--untar",
		"--untardir", p.absChartHome(),
		// OCI pull combine the repo and the chart name into one URL
		p.Repo + "/" + p.Name,
	}
	if p.Version != "" {
		args = append(args, "--version", p.Version)
	}
	return args, nil
}

// chartExistsLocally will return true if the chart does exist in
// local chart home.
func (p *HelmChartInflationGeneratorPlugin) chartExistsLocally() (string, bool) {
	path := filepath.Join(p.absChartHome(), p.Name)
	s, err := os.Stat(path)
	if err != nil {
		return path, false
	}
	return path, s.IsDir()
}

// checkHelmVersion will return an error if the helm version is not V3
func (p *HelmChartInflationGeneratorPlugin) checkHelmVersion() error {
	stdout, err := p.runHelmCommand([]string{"version", "-c", "--short"})
	if err != nil {
		return err
	}
	r, err := regexp.Compile(`v?\d+(\.\d+)+`)
	if err != nil {
		return err
	}
	v := r.FindString(string(stdout))
	if v == "" {
		return fmt.Errorf("cannot find version string in %s", string(stdout))
	}
	if v[0] == 'v' {
		v = v[1:]
	}
	majorVersion := strings.Split(v, ".")[0]
	if majorVersion != "3" {
		return fmt.Errorf("this HelmChartInflationGeneratorPlugin requires helm V3 but got v%s", v)
	}
	return nil
}

func isOciRepo(repo string) bool {
	return strings.HasPrefix(repo, "oci://")
}
