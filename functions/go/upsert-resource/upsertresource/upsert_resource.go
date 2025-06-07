package upsertresource

import (
	"strings"

	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// UpsertResource upserts resources to input list of resources
// resources are uniquely identifies by Group, Kind, Name, Namespace and input file path
type UpsertResource struct {
	// List input resources to upsert
	List *yaml.RNode
}

const (
	DestPathAnnotation = "config.kubernetes.io/target-path"
	ListKind           = "List"
)

// Filter implements UpsertResource as a yaml.Filter
func (ur UpsertResource) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	resources, err := unwrap(ur.List)
	if err != nil {
		return nodes, err
	}
	for _, resource := range resources {
		nodes, err = UpsertSingleResource(nodes, resource)
		if err != nil {
			return nodes, err
		}
	}
	return nodes, nil
}

// unwrap unwraps List and returns the list to input resources to upsert
func unwrap(node *yaml.RNode) ([]*yaml.RNode, error) {
	rm, err := node.GetMeta()
	if err != nil {
		return nil, err
	}
	var resources []*yaml.RNode
	if rm.Kind == ListKind {
		items := node.Field("items")
		if items != nil {
			for i := range items.Value.Content() {
				// add items
				resources = append(resources, yaml.NewRNode(items.Value.Content()[i]))
			}
			return resources, nil
		}
		// the list has no items
		return []*yaml.RNode{}, nil
	}
	// the resource is a plain single node to upsert
	return []*yaml.RNode{node}, nil
}

// UpsertSingleResource upserts input resource to the list of nodes
func UpsertSingleResource(nodes []*yaml.RNode, resource *yaml.RNode) ([]*yaml.RNode, error) {
	replaced, err := ReplaceResource(nodes, resource)
	if err != nil {
		return nodes, err
	}
	if !replaced {
		return AddResource(nodes, resource)
	}
	return nodes, nil
}

// ReplaceResource checks if the inputResource matches with any of the resources in nodes list
// and replaces it with inputResource
// nodes are matched based on Group, Kind, Name and Namespace
func ReplaceResource(nodes []*yaml.RNode, inputResource *yaml.RNode) (bool, error) {
	found := false
	inputMeta, err := inputResource.GetMeta()
	if err != nil {
		return false, err
	}
	for i := range nodes {
		rMeta, err := nodes[i].GetMeta()
		if err != nil {
			return false, err
		}
		// skip processing resource if it is a function config
		// TODO: remove this check after we stop support for v0.X
		if IsFunctionConfig(rMeta) {
			continue
		}
		// check if there is a match and replace the resource
		if IsSameResource(inputMeta, rMeta) {
			nodes[i] = inputResource.Copy()
			a := combineInputAndMatchedAnnotations(inputMeta.Annotations, rMeta.Annotations)
			err = nodes[i].SetAnnotations(a)
			if err != nil {
				return false, err
			}
			// found a matching resource
			// but continue to replace other instances of the resource
			found = true
		}
	}
	return found, nil
}

// AddResource appends the inputResource to the list of input nodes
// it also cleans up the meta annotations so that resource is created in new file
// by the function orchestrator
func AddResource(nodes []*yaml.RNode, inputResource *yaml.RNode) ([]*yaml.RNode, error) {
	newNode := inputResource.Copy()
	meta, err := newNode.GetMeta()
	if err != nil {
		return nodes, err
	}
	// remove function, path and index annotations from the result
	// removing path and index annotations makes orchestrator write resource
	// to a new file
	removeFnPathIndexAnnotations(meta.Annotations)
	path := inputResource.GetAnnotations()[DestPathAnnotation]
	if path != "" {
		meta.Annotations[kioutil.PathAnnotation] = path
	}
	delete(meta.Annotations, DestPathAnnotation)
	err = newNode.SetAnnotations(meta.Annotations)
	if err != nil {
		return nodes, err
	}
	nodes = append(nodes, newNode)
	return nodes, nil
}

// IsSameResource returns true if metadata of two resources
// have same Group, Kind, Name, Namespace
// TODO: phanimarupaka move this to common util https://github.com/kptdev/kpt/issues/2043
func IsSameResource(inputResourceMeta, targetResourceMeta yaml.ResourceMeta) bool {
	g1, _ := ParseGroupVersion(inputResourceMeta.APIVersion)
	g2, _ := ParseGroupVersion(targetResourceMeta.APIVersion)
	return g1 == g2 && inputResourceMeta.Kind == targetResourceMeta.Kind &&
		inputResourceMeta.Name == targetResourceMeta.Name &&
		inputResourceMeta.Namespace == targetResourceMeta.Namespace &&
		upsertPathMatch(inputResourceMeta, targetResourceMeta)
}

// upsertPathMatch checks if the target-path specified by user in input resource matches
// the path of target resource
func upsertPathMatch(inputResourceMeta, targetResourceMeta yaml.ResourceMeta) bool {
	return inputResourceMeta.Annotations[DestPathAnnotation] == "" ||
		inputResourceMeta.Annotations[DestPathAnnotation] == targetResourceMeta.Annotations[kioutil.PathAnnotation]
}

// ParseGroupVersion parses a KRM metadata apiVersion field.
func ParseGroupVersion(apiVersion string) (group, version string) {
	if i := strings.Index(apiVersion, "/"); i > -1 {
		return apiVersion[:i], apiVersion[i+1:]
	}
	return "", apiVersion
}

// combineInputAndMatchedAnnotations combines user provided non-meta annotations from inputResource,
// with path and index annotations from matched resource and returns the result
func combineInputAndMatchedAnnotations(inputResourceAnno, matchedResourceAnno map[string]string) map[string]string {
	if inputResourceAnno == nil {
		inputResourceAnno = make(map[string]string)
	}
	if matchedResourceAnno == nil {
		matchedResourceAnno = make(map[string]string)
	}
	res := make(map[string]string)
	// retain the annotations from the input resource in fn-config,
	// these should be written to matched resource
	for k, v := range inputResourceAnno {
		res[k] = v
	}
	// retain the path and index annotation from matched resource to result
	res[kioutil.PathAnnotation] = matchedResourceAnno[kioutil.PathAnnotation]
	res[kioutil.IndexAnnotation] = matchedResourceAnno[kioutil.IndexAnnotation]
	// remove function meta annotations from the result as they should
	// not be written to output resource
	removeFnAnnotations(res)
	// this annotation should be used to only determine the path to write resource to
	// and should be removed from the output resource
	delete(res, DestPathAnnotation)
	return res
}

// removeFnPathIndexAnnotations removes index, path and fn annotations
func removeFnPathIndexAnnotations(a map[string]string) {
	removeFnAnnotations(a)
	delete(a, kioutil.PathAnnotation)
	delete(a, kioutil.IndexAnnotation)
}

// removeFnAnnotations removes fn annotations
// TODO: phanimarupaka remove this method after we drop support for kpt 0.X
func removeFnAnnotations(a map[string]string) {
	delete(a, runtimeutil.FunctionAnnotationKey)
	// using hard coded key as this annotation is deprecated and not exposed by kyaml
	delete(a, "config.k8s.io/function")
}

// IsFunctionConfig returns true if input resource meta has function config annotation
func IsFunctionConfig(rMeta yaml.ResourceMeta) bool {
	return rMeta.Annotations != nil &&
		(rMeta.Annotations[runtimeutil.FunctionAnnotationKey] != "" || rMeta.Annotations["config.k8s.io/function"] != "")
}
