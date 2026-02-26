// Copyright 2026 The kpt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package starlark

import (
	"encoding/json"
	"os"
	"strings"
	"sync"

	"github.com/qri-io/starlib/util"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type Context struct {
	resourceList starlark.Value
}

func (c *Context) predeclared() (starlark.StringDict, error) {
	e, err := env()
	if err != nil {
		return nil, err
	}
	dict := starlark.StringDict{
		"resource_list": c.resourceList,
		"open_api":      &LazyInitializationOpenapi{},
		"environment":   e,
	}

	return starlark.StringDict{
		"ctx": starlarkstruct.FromStringDict(starlarkstruct.Default, dict),
	}, nil
}

func oa() (starlark.Value, error) {
	return interfaceToValue(openapi.Schema())
}

func env() (starlark.Value, error) {
	env := map[string]interface{}{}
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) < 2 {
			continue
		}
		env[pair[0]] = pair[1]
	}
	value, err := util.Marshal(env)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return value, nil
}

type LazyInitializationOpenapi struct {
	once sync.Once
	val  starlark.Value
}

var _ starlark.Mapping = &LazyInitializationOpenapi{}

func (v *LazyInitializationOpenapi) init() {
	o, err := oa()
	if err != nil {
		panic(err)
	}
	v.val = o
}

func (v *LazyInitializationOpenapi) String() string {
	v.once.Do(v.init)
	return v.val.String()
}

func (v *LazyInitializationOpenapi) Type() string {
	v.once.Do(v.init)
	return v.val.Type()
}

func (v *LazyInitializationOpenapi) Freeze() {
	v.once.Do(v.init)
	v.val.Freeze()
}

func (v *LazyInitializationOpenapi) Truth() starlark.Bool {
	v.once.Do(v.init)
	return v.val.Truth()
}

func (v *LazyInitializationOpenapi) Hash() (uint32, error) {
	v.once.Do(v.init)
	return v.val.Hash()
}

func (v *LazyInitializationOpenapi) Get(val starlark.Value) (starlark.Value, bool, error) {
	v.once.Do(v.init)
	m, ok := v.val.(starlark.Mapping)
	if ok {
		return m.Get(val)
	}
	return nil, false, nil
}

func interfaceToValue(i interface{}) (starlark.Value, error) {
	b, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	var in map[string]interface{}
	if err := yaml.Unmarshal(b, &in); err != nil {
		return nil, errors.Wrap(err)
	}

	value, err := util.Marshal(in)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return value, nil
}
