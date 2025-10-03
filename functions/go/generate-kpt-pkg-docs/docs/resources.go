package docs

import (
	"fmt"
	"path"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const k8sRefURL = "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22"

// findResourcePath attempts to find the path of a given resource using the PathAnnotation with fallback to using LegacyPathAnnotation
func findResourcePath(r *yaml.RNode) (string, error) {
	anno := r.GetAnnotations()
	path, exists := anno[kioutil.PathAnnotation]
	if exists {
		return path, nil
	}
	//nolint
	path, exists = anno[kioutil.LegacyPathAnnotation]
	if exists {
		return path, nil
	}
	res := resid.NewResIdWithNamespace(resid.GvkFromNode(r), r.GetName(), r.GetNamespace())
	return "", fmt.Errorf("unable find resource path for %s", res.Gvk.String())
}

// getResourceDocsLink returns documentation link for a given GVK
func getResourceDocsLink(r resid.Gvk) string {
	switch {
	case r.Kind == "ConfigConnectorContext" && r.Group == "core.cnrm.cloud.google.com":
		// KCC config. No reference doc yet.
		// https://github.com/GoogleCloudPlatform/k8s-config-connector/issues/344
		return getMdLink(r.Kind, "https://cloud.google.com/config-connector/docs/how-to/advanced-install#addon-configuring")
	case r.Kind == "ConfigManagement" && r.Group == "configmanagement.gke.io":
		// ACM Operator config
		return getMdLink(r.Kind, "https://cloud.google.com/anthos-config-management/docs/configmanagement-fields")
	case strings.HasSuffix(r.Group, ".cnrm.cloud.google.com"):
		// KCC resource
		groupPrefix := strings.Split(r.Group, ".")
		// possibly invalid KCC resource
		if len(groupPrefix) < 1 {
			return ""
		}
		kccDoc := fmt.Sprintf("https://cloud.google.com/config-connector/docs/reference/resource-docs/%s/%s", groupPrefix[0], strings.ToLower(r.Kind))
		return getMdLink(r.Kind, kccDoc)
	case r.Group == "" && r.Version == "v1":
		// K8s core resource
		anchor := fmt.Sprintf("%s-%s-core", strings.ToLower(r.Kind), r.Version)
		return getMdLink(r.Kind, fmt.Sprintf("%s/#%s", k8sRefURL, anchor))
	case strings.HasSuffix(r.Group, ".k8s.io"):
		// K8s resource
		anchor := fmt.Sprintf("%s-%s-%s", strings.ToLower(r.Kind), r.Version, strings.ReplaceAll(r.Group, ".", "-"))
		return getMdLink(r.Kind, fmt.Sprintf("%s/#%s", k8sRefURL, anchor))
	}
	return ""
}

// filterResources returns a slice of resources skipping any resouces in skipFiles
func filterResources(nodes []*yaml.RNode, skipFiles map[string]bool) []*yaml.RNode {
	filtered := []*yaml.RNode{}
	for _, r := range nodes {
		if !shouldSkipResource(r, skipFiles) {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// shouldSkipResource returns true if resource path is present in skipFiles
func shouldSkipResource(r *yaml.RNode, skipFiles map[string]bool) bool {
	rPath, err := findResourcePath(r)
	if err != nil {
		return true
	}
	return skipFiles[rPath] || skipFiles[path.Dir(rPath)]
}
