package main

import (
	"fmt"
	"regexp"

	"sigs.k8s.io/kustomize/api/filters/fieldspec"
	"sigs.k8s.io/kustomize/api/filters/filtersutil"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var _ kio.Filter = LocationFilter{}

// LocationFilter is used as Filter and Transformer
// As a Filter, it provides the Filter() function
// As a Transformer, it provides the Config function but no Transform function (redundant).
type LocationFilter struct {
	Region        string
	Zone          string
	RegionFsSlice []types.FieldSpec
	ZoneFsSlice   []types.FieldSpec
}

func (f LocationFilter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	_, err := kio.FilterAll(yaml.FilterFunc(
		func(node *yaml.RNode) (*yaml.RNode, error) {
			var fns []yaml.Filter
			for _, fs := range f.RegionFsSlice {
				fn := fieldspec.Filter{
					SetValue:  updateLocationFn(fs.RegexPattern, f.Region),
					FieldSpec: fs,
				}
				fns = append(fns, fn)
			}
			for _, fs := range f.ZoneFsSlice {
				fn := fieldspec.Filter{
					SetValue:  updateLocationFn(fs.RegexPattern, f.Zone),
					FieldSpec: fs,
				}
				fns = append(fns, fn)
			}
			return node.Pipe(fns...)
		})).Filter(nodes)
	return nodes, err
}

func updateLocationFn(regexPath, location string) filtersutil.SetFn {
	return func(node *yaml.RNode) (err error) {
		if regexPath == "" {
			return node.PipeE(updater{location: location})
		}
		defer func() {
			// recover from regex panic.
			if recover() != nil {
				err = fmt.Errorf("invalid regex pattern %v", regexPath)
			}
		}()
		re := regexp.MustCompile(regexPath)
		match := re.FindStringSubmatch(node.YNode().Value)
		namedGroup := make(map[string]string)
		for i, name := range re.SubexpNames() {
			if i != 0 && name != "" {
				namedGroup[name] = match[i]
			}
		}
		newLocation := ""
		if prefixStr, ok := namedGroup["prefix"]; ok {
			newLocation = newLocation + prefixStr
		}
		newLocation = newLocation + location
		if suffixStr, ok := namedGroup["suffix"]; ok {
			newLocation = newLocation + suffixStr
		}
		return node.PipeE(updater{location: newLocation})
	}
}

type updater struct {
	location string
}

func (u updater) Filter(rn *yaml.RNode) (*yaml.RNode, error) {
	return rn.Pipe(yaml.FieldSetter{StringValue: u.location})
}
