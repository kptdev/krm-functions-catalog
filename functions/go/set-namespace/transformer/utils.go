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
package set_namespace

import "github.com/kptdev/krm-functions-sdk/go/fn"

func NestedStringOrDie(o *fn.KubeObject, fields ...string) string {
	val, _, err := o.NestedString(fields...)
	if err != nil {
		panic(err)
	}
	return val
}

func SetNestedStringOrDie(o *fn.SubObject, value string, fields ...string) {
	err := o.SetNestedString(value, fields...)
	if err != nil {
		panic(err)
	}
}
