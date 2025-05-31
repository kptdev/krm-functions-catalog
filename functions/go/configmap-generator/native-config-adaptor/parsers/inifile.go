package parsers

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/configmap-generator/fn"
	"gopkg.in/ini.v1"
)

const (
	ConfigMapPrefix = "cm."
)

func ToIniFile(canonicalObject *fn.KubeObject, fileRef, fileName string, IsFlatten bool) (*fn.KubeObject, error) {
	if !IsFlatten {
		return nil, fmt.Errorf("not implemented")
	}
	cfg := ini.Empty(ini.LoadOptions{
		SkipUnrecognizableLines: true,
		AllowBooleanKeys:        true,
		IgnoreInlineComment:     false,
	})
	for key, value := range canonicalObject.NestedStringMapOrDie("data") {
		prefixSectionKey := strings.Split(key, ".")
		var createSectionErr error
		section, err := cfg.GetSection(prefixSectionKey[1])
		if err != nil {
			section, createSectionErr = cfg.NewSection(prefixSectionKey[1])
			if createSectionErr != nil {
				return nil, createSectionErr
			}
		}
		section.NewKey(prefixSectionKey[2], value)
	}
	out := &bytes.Buffer{}
	ini.PrettyFormat = false
	cfg.WriteTo(out)
	newNonKrmObject := fn.NewNonKrmResource()
	p, e := fn.NewFromTypedObject(newNonKrmObject)
	p.SetName(fileRef)
	p.SetNestedStringOrDie(fileName, "spec", "filename")
	p.SetNestedStringOrDie(out.String(), "spec", "content")
	return p, e
}

// Read IniFile from content, write to Object
func FromIniFile(object *fn.KubeObject, content string, ShouldFlatten bool) error {
	cfg, err := ini.LoadSources(ini.LoadOptions{SkipUnrecognizableLines: true}, []byte(content))
	if err != nil {
		return err
	}
	if ShouldFlatten {
		data := map[string]string{}
		for _, section := range cfg.Sections() {
			for _, key := range section.Keys() {
				cmKey := ConfigMapPrefix + section.Name() + "." + key.Name()
				data[cmKey] = key.Value()
			}
		}
		if err = object.SetNestedStringMap(data, "data"); err != nil {
			return err
		}
		return nil
	}

	spec := object.UpsertMap("spec")
	for _, section := range cfg.Sections() {
		subObject := spec.UpsertMap(section.Name())
		for _, key := range section.Keys() {
			if err = subObject.SetNestedString(key.Value(), key.Name()); err != nil {
				return err
			}
		}
	}
	return nil
}
