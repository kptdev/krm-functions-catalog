package transformer


// https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/
const (
	AppLabelPrefix = "app.kubernetes.io/"
	AppName = AppLabelPrefix + "name"
	AppInstance = AppLabelPrefix + "instance"

	// TBD
	AppVersion = AppLabelPrefix + "version"
	AppComponent = AppLabelPrefix + "component"
	AppPartOf = AppLabelPrefix + "part-of"
	AppManagedBy = AppLabelPrefix + "managed-by"
)

const (
	PackageContextKind = "ConfigMap"
	PackageContextName = "kptfile.kpt.dev"
)


const (
	SetLabelFnKind = "ConfigMap"
	SetLabelFnName = "recommended-labels"
)