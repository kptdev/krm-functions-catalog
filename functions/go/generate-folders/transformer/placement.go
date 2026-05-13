// Copyright 2026 The kpt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package generate_folders

import (
	"strconv"

	"github.com/kptdev/krm-functions-sdk/go/fn"
)

func newSourcePlacement(obj *fn.KubeObject) *sourcePlacement {
	path := obj.PathAnnotation()
	if path == "" {
		return nil
	}

	nextIndex := obj.IndexAnnotation()
	if nextIndex < 0 {
		nextIndex = 0
	} else {
		nextIndex++
	}

	return &sourcePlacement{
		path:      path,
		nextIndex: nextIndex,
	}
}

func applySourcePlacement(folderObj *fn.KubeObject, placement *sourcePlacement) {
	if placement == nil || placement.path == "" {
		return
	}

	index := strconv.Itoa(placement.nextIndex)
	placement.nextIndex++

	_ = folderObj.SetAnnotation(internalPathAnnotation, placement.path)
	_ = folderObj.SetAnnotation(configPathAnnotation, placement.path)
	_ = folderObj.SetAnnotation(internalIndexAnnotation, index)
	_ = folderObj.SetAnnotation(configIndexAnnotation, index)
}
