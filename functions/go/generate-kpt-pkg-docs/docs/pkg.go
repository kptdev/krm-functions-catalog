package docs

import (
	"fmt"
	"strings"

	kptfilev1 "github.com/kptdev/kpt/pkg/api/kptfile/v1"
	"github.com/kptdev/kpt/pkg/kptfile/kptfileutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const bpTitleAnnotation = "blueprints.cloud.google.com/title"

// findPkgs finds all kptfiles and associated pkg paths
func findPkgs(nodes []*yaml.RNode) (map[string]*kptfilev1.KptFile, error) {
	kptfiles := map[string]*kptfilev1.KptFile{}
	for _, node := range nodes {
		if node.GetKind() == kptfilev1.TypeMeta.Kind {
			s, err := node.String()
			if err != nil {
				return nil, err
			}
			kf, err := kptfileutil.DecodeKptfile(strings.NewReader(s))
			if err != nil {
				return nil, fmt.Errorf("failed to decode Kptfile: %w", err)
			}
			kfPath, err := findResourcePath(node)
			if err != nil {
				return nil, err
			}
			kptfiles[kfPath] = kf
		}
	}
	return kptfiles, nil
}

// getFnCfgPaths returns function config filepaths in a Kptfile
func getFnCfgPaths(kf *kptfilev1.KptFile) []string {
	if kf.Pipeline == nil {
		return nil
	}
	fnCfgPaths := []string{}
	for _, fn := range kf.Pipeline.Mutators {
		if fn.ConfigPath != "" {
			fnCfgPaths = append(fnCfgPaths, fn.ConfigPath)
		}
	}
	for _, fn := range kf.Pipeline.Validators {
		if fn.ConfigPath != "" {
			fnCfgPaths = append(fnCfgPaths, fn.ConfigPath)
		}
	}
	return fnCfgPaths
}

// getBlueprintTitle returns the title of a blueprint as markdown heading falling back to pkg name
func getBlueprintTitle(kf *kptfilev1.KptFile) string {
	title, exists := kf.Annotations[bpTitleAnnotation]
	if exists {
		return getMdHeading(title, 1)
	}
	return getMdHeading(kf.Name, 1)
}
