package kpt

import (
	"fmt"

	"github.com/kptdev/krm-functions-sdk/go/fn"
)

type PackageContext struct {
	Name       string
	ParentPath string
}

func FindPackageContext(objects fn.KubeObjects) (*PackageContext, error) {
	matches := objects.Where(fn.IsLocalConfig).Where(fn.IsName("kptfile.kpt.dev"))
	if len(matches) == 0 {
		return nil, fmt.Errorf("unable to find package context object")
	}
	if len(matches) != 1 {
		return nil, fmt.Errorf("found multiple matches for package-context object")
	}
	packageContext := matches[0]

	packageName, _, err := packageContext.NestedString("data", "name")
	if err != nil {
		return nil, fmt.Errorf("error getting data.name from package context object: %w", err)
	}
	if packageName == "" {
		return nil, fmt.Errorf("package name (data.name) not set in package context object")
	}
	c := &PackageContext{
		Name: packageName,
	}

	c.ParentPath = packageContext.GetAnnotation("config.kubernetes.io/parent-path")

	return c, nil

}
