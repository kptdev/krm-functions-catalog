package starlark

import (
	"github.com/kptdev/krm-functions-sdk/go/fn"
)

func Process(resourceList *fn.ResourceList) (bool, error) {
	err := func() error {
		sr := &StarlarkRun{}
		if err := sr.Config(resourceList.FunctionConfig); err != nil {
			return err
		}
		return sr.Transform(resourceList)
	}()
	if err != nil {
		resourceList.Results = []*fn.Result{
			{
				Message:  err.Error(),
				Severity: fn.Error,
			},
		}
		return false, nil
	}
	return true, nil
}
