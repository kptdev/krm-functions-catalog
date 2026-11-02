package main

import (
	"fmt"
	"os"
	"strings"

	fnresult "github.com/kptdev/kpt/api/fnresult/v1"
	"github.com/kptdev/kpt/api/kubeobject"
	"github.com/kptdev/krm-functions-sdk/go/fn"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
)

const (
	annotationKeysLiteral = "annotationKeys"
	annotationDelimeter   = ","
)

// nolint
func main() {
	if err := fn.AsMain(kubeobject.ResourceListProcessorFunc(deleteAnnotations)); err != nil {
		os.Exit(1)
	}
}

func deleteAnnotations(rl *kubeobject.ResourceList) (bool, error) {

	var annotationKeys []string
	annotationKeys, err := getAnnotationKeys(rl.FunctionConfig)

	if err != nil {
		rl.Results = append(rl.Results, kubeobject.ErrorResult(err))
		return false, err
	}

	items, err := processResources(rl.Items, annotationKeys)
	if err != nil {
		rl.Results = append(rl.Results, kubeobject.ErrorResult(err))
		return false, err
	}

	for _, resultItem := range items {
		rl.Results = append(rl.Results, resultItem)
	}

	return true, nil
}

func processResources(objects []*kubeobject.KubeObject, annotationKeys []string) ([]*fnresult.ResultItem, error) {
	var resultItems []*fnresult.ResultItem
	var currObjName string
	var prunedObjCtr int
	for _, o := range objects {
		if o.GetKind() == "" || o.GetName() == "" || o.GetAPIVersion() == "" {
			continue
		}

		for _, annotationKey := range annotationKeys {
			removed, err := o.RemoveNestedField("metadata", "annotations", annotationKey)
			if err != nil {
				return nil, err
			}

			if removed {
				if o.GetName() != currObjName {
					prunedObjCtr++
					currObjName = o.GetName()
				}

				itemFilePath := o.GetAnnotations()["internal.config.kubernetes.io/path"]

				resultItems = append(resultItems, &fnresult.ResultItem{
					Message: fmt.Sprintf("Annonation: [%s] removed from resource: [%s]", annotationKey, o.GetName()),
					File: &framework.File{
						Path: itemFilePath,
					},
					Severity: framework.Info,
				})
			}
		}
	}

	if len(resultItems) > 0 {
		infoResultSlice := []*fnresult.ResultItem{}
		infoResultSlice = append(infoResultSlice, &fnresult.ResultItem{
			Severity: framework.Info,
			Message:  fmt.Sprintf("The provided annotations were deleted from %d resource(s)", prunedObjCtr),
		})

		resultItems = append(infoResultSlice, resultItems...)
	} else if len(resultItems) == 0 {
		resultItems = append(resultItems, &fnresult.ResultItem{
			Message:  "None of the resources had the provided annotations to delete",
			Severity: framework.Info,
		})
	}

	return resultItems, nil
}

// getAnnotationKeys gets the keys to delete from resources from the functionConfig
func getAnnotationKeys(fc *kubeobject.KubeObject) ([]string, error) {
	annotationKeysString, found, err := fc.NestedString("data", annotationKeysLiteral)

	if !found || annotationKeysString == "" {
		return nil, fmt.Errorf("%s was not provided as part of the config or paramters to the function, %w", annotationKeysLiteral, err)
	}

	annotationKeys := strings.Split(strings.TrimSpace(annotationKeysString), annotationDelimeter)
	return annotationKeys, nil
}
