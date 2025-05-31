package transformer

import (
	"fmt"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-labels/setlabels"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ fn.Runner = &SetStandardLabels{}

type SetStandardLabels struct {
	forDeployment string
	// TBD: validate whether the recommended labels from the ResourceList.Items are set as expected.
	validateOnly bool
}

// FnConfigFromLabel creates a FunctionConfig object for `set-label` KRM function based on the recommended labels `data`.
func FnConfigFromLabels(data map[string]string) *fn.KubeObject {
	fnConfig := fn.NewEmptyKubeObject()
	fnConfig.SetKind(SetLabelFnKind)
	fnConfig.SetAPIVersion("v1")
	fnConfig.SetName(SetLabelFnName)
	fnConfig.SetAnnotation(fn.KptLocalConfig, "true")
	fnConfig.SetNestedStringMap(data, "data")
	return fnConfig
}

// RecommendedLabels find out the recommended labels for the given package.
func (r *SetStandardLabels) RecommendedLabels(items fn.KubeObjects) (map[string]string, error) {
	data := map[string]string{}

	// Get root package name from Kptfile
	kptfile := items.GetRootKptfile()
	if kptfile == nil {
		return data, fmt.Errorf("no `Kptfile` are found. This function is for kpt pacakge. please run `kpt pkg init`")
	}
	pkgName := kptfile.GetName()

	// Assign blueprint pkg name to "app.kubernetes.io/name" and deployment pkg name to "app.kubernetes.io/instance"
	if r.IsDeploymentPackage(items) {
		data[AppInstance] = pkgName
	} else {
		data[AppName] = pkgName
	}

	return data, nil
}

func (r *SetStandardLabels) IsDeploymentPackage(items fn.KubeObjects) bool {
	if r.forDeployment == "" {
		// We check the existence of package-context.yaml file to tell whether it's a deployable package or not.
		// This package-context.yaml is a usage of variant construction that serves for deployment purpose.
		// If the usage of package-context.yaml is changed, this function should change the condition as well.
		forDeployments := items.Where(fn.IsGroupKind(schema.GroupKind{Kind: PackageContextKind})).Where(fn.IsName(PackageContextName))
		if len(forDeployments) == 0 {
			return false
		}
		return true
	}
	return strings.ToLower(r.forDeployment) == "true"
}

func NewSetLabelTransformer(fnConfig *fn.KubeObject) *setlabels.SetLabels {
	data, _, _ := fnConfig.NestedStringMap("data")
	return &setlabels.SetLabels{Labels: data}
}

// Run is the main function that leverages another transformer function `SetLabels` to update recommended labels
// that follows some kpt specific conventions:
// - "app.kubernetes.io/name" comes from the current blueprint package name, or (if current is deployment) the upstream blueprint package.
// - "app.kubernetes.io/instance" is only set for deployment package, using its package name.
func (r *SetStandardLabels) Run(ctx *fn.Context, _ *fn.KubeObject, items fn.KubeObjects, results *fn.Results) bool {
	recommendedLabels, err := r.RecommendedLabels(items)
	if err != nil {
		results.ErrorE(err)
		return false
	}
	newFnConfig := FnConfigFromLabels(recommendedLabels)
	setLabelsRunner := NewSetLabelTransformer(newFnConfig)

	return setLabelsRunner.Run(ctx, newFnConfig, items, results)
}
