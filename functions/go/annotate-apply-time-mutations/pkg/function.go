// Copyright 2021-2026 The kpt Authors
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

package pkg

import (
	"fmt"
	"os"
	"strconv"

	"github.com/kptdev/kpt/api/kubeobject"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/cli-utils/pkg/object/mutation"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func Process(resourceList *kubeobject.ResourceList) (bool, error) {
	targetMutations := map[mutation.ResourceReference]mutation.ApplyTimeMutation{}

	// scan all the items for ApplyTimeMutation objects
	fmt.Fprintln(os.Stderr, "Scanning items for ApplyTimeMutation objects")
	err := visitItems(resourceList, func(itemIndex int, objMeta yaml.ResourceIdentifier, objFile kubeobject.File, obj *yaml.RNode) (kubeobject.Results, bool, error) {
		fmt.Fprintf(os.Stderr, "Scanning (Item: %d, Object: %+v, File: %+v)\n", itemIndex, objMeta, objFile)
		var fnResults kubeobject.Results
		scanner := ObjectScanner{}
		atm, err := scanner.Scan(obj)
		if err != nil {
			fnResults = append(fnResults, &kubeobject.Result{
				Severity:    kubeobject.Error,
				Message:     fmt.Sprintf("Error scanning object (item: %d): %v", itemIndex, err),
				ResourceRef: &objMeta,
				File:        &objFile,
			})
			// not fatal, do not return error
			return fnResults, false, nil
		}
		if atm == nil {
			// no ApplyTimeMutation object found
			return fnResults, false, nil
		}
		fmt.Fprintf(os.Stderr, "Found valid ApplyTimeMutation object (item: %d)\n", itemIndex)
		fnResults = append(fnResults, &kubeobject.Result{
			Severity:    kubeobject.Info,
			Message:     fmt.Sprintf("Found valid ApplyTimeMutation object (item: %d)", itemIndex),
			ResourceRef: &objMeta,
			File:        &objFile,
		})
		targetMutations[atm.Spec.TargetRef] = atm.Spec.Substitutions
		return fnResults, false, nil
	})
	if err != nil {
		// fatal error
		return false, err
	}

	// Scan for objects targetted by ApplyTimeMutation objects
	fmt.Fprintln(os.Stderr, "Scanning items for mutation targets")
	err = visitItems(resourceList, func(itemIndex int, objMeta yaml.ResourceIdentifier, objFile kubeobject.File, obj *yaml.RNode) (kubeobject.Results, bool, error) {
		fmt.Fprintf(os.Stderr, "Scanning (Item: %d, Object: %+v, File: %+v)\n", itemIndex, objMeta, objFile)
		var fnResults kubeobject.Results

		atm, found := targetMutations[refWithVersion(objMeta)]
		if !found {
			atm, found = targetMutations[refWithoutVersion(objMeta)]
			if !found {
				// no match found
				// not fatal, do not return error
				return fnResults, false, nil
			}
		}

		fmt.Fprintf(os.Stderr, "Writing annotation (Item: %d, Object: %+v, File: %+v)\n", itemIndex, objMeta, objFile)
		err = WriteAnnotation(obj, atm)
		if err != nil {
			fnResults = append(fnResults, &kubeobject.Result{
				Severity:    kubeobject.Error,
				Message:     fmt.Sprintf("Error writing annotation (item: %d): %v", itemIndex, err),
				ResourceRef: &objMeta,
				File:        &objFile,
			})
			// not fatal, do not return error
			return fnResults, false, nil
		}
		fnResults = append(fnResults, &kubeobject.Result{
			Severity:    kubeobject.Info,
			Message:     fmt.Sprintf("Wrote apply-time-mutation annotation (item: %d)", itemIndex),
			ResourceRef: &objMeta,
			File:        &objFile,
		})
		return fnResults, true, nil
	})
	if err != nil {
		// fatal error
		return false, err
	}

	// Scan all the items for apply-time-mutation comments
	fmt.Fprintln(os.Stderr, "Scanning items for apply-time-mutation comments")
	err = visitItems(resourceList, func(itemIndex int, objMeta yaml.ResourceIdentifier, objFile kubeobject.File, obj *yaml.RNode) (kubeobject.Results, bool, error) {
		fmt.Fprintf(os.Stderr, "Scanning (Item: %d, Object: %+v, File: %+v)\n", itemIndex, objMeta, objFile)
		var fnResults kubeobject.Results
		scanner := CommentScanner{
			ObjMeta: objMeta,
			ObjFile: objFile,
		}
		scanResults, err := scanner.Scan(obj)
		if err != nil {
			fnResults = append(fnResults, &kubeobject.Result{
				Severity:    kubeobject.Error,
				Message:     fmt.Sprintf("Error scanning object (item: %d): %v", itemIndex, err),
				ResourceRef: &objMeta,
				File:        &objFile,
			})
			// not fatal, do not return error
			return fnResults, false, nil
		}
		if len(scanResults) == 0 {
			// no apply-time-mutation comment found
			return fnResults, false, nil
		}
		subs := make(mutation.ApplyTimeMutation, 0, len(fnResults))
		for _, scanResult := range scanResults {
			fmt.Fprintf(os.Stderr, "Found valid apply-time-mutation comment (item: %d, field: %q): %s\n", itemIndex, scanResult.Path, scanResult.Comment)
			subs = append(subs, scanResult.Substitution)
			fnResults = append(fnResults, &kubeobject.Result{
				Severity:    kubeobject.Info,
				Message:     fmt.Sprintf("Found valid apply-time-mutation comment (item: %d): %s", itemIndex, scanResult.Comment),
				ResourceRef: &objMeta,
				File:        &objFile,
				Field: &kubeobject.Field{
					Path:         scanResult.Path,
					CurrentValue: fmt.Sprintf("%v", scanResult.Value),
				},
			})
		}
		fmt.Fprintf(os.Stderr, "Writing annotation (Item: %d, Object: %+v, File: %+v)\n", itemIndex, objMeta, objFile)
		err = WriteAnnotation(obj, subs)
		if err != nil {
			fnResults = append(fnResults, &kubeobject.Result{
				Severity:    kubeobject.Error,
				Message:     fmt.Sprintf("Error writing annotation (item: %d): %v", itemIndex, err),
				ResourceRef: &objMeta,
				File:        &objFile,
			})
			// not fatal, do not return error
			return fnResults, false, nil
		}
		fnResults = append(fnResults, &kubeobject.Result{
			Severity:    kubeobject.Info,
			Message:     fmt.Sprintf("Wrote apply-time-mutation annotation (item: %d)", itemIndex),
			ResourceRef: &objMeta,
			File:        &objFile,
		})
		return fnResults, true, nil
	})
	if err != nil {
		// fatal error
		return false, err
	}

	return true, nil
}

