package docs

import (
	"fmt"
	"path"
	"strings"

	kptfilev1 "github.com/kptdev/kpt/pkg/api/kptfile/v1"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// blueprint represents a kpt pkg with a root kptfile, resources and any additional subpackages.
type blueprint struct {
	rootKf   *kptfilev1.KptFile
	pkgName  string
	kfs      map[string]*kptfilev1.KptFile
	nodes    []*yaml.RNode
	repoPath string
}

// blueprintReadme represents a markdown readme for a blueprint
type blueprintReadme struct {
	content       *strings.Builder
	bp            blueprint
	filteredNodes []*yaml.RNode
	generators    []generateSection
}

// generateSection is a function that adds a readme section to a blueprint readme
type generateSection func(*blueprintReadme) error

// newBlueprintReadme initializes a blueprint readme
func newBlueprintReadme(n []*yaml.RNode, repoPath, pkgName string) (blueprintReadme, error) {
	// deep copy resources to prevent any changes to resources
	nodes := []*yaml.RNode{}
	for _, r := range n {
		nodes = append(nodes, r.Copy())
	}
	// find all packages
	pkgs, err := findPkgs(nodes)
	if err != nil {
		return blueprintReadme{}, err
	}
	// rootKF must be present
	rootKf, hasRootKf := pkgs[kptfilev1.KptFileName]
	if !hasRootKf {
		return blueprintReadme{}, fmt.Errorf("unable to find root Kptfile, please include --include-meta-resources flag if a Kptfile is present")
	}
	// specific files we want to omit from readme including Kptfile, subpkgs and any fn configs
	skipFiles := map[string]bool{kptfilev1.KptFileName: true}
	for _, fnCfg := range getFnCfgPaths(rootKf) {
		skipFiles[fnCfg] = true
	}
	for pkgPath := range pkgs {
		if pkgPath != kptfilev1.KptFileName {
			skipFiles[path.Dir(pkgPath)] = true
		}
	}
	// if no explicit pkg name, use kf pkgname
	if pkgName == "" {
		pkgName = rootKf.Name
	}
	b := blueprint{rootKf: rootKf, kfs: pkgs, nodes: nodes, repoPath: repoPath, pkgName: pkgName}
	return blueprintReadme{content: &strings.Builder{}, bp: b, filteredNodes: filterResources(nodes, skipFiles)}, nil
}

func (r *blueprintReadme) write(d string) {
	r.content.WriteString(d)
}

func (r *blueprintReadme) writeLn(d string) {
	r.write(fmt.Sprintf("%s\n", d))
}

func (r *blueprintReadme) render() error {
	for _, generator := range r.generators {
		err := generator(r)
		if err != nil {
			return err
		}
		r.writeLn("")
	}
	return nil
}

func (r *blueprintReadme) addSection(g generateSection) {
	r.generators = append(r.generators, g)
}

func (r *blueprintReadme) string() string {
	return r.content.String()
}
