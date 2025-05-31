package fn

type Generator interface {
	Generate(context *Context, functionConfig *KubeObject, items KubeObjects) KubeObjects
}
