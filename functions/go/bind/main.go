package main

import (
	"fmt"
	"os"

	"github.com/kptdev/krm-functions-catalog/functions/go/bind/pkg/rename"
	"github.com/kptdev/krm-functions-sdk/go/fn"
)

func main() {
	// TODO: fn.AsMain should support an "easy mode" where it runs against a directory
	if err := fn.AsMain(fn.ResourceListProcessorFunc(Run)); err != nil {
		os.Exit(1)
	}
}

type Bind struct {
	Object *fn.KubeObject `json:"object,omitempty"`
}

func Run(rl *fn.ResourceList) (bool, error) {
	f := Bind{}

	err := f.LoadConfig(rl.FunctionConfig)
	if err != nil {
		rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(fmt.Errorf("functionConfig error: %w", err), rl.FunctionConfig))
		return true, nil
	}

	if f.Object == nil {
		return false, fmt.Errorf("`data.object` should not be empty")
	}

	if err := f.Transform(rl.Items); err != nil {
		rl.Results = append(rl.Results, fn.ErrorResult(err))
	}
	return true, nil
}

func (f *Bind) LoadConfig(fnConfig *fn.KubeObject) error {
	f.Object = fnConfig

	return nil
}

func (f *Bind) Transform(objects fn.KubeObjects) error {
	// Find the matching localconfig object
	var bindingObject *fn.KubeObject
	{
		// TODO: Match GK or GVK?
		gvk := f.Object.GroupVersionKind()

		matchAnnotations := map[string]string{
			"config.kubernetes.io/local-config": "binding",
		}
		matches := objects.Where(fn.HasAnnotations(matchAnnotations)).Where(fn.IsGVK(gvk.Group, gvk.Version, gvk.Kind))

		if len(matches) == 0 {
			return fmt.Errorf("no match found for binding object of kind %s/%s", f.Object.GetAPIVersion(), f.Object.GetKind())
		}
		if len(matches) != 1 {
			return fmt.Errorf("multiple matches found for binding object of kind %s/%s", f.Object.GetAPIVersion(), f.Object.GetKind())
		}
		bindingObject = matches[0]
	}

	// Sync the name and namespace to the target value
	if bindingObject.GetName() != f.Object.GetName() || bindingObject.GetNamespace() != f.Object.GetNamespace() {
		if err := rename.Rename(bindingObject, f.Object.GetName(), f.Object.GetNamespace(), objects); err != nil {
			return err
		}
	}

	return nil
}
