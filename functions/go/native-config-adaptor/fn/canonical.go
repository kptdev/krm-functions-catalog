package fn

func NewNonKrmResource() *NonKrmResource {
	resource := &NonKrmResource{}
	resource.Kind = NonKrmKind
	resource.APIVersion = KptFunctionApiVersion
	return resource
}

type NonKrmResource struct {
	TypeMeta `json:",inline" yaml:",inline"`
	NameMeta `json:",inline" yaml:",inline"`
	Spec     NonKrmResourceSpec `json:"spec,omitempty" yaml:"spec,omitempty"`
}

type NonKrmResourceSpec struct {
	FileName  string `json:"fileName,omitempty" yaml:"fileName,omitempty"`
	LocalPath string `json:"localPath,omitempty" yaml:"localPath,omitempty"`
	Content   string `json:"content,omitempty" yaml:"content,omitempty"`
}

// TypeMeta partially copies apimachinery/pkg/apis/meta/v1.TypeMeta
// No need for a direct dependence; the fields are stable.
type TypeMeta struct {
	// APIVersion is the apiVersion field of a Resource
	APIVersion string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	// Kind is the kind field of a Resource
	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`
}

// NameMeta contains name information.
type NameMeta struct {
	// Name is the metadata.name field of a Resource
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Namespace is the metadata.namespace field of a Resource
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}
