package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/kptdev/krm-functions-catalog/archived/functions/go/bind/pkg/rename"
	"github.com/kptdev/krm-functions-catalog/archived/functions/go/set-gcp-resource-ids/pkg/kpt"
	"github.com/kptdev/krm-functions-sdk/go/fn"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func main() {
	// TODO: fn.AsMain should support an "easy mode" where it runs against a directory
	if err := fn.AsMain(fn.ResourceListProcessorFunc(Run)); err != nil {
		os.Exit(1)
	}
}

type SetNamePrefix struct {
	Prefix    string `json:"prefix,omitempty"`
	OldPrefix string `json:"oldPrefix,omitempty"`
}

func Run(rl *fn.ResourceList) (bool, error) {
	f := SetNamePrefix{
		OldPrefix: "packagename",
	}

	packageContext, err := kpt.FindPackageContext(rl.Items)
	if err != nil {
		return false, err
	}
	f.Prefix = packageContext.Name

	if err := f.LoadConfig(rl.FunctionConfig); err != nil {
		rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(fmt.Errorf("functionConfig error: %w", err), rl.FunctionConfig))
		return true, nil
	}

	if f.Prefix == "" {
		return false, fmt.Errorf("`data.prefix` should not be empty")
	}

	if err := f.Transform(rl.Items); err != nil {
		rl.Results = append(rl.Results, fn.ErrorResult(err))
	}
	return true, nil
}

func (f *SetNamePrefix) LoadConfig(fnConfig *fn.KubeObject) error {
	if fnConfig != nil {
		switch { //TODO: fnConfig.GroupVersionKind()
		case fnConfig.IsGVK("", "v1", "ConfigMap"):
			data := fnConfig.UpsertMap("data") // TODO: Why does GetMap fail?
			if s, ok, _ := data.NestedString("prefix"); ok {
				f.Prefix = s
			}
			if s, ok, _ := data.NestedString("oldPrefix"); ok {
				f.OldPrefix = s
			}

		default:
			gvk := schema.GroupVersionKind{}
			return fmt.Errorf("unknown functionConfig Kind %v", gvk) //TODO: fnConfig.GroupVersionKind())
		}
	}

	return nil
}

func (f *SetNamePrefix) Transform(objects fn.KubeObjects) error {
	for _, object := range objects {
		if object.IsLocalConfig() {
			continue
		}
		if kpt.IsResourceGroup(object) {
			continue // Should ResourceGroup be marked as local config?
		}
		oldName := object.GetName()
		newNamespace := object.GetNamespace() // do not change namespace
		if oldName != "" {
			if oldName == f.OldPrefix {
				newName := f.Prefix
				if err := rename.Rename(object, newName, newNamespace, objects); err != nil {
					return err
				}
			} else if strings.HasPrefix(oldName, f.OldPrefix+"-") {
				suffix := strings.TrimPrefix(oldName, f.OldPrefix+"-")
				newName := f.Prefix + "-" + suffix
				if err := rename.Rename(object, newName, newNamespace, objects); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
