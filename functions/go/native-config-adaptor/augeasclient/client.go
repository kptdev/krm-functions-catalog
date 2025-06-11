package augeasclient

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"honnef.co/go/augeas"
)

type FileFormat string

const (
	IniFile     FileFormat = "IniFile"
	Erlang                 = "Erlang"
	Json                   = "Json"
	Puppetfile             = "Puppetfile"
	Group                  = "Group"
	Known_Hosts            = "Known_Hosts"
)
const AugeasFilePrefix = "/files"
const ConfigMapPrefixMarker = "cm"

// http://augeas.net/stock_lenses.html
var AugeasLenses = map[FileFormat]bool{
	IniFile:     false,
	Erlang:      false,
	Json:        false,
	Puppetfile:  false,
	Group:       true,
	Known_Hosts: true,
}

type AugeasConfigSpec struct {
	Source []*AugeasConfigSource `json:"source,omitempty" yaml:"source,omitempty"`
}

type AugeasConfigSource struct {
	LocalFileRef string `json:"localFileRef,omitempty" yaml:"localFileRef,omitempty"`
	LocalFile    string `json:"localFile,omitempty" yaml:"localFile,omitempty"`
	Format       string `json:"format,omitempty" yaml:"format,omitempty"`
	AsConfigMap  bool   `json:"asConfigMap,omitempty" yaml:"asConfigMap,omitempty"`
}

func Generate(name string, spec *AugeasConfigSpec) (fn.KubeObjects, error) {
	ag, err := augeas.New("/", "", augeas.None)
	if err != nil {
		return nil, err
	}
	defer ag.Close()

	var newObjects []*fn.KubeObject
	for _, source := range spec.Source {
		autoload, ok := AugeasLenses[FileFormat(source.Format)]
		if !ok {
			return nil, fmt.Errorf("unknown lenses format %v", source.Format)
		}
		// TODO: use ag.transform
		if !autoload {
			lensPath := "/augeas/load/" + source.Format + "/lens"
			lens := source.Format + ".lns"
			if err = ag.Set(lensPath, lens); err != nil {
				return nil, err
			}
		}
		lensInclude := "/augeas/load/" + source.Format + "/incl"
		ag.Set(lensInclude, filepath.Join(AugeasFilePrefix, source.Format))
		ag.Load()

		object := CreateCanonicalObject(name, source)
		if source.AsConfigMap {
			WalkAugeasAndBuildFlattenObject(ag, object.UpsertMap("data"), filepath.Join(AugeasFilePrefix, source.LocalFile), ConfigMapPrefixMarker)
		} else {
			WalkAugeasAndBuildStructuredObject(ag, object.UpsertMap("spec"), filepath.Join(AugeasFilePrefix, source.LocalFile))
		}
		newObjects = append(newObjects, object)
		cmObject := StoreRawDataInConfigMap(name, source.LocalFile)
		newObjects = append(newObjects, cmObject)
	}
	return newObjects, nil
}

func StoreRawDataInConfigMap(name, fPath string) *fn.KubeObject {
	object := fn.NewEmptyKubeObject()
	object.SetKind("ConfigMap")
	object.SetAPIVersion("v1")
	object.SetName(name)
	content, _ := ioutil.ReadFile(fPath)
	data := map[string]string{
		filepath.Base(fPath): string(content),
	}
	object.SetNestedStringMap(data, "data")
	return object
}

func CreateCanonicalObject(name string, source *AugeasConfigSource) *fn.KubeObject {
	lenseNameSlitted := strings.Split(source.Format, "_")
	camelcaseLense := ""
	for _, segment := range lenseNameSlitted {
		camelcaseLense += strings.ToUpper(string(segment[0])) + segment[1:]
	}

	object := fn.NewEmptyKubeObject()

	if source.AsConfigMap {
		object.SetKind("ConfigMap")
		object.SetAPIVersion("v1")
		object.SetName(name + "-internal")
		object.SetAnnotation(fn.KptLocalConfig, "true")
	} else {
		object.SetKind(camelcaseLense)
		object.SetAPIVersion("config.kpt.dev/v1alpha1")
		object.SetName(name + "-internal")
		object.SetAnnotation(fn.KptLocalConfig, "true")
	}
	return object
}

func WalkAugeasAndBuildFlattenObject(ag augeas.Augeas, object *fn.SubObject, path string, prefix string) error {
	// Augueas index starts with 1.
	// slice objects
	_, err := ag.Get(filepath.Join(path, "*[1]"))
	if err != nil {
		value, e := ag.Get(path)
		if e != nil {
			return e
		}
		return object.SetNestedString(value, prefix+"."+filepath.Base(path))
	}
	for i := 1; ; i++ {
		branchPath := filepath.Join(path, fmt.Sprintf("*[%d]", i))
		branchNode, err := ag.Get(branchPath)
		if err != nil {
			return err
		}

		subNodes, err := ag.Match(filepath.Join(branchPath, "*"))
		if err != nil {
			return err
		}
		for _, subNode := range subNodes {
			WalkAugeasAndBuildFlattenObject(ag, object, subNode, prefix+"."+branchNode)
		}
	}
	return nil
}

func WalkAugeasAndBuildStructuredObject(ag augeas.Augeas, object *fn.SubObject, path string) error {
	// Augueas index starts with 1.

	// Walk Slice objects
	_, err := ag.Get(filepath.Join(path, "*[1]"))
	if err == nil {
		for i := 1; ; i++ {
			branchPath := filepath.Join(path, fmt.Sprintf("*[%d]", i))
			branchNode, err := ag.Get(branchPath)
			if err != nil {
				return err
			}

			subNodes, err := ag.Match(filepath.Join(branchPath, "*"))
			if err != nil {
				return err
			}
			for _, subNode := range subNodes {
				subObject := object.UpsertMap(branchNode)
				WalkAugeasAndBuildStructuredObject(ag, subObject, subNode)
			}
		}
	} else {
		// Walk leaf object. this is adjusted and simplified for INI file
		val, e := ag.Get(path)
		if e != nil {
			return nil
		}
		object.SetNestedString(val, filepath.Base(path))
	}
	return nil
}
