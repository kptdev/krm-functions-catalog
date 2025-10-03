package docs

import (
	"bytes"
	"fmt"
	"html/template"
	"sort"
	"strings"

	kptfilev1 "github.com/kptdev/kpt/pkg/api/kptfile/v1"
	"github.com/kptdev/krm-functions-catalog/functions/go/list-setters/listsetters"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// GenerateBlueprintReadme generates markdown readme and title if present for a blueprint
func GenerateBlueprintReadme(nodes []*yaml.RNode, repoPath, pkgName string) (string, string, error) {
	r, err := newBlueprintReadme(nodes, repoPath, pkgName)
	if err != nil {
		return "", "", err
	}
	// individual sections in the readme
	blueprintSections := []generateSection{
		generateDescriptionSection,
		generateSetterTableSection,
		generateSubPkgSection,
		generateResourceTableSection,
		generateResourceRefsSection,
		generateUsageSection,
	}
	for _, section := range blueprintSections {
		r.addSection(section)
	}
	// render readme
	err = r.render()
	if err != nil {
		return "", "", err
	}

	return getBlueprintTitle(r.bp.rootKf), r.string(), nil

}

// generateDescriptionSection generates the description
func generateDescriptionSection(r *blueprintReadme) error {
	// empty description
	if r.bp.rootKf.Info.Description == "" {
		return nil
	}
	// literal style description will include a newline
	if strings.HasSuffix(r.bp.rootKf.Info.Description, "\n") {
		r.write(r.bp.rootKf.Info.Description)
	} else {
		r.writeLn(r.bp.rootKf.Info.Description)
	}
	return nil
}

// generateSetterTableSection generates a markdown table of setters used in the package
func generateSetterTableSection(r *blueprintReadme) error {
	ls := listsetters.New()
	_, err := ls.Filter(r.bp.nodes)
	if err != nil {
		return err
	}
	r.write(getMdHeading("Setters", 2))
	setters := ls.GetResults()
	if len(setters) == 0 {
		r.writeLn("This package has no top-level setters. See sub-packages.")
		return nil
	} else {
		buf := &bytes.Buffer{}
		table := newMarkdownTable([]string{"Name", "Value", "Type", "Count"}, buf)
		for _, setter := range setters {
			table.Append([]string{setter.Name, setter.Value, setter.Type, fmt.Sprintf("%d", setter.Count)})
		}
		table.Render()
		r.write(buf.String())
		return nil
	}
}

// generateResourceTableSection generates subpkg section with links to subpkgs if any
func generateSubPkgSection(r *blueprintReadme) error {
	subPkgLinks := []string{}
	for path, pkg := range r.bp.kfs {
		// ignore rootpkg
		if path == kptfilev1.KptFileName {
			continue
		}
		// path of the form subpkg/Kptfile, we only require subpkg/
		pkgPath := strings.TrimSuffix(path, fmt.Sprintf("/%s", kptfilev1.KptFileName))
		subPkgLinks = append(subPkgLinks, getMdLink(pkg.Name, pkgPath))
	}

	r.write(getMdHeading("Sub-packages", 2))
	if len(subPkgLinks) == 0 {
		r.writeLn("This package has no sub-packages.")
	} else {
		sort.Strings(subPkgLinks)
		for _, link := range subPkgLinks {
			r.writeLn(getMdListItem(link))
		}
	}
	return nil
}

// generateResourceTableSection generates a markdown table of resources in the package
func generateResourceTableSection(r *blueprintReadme) error {
	buf := &bytes.Buffer{}
	table := newMarkdownTable([]string{"File", "APIVersion", "Kind", "Name", "Namespace"}, buf)
	for _, r := range r.filteredNodes {
		path, err := findResourcePath(r)
		if err != nil {
			return err
		}
		table.Append([]string{path, r.GetApiVersion(), r.GetKind(), r.GetName(), r.GetNamespace()})
	}
	r.write(getMdHeading("Resources", 2))
	if len(r.filteredNodes) == 0 {
		r.writeLn("This package has no top-level resources. See sub-packages.")
	} else {
		table.Render()
		r.write(buf.String())
	}
	return nil
}

// generateResourceRefsSection generates resource references with links to c.g.c or k8s docs
func generateResourceRefsSection(r *blueprintReadme) error {
	resourcesLinks := []string{}
	gvkSet := map[resid.Gvk]bool{}

	// dedupe multiple resources of same gvk
	for _, item := range r.filteredNodes {
		r := resid.GvkFromNode(item)
		_, exists := gvkSet[r]
		if !exists {
			gvkSet[r] = true
		}
	}

	// generate links for each gvk, if no link document kind
	for r := range gvkSet {
		link := getResourceDocsLink(r)
		if link != "" {
			resourcesLinks = append(resourcesLinks, link)
		} else {
			resourcesLinks = append(resourcesLinks, r.Kind)
		}

	}

	r.write(getMdHeading("Resource References", 2))
	if len(resourcesLinks) == 0 {
		r.writeLn("This package has no top-level resources. See sub-packages.")
	} else {
		sort.Strings(resourcesLinks)
		for _, l := range resourcesLinks {
			r.writeLn(getMdListItem(l))
		}
	}
	return nil

}

// generateUsageSection generates usage section describing how to use the package
func generateUsageSection(r *blueprintReadme) error {
	tmpl := strings.NewReplacer("¬", "`").Replace(`1.  Clone the package:
    ¬¬¬shell
    kpt pkg get {{.RepoPath}}{{.Pkgname}}@${VERSION}
    ¬¬¬
    Replace ¬${VERSION}¬ with the desired repo branch or tag
    (for example, ¬main¬).

1.  Move into the local package:
    ¬¬¬shell
    cd "./{{.Pkgname}}/"
    ¬¬¬

1.  Edit the function config file(s):
    - setters.yaml

1.  Execute the function pipeline
    ¬¬¬shell
    kpt fn render
    ¬¬¬

1.  Initialize the resource inventory
    ¬¬¬shell
    kpt live init --namespace ${NAMESPACE}
    ¬¬¬
    Replace ¬${NAMESPACE}¬ with the namespace in which to manage
    the inventory ResourceGroup (for example, ¬config-control¬).

1.  Apply the package resources to your cluster
    ¬¬¬shell
    kpt live apply
    ¬¬¬

1.  Wait for the resources to be ready
    ¬¬¬shell
    kpt live status --output table --poll-until current
    ¬¬¬`)
	t, err := template.New("usage").Parse(tmpl)
	if err != nil {
		return err
	}
	r.write(getMdHeading("Usage", 2))

	err = t.Execute(r.content, struct {
		Pkgname  string
		RepoPath string
	}{
		Pkgname:  r.bp.pkgName,
		RepoPath: r.bp.repoPath,
	},
	)
	if err != nil {
		return err
	}
	return nil
}
