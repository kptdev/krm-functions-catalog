package v1alpha1

import (
	"fmt"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// KptFileName is the name of the KptFile
const (
	KptFileName       = "Kptfile"
	KptFileGroup      = "kpt.dev"
	KptFileVersion    = "v1alpha1"
	KptFileAPIVersion = KptFileGroup + "/" + KptFileVersion
)

// TypeMeta is the TypeMeta for KptFile instances.
var TypeMeta = yaml.ResourceMeta{
	TypeMeta: yaml.TypeMeta{
		APIVersion: KptFileAPIVersion,
		Kind:       KptFileName,
	},
}

// KptFile contains information about a package managed with kpt
type KptFile struct {
	yaml.ResourceMeta `yaml:",inline"`

	// CloneFrom records where the package was originally cloned from
	Upstream *Upstream `yaml:"upstream,omitempty"`

	// PackageMeta contains information about the package
	PackageMeta *PackageMeta `yaml:"packageMetadata,omitempty"`

	Dependencies []Dependency `yaml:"dependencies,omitempty"`

	// OpenAPI contains additional schema for the resources in this package
	// Uses interface{} instead of Node to work around yaml serialization issues
	// See https://github.com/go-yaml/yaml/issues/518 and
	// https://github.com/go-yaml/yaml/issues/575
	OpenAPI interface{} `yaml:"openAPI,omitempty"`

	// Functions contains configuration for running functions
	Functions Functions `yaml:"functions,omitempty"`

	// Parameters for inventory object.
	Inventory *Inventory `yaml:"inventory,omitempty"`
}

// Inventory encapsulates the parameters for the inventory object. All of the
// the parameters are required if any are set.
type Inventory struct {
	Namespace string `yaml:"namespace,omitempty"`
	Name      string `yaml:"name,omitempty"`
	// Unique label to identify inventory object in cluster.
	InventoryID string            `yaml:"inventoryID,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

type Functions struct {
	// AutoRunStarlark will cause starlark functions to automatically be run.
	AutoRunStarlark bool `yaml:"autoRunStarlark,omitempty"`

	// StarlarkFunctions is a list of starlark functions to run
	StarlarkFunctions []StarlarkFunction `yaml:"starlarkFunctions,omitempty"`
}

type StarlarkFunction struct {
	// Name is the name that will be given to the program
	Name string `yaml:"name,omitempty"`
	// Path is the path to the *.star script to run
	Path string `yaml:"path,omitempty"`
}

type Dependency struct {
	Name            string `yaml:"name,omitempty"`
	Upstream        `yaml:",inline,omitempty"`
	EnsureNotExists bool       `yaml:"ensureNotExists,omitempty"`
	Strategy        string     `yaml:"updateStrategy,omitempty"`
	Functions       []Function `yaml:"functions,omitempty"`
	AutoSet         bool       `yaml:"autoSet,omitempty"`
}

type PackageMeta struct {
	// URL is the location of the package.  e.g. https://github.com/example/com
	URL string `yaml:"url,omitempty"`

	// Email is the email of the package maintainer
	Email string `yaml:"email,omitempty"`

	// License is the package license
	License string `yaml:"license,omitempty"`

	// Version is the package version
	Version string `yaml:"version,omitempty"`

	// Tags can be indexed and are metadata about the package
	Tags []string `yaml:"tags,omitempty"`

	// Man is the path to documentation about the package
	Man string `yaml:"man,omitempty"`

	// ShortDescription contains a short description of the package.
	ShortDescription string `yaml:"shortDescription,omitempty"`
}

// OriginType defines the type of origin for a package
type OriginType string

const (
	// GitOrigin specifies a package as having been cloned from a git repository
	GitOrigin   OriginType = "git"
	StdinOrigin OriginType = "stdin"
)

// Upstream defines where a package was cloned from
type Upstream struct {
	// Type is the type of origin.
	Type OriginType `yaml:"type,omitempty"`

	// Git contains information on the origin of packages cloned from a git repository.
	Git Git `yaml:"git,omitempty"`

	Stdin Stdin `yaml:"stdin,omitempty"`
}

type Stdin struct {
	FilenamePattern string `yaml:"filenamePattern,omitempty"`

	Original string `yaml:"original,omitempty"`
}

// Git contains information on the origin of packages cloned from a git repository.
type Git struct {
	// Commit is the git commit that the package was fetched at
	Commit string `yaml:"commit,omitempty"`

	// Repo is the git repository the package was cloned from.  e.g. https://
	Repo string `yaml:"repo,omitempty"`

	// RepoDirectory is the sub directory of the git repository that the package was cloned from
	Directory string `yaml:"directory,omitempty"`

	// Ref is the git ref the package was cloned from
	Ref string `yaml:"ref,omitempty"`
}

type Function struct {
	Config yaml.Node `yaml:"config,omitempty"`
	Image  string    `yaml:"image,omitempty"`
}

// ReadFile reads the KptFile node
func ReadFile(node *yaml.RNode) (*KptFile, error) {
	kpgfile := &KptFile{ResourceMeta: TypeMeta}
	s, err := node.String()
	if err != nil {
		return &KptFile{}, err
	}
	f := strings.NewReader(s)
	d := yaml.NewDecoder(f)
	d.KnownFields(true)
	if err = d.Decode(&kpgfile); err != nil {
		return &KptFile{}, fmt.Errorf("please make sure the package has a valid 'v1alpha1' Kptfile: %s", err)
	}
	return kpgfile, nil
}
