package transformer

import (
	"fmt"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-labels/transformer"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

const (
	AppNameKey = "app.kubernetes.io/name"
	builtInName = "builtin"
	SetLabelFnConfigKind = "SetLabels"
)

func SetDefaultLabels(rl *fn.ResourceList) (bool, error) {
	rl.FunctionConfig = NewFnConfigFromKptfile(rl.Items)
	if rl.FunctionConfig == nil {
		rl.LogResult(fmt.Errorf("no Kptfile found from the input resources. Require at least one Kptfile"))
		return true, nil
	}
	return transformer.SetLabels(rl)
}

// AddAppName use the root Kptfile `metadata.name` as the "app.kubernetes.io/name" label value.
// This convention believes that all KRM resources under the same a kpt package should be served for a specific application.
// A kpt package can contain nested sub packages. An application can be composed by other apps.
// So we adds the app name label based on their root Kptfile.
func AddAppName(recommendedLabels map[string]string, items fn.KubeObjects) error {
	rootKptfile := items.GetRootKptfile()
	if rootKptfile == nil {
		return fmt.Errorf("no Kptfile found from the input resources. Require at least one Kptfile")
	}
	recommendedLabels[AppNameKey] = rootKptfile.GetName()
	return nil
}

// NewFnConfigFromKptfile generates the `SetLabel` object as the `functionConfig` for SetLabel transformer.
// See https://catalog.kpt.dev/set-labels/v0.1/ about the `SetLabels` kind structure and usage.
func NewFnConfigFromKptfile(items fn.KubeObjects) *fn.KubeObject {
	fnConfig := fn.NewEmptyKubeObject()
	fnConfig.SetAPIVersion(fn.KptFunctionApiVersion)
	fnConfig.SetKind(SetLabelFnConfigKind)
	fnConfig.SetName(builtInName)

	recommendedLabels := map[string]string{}
	if err := AddAppName(recommendedLabels, items); err != nil {
		return	nil
	}
	fnConfig.SetNestedStringMapOrDie(recommendedLabels,"labels")
	return fnConfig
}
