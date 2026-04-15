package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/kptdev/krm-functions-sdk/go/fn"
)

const (
	annotationKeysLiteral = "annotationKeys"
	annotationDelimeter   = ","
)

// nolint
func main() {
	if err := fn.AsMain(fn.ResourceListProcessorFunc(deleteAnnotations)); err != nil {
		os.Exit(1)
	}
}

func deleteAnnotations(rl *fn.ResourceList) (bool, error) {

	var annotationKeys []string
	annotationKeys, err := getAnnotationKeys(rl.FunctionConfig)

	if err != nil {
		rl.Results = append(rl.Results, fn.ErrorResult(err))
		return false, err
	}

	items, err := processResources(rl.Items, annotationKeys)
	if err != nil {
		rl.Results = append(rl.Results, fn.ErrorResult(err))
		return false, err
	}

	for _, resultItem := range items {
		rl.Results = append(rl.Results, resultItem)
	}

	return true, nil
}

func processResources(objects []*fn.KubeObject, annotationKeys []string) ([]*fn.Result, error) {
	var resultItems []*fn.Result
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

				resultItems = append(resultItems, &fn.Result{
					Message: fmt.Sprintf("Annonation: [%s] removed from resource: [%s]", annotationKey, o.GetName()),
					File: &fn.File{
						Path: itemFilePath,
					},
					Severity: fn.Info,
				})
			}
		}
	}

	if len(resultItems) > 0 {
		infoResultSlice := []*fn.Result{}
		infoResultSlice = append(infoResultSlice, &fn.Result{
			Severity: fn.Info,
			Message:  fmt.Sprintf("The provided annotations were deleted from %d resource(s)", prunedObjCtr),
		})

		resultItems = append(infoResultSlice, resultItems...)
	} else if len(resultItems) == 0 {
		resultItems = append(resultItems, &fn.Result{
			Message:  "None of the resources had the provided annotations to delete",
			Severity: fn.Info,
		})
	}

	return resultItems, nil
}

// getAnnotationKeys gets the keys to delete from resources from the functionConfig
func getAnnotationKeys(fc *fn.KubeObject) ([]string, error) {
	annotationKeysString, found, err := fc.NestedString("data", annotationKeysLiteral)

	if !found || annotationKeysString == "" {
		return nil, fmt.Errorf("%s was not provided as part of the config or paramters to the function, %w", annotationKeysLiteral, err)
	}

	annotationKeys := strings.Split(strings.TrimSpace(annotationKeysString), annotationDelimeter)
	return annotationKeys, nil
}
