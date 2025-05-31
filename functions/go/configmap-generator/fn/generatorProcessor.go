package fn

import (
	"fmt"
	"reflect"
)

type generatorProcessor struct {
	fnGenerator Generator
}

func (r generatorProcessor) Process(rl *ResourceList) (bool, error) {
	ctx := &Context{results: &rl.Results}
	r.config(ctx, rl.FunctionConfig)
	rl.Items = r.fnGenerator.Generate(ctx, rl.FunctionConfig, rl.Items)
	return true, nil
}

func (r *generatorProcessor) config(ctx *Context, o *KubeObject) {
	fnName := reflect.ValueOf(r.fnGenerator).Elem().Type().Name()
	switch true {
	case o.IsEmpty():
		ctx.Result("`FunctionConfig` is not given", Info)
	case o.IsGVK("", "v1", "ConfigMap"):
		data := o.NestedStringMapOrDie("data")
		fnRunnerElem := reflect.ValueOf(r.fnGenerator).Elem()
		for i := 0; i < fnRunnerElem.NumField(); i++ {
			if fnRunnerElem.Field(i).Kind() == reflect.Map {
				fnRunnerElem.Field(i).Set(reflect.ValueOf(data))
				break
			}
		}
	case o.IsGVK("fn.kpt.dev", "v1alpha1", fnName):
		o.AsOrDie(r.fnGenerator)
	default:
		ctx.ResultErrAndDie(fmt.Sprintf("unknown FunctionConfig `%v`, expect %v", o.GetKind(), fnName), o)
	}
}