func visitItems(resourceList *kubeobject.ResourceList, fn func(int, yaml.ResourceIdentifier, kubeobject.File, *yaml.RNode) (kubeobject.Results, bool, error)) error {
	for i, nodeObj := range resourceList.Items {
		node := nodeObj.CopyToResourceNode()
		filePath, fileIndexStr, err := kioutil.GetFileAnnotations(node)
		if err != nil {
			return fmt.Errorf("item %d has invalid file annotations: %w", i, err)
		}
		meta, err := node.GetMeta()
		if err != nil {
			return fmt.Errorf("item %d has invalid metadata: %w", i, err)
		}
		fileIndex, err := strconv.Atoi(fileIndexStr)
		if err != nil {
			return fmt.Errorf("item %d has invalid file index: %q", i, fileIndexStr)
		}
		objMeta := meta.GetIdentifier()
		objFile := kubeobject.File{
			Path:  filePath,
			Index: fileIndex,
		}
		itemResults, modified, err := fn(i, objMeta, objFile, node)
		resourceList.Results = append(resourceList.Results, itemResults...)
		if err != nil {
			// fatal error
			return err
		}
		if modified {
			resourceList.Items[i] = kubeobject.NewKubeObjectFromResourceNode(node)
		}
	}
	return nil
}

func refWithVersion(objMeta yaml.ResourceIdentifier) mutation.ResourceReference {
	return mutation.ResourceReference{
		APIVersion: objMeta.APIVersion,
		Kind:       objMeta.Kind,
		Name:       objMeta.Name,
		Namespace:  objMeta.Namespace,
	}
}

func refWithoutVersion(objMeta yaml.ResourceIdentifier) mutation.ResourceReference {
	return mutation.ResourceReference{
		Group:     schema.FromAPIVersionAndKind(objMeta.APIVersion, objMeta.Kind).Group,
		Kind:      objMeta.Kind,
		Name:      objMeta.Name,
		Namespace: objMeta.Namespace,
	}
}
