package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	kptfilev1 "github.com/kptdev/kpt/pkg/api/kptfile/v1"
	"github.com/kptdev/kpt/pkg/kptfile/kptfileutil"
)

var (
	kccAPIVersionRegex = regexp.MustCompile(`^([\w]+)\.cnrm\.cloud\.google\.com\/[\w]+$`)
)

const (
	projectIDSetterName = "project-id"
	projectIDAnnotation = "cnrm.cloud.google.com/project-id"
)

func findSetterNode(nodes []*yaml.RNode, path string) (*yaml.RNode, error) {
	for _, node := range nodes {
		np := node.GetAnnotations()[kioutil.PathAnnotation]
		if np == path {
			return node, nil
		}
	}
	return nil, fmt.Errorf(`file %s doesn't exist, please ensure the file specified in "configPath" exists and retry`, path)
}

func findKptfiles(nodes []*yaml.RNode) ([]*kptfilev1.KptFile, error) {
	kptfiles := []*kptfilev1.KptFile{}
	for _, node := range nodes {
		if node.GetKind() == kptfilev1.TypeMeta.Kind {
			s, err := node.String()
			if err != nil {
				return nil, fmt.Errorf("unable to read Kptfile: %w", err)
			}
			kf, err := kptfileutil.DecodeKptfile(strings.NewReader(s))
			if err != nil {
				return nil, fmt.Errorf("failed to read Kptfile: %w", err)
			}
			kptfiles = append(kptfiles, kf)
		}
	}
	if len(kptfiles) == 0 {
		return nil, fmt.Errorf("unable to find Kptfile, please include --include-meta-resources flag if a Kptfile is present")
	}
	return kptfiles, nil
}

func setKptfile(nodes []*yaml.RNode, kf *kptfilev1.KptFile) error {
	b, err := yaml.Marshal(kf)
	if err != nil {
		return fmt.Errorf("failed to marshal updated Kptfile: %w", err)
	}
	kNode, err := yaml.Parse(string(b))
	if err != nil {
		return fmt.Errorf("failed to parse updated Kptfile: %w", err)
	}

	for i := range nodes {
		if nodes[i].GetAnnotations()[kioutil.PathAnnotation] == kf.Annotations[kioutil.PathAnnotation] {
			nodes[i] = kNode
			return nil
		}
	}
	return nil

}

func setSettersOnKptfile(nodes []*yaml.RNode, kf *kptfilev1.KptFile, projectID string) error {
	if kf.Pipeline == nil {
		kf.Pipeline = &kptfilev1.Pipeline{}
	}
	for _, fn := range kf.Pipeline.Mutators {
		if !strings.Contains(fn.Image, "apply-setters") {
			continue
		}
		if fn.ConfigMap != nil {
			if fn.ConfigMap[projectIDSetterName] == "" {
				fn.ConfigMap[projectIDSetterName] = projectID
			}
			if err := setKptfile(nodes, kf); err != nil {
				return fmt.Errorf("failed to update Kptfile file: %w", err)
			}
			return nil
		} else if fn.ConfigPath != "" {
			settersConfig, err := findSetterNode(nodes, fn.ConfigPath)
			if err != nil {
				return fmt.Errorf("failed to find setter file: %w", err)
			}
			dataMap := settersConfig.GetDataMap()
			if dataMap[projectIDSetterName] == "" {
				dataMap[projectIDSetterName] = projectID
				settersConfig.SetDataMap(dataMap)
			}
			return nil
		} else {
			return fmt.Errorf("unable to find `ConfigMap` or `configPath` as the `functionConfig` for apply-setters")
		}
	}

	fn := kptfilev1.Function{
		Image: "ghcr.io/kptdev/krm-functions-catalog/apply-setters:v0.2",
		ConfigMap: map[string]string{
			projectIDSetterName: projectID,
		},
	}
	kf.Pipeline.Mutators = append(kf.Pipeline.Mutators, fn)
	if err := setKptfile(nodes, kf); err != nil {
		return fmt.Errorf("failed to update Kptfile file: %w", err)
	}

	return nil
}

func setSetters(nodes []*yaml.RNode, projectID string) error {
	kptfiles, err := findKptfiles(nodes)
	if err != nil {
		return fmt.Errorf("failed to find Kptfile: %v", err)
	}

	for _, kf := range kptfiles {
		if err := setSettersOnKptfile(nodes, kf, projectID); err != nil {
			return fmt.Errorf("error updating Kptfile: %w", err)
		}
	}

	return nil
}

func setProjectIDAnnotation(nodes []*yaml.RNode, projectID string) error {
	for _, node := range nodes {
		matches := kccAPIVersionRegex.FindStringSubmatch(node.GetApiVersion())
		// Check if it's a Config Connector resource (apiVersion: *.cnrm.cloud.google.com/*).
		// Ignore Config Connector system resources (apiVersion: core.cnrm.cloud.google.com/*).
		if len(matches) == 2 && matches[1] != "core" {
			annotations := node.GetAnnotations()
			if _, ok := annotations[projectIDAnnotation]; !ok {
				annotations[projectIDAnnotation] = projectID
				if err := node.SetAnnotations(annotations); err != nil {
					return fmt.Errorf("failed to set project-id annotation: %w", err)
				}
				continue
			}
		}
	}
	return nil
}

type projectIDProcessor struct{}

func (p *projectIDProcessor) Process(resourceList *framework.ResourceList) error {
	dm := resourceList.FunctionConfig.GetDataMap()
	projectID := dm[projectIDSetterName]

	if err := setSetters(resourceList.Items, projectID); err != nil {
		return fmt.Errorf("failed to set project-id setter: %w", err)
	}
	if err := setProjectIDAnnotation(resourceList.Items, projectID); err != nil {
		return fmt.Errorf("failed to set project-id annotation: %w", err)
	}

	return nil
}

func main() {
	pp := projectIDProcessor{}
	cmd := command.Build(&pp, command.StandaloneEnabled, false)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
