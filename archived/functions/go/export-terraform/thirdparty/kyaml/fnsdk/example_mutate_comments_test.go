package fnsdk_test

import (
	"os"
	"strings"

	"github.com/kptdev/krm-functions-catalog/archived/functions/go/export-terraform/thirdparty/kyaml/fnsdk"
)

// In this example, we mutate line comments for field metadata.name.
// Some function may want to store some information in the comments (e.g.
// apply-setters function: https://catalog.kpt.dev/apply-setters/v0.2/)

func Example_dMutateComments() {
	if err := fnsdk.AsMain(fnsdk.ResourceListProcessorFunc(mutateComments)); err != nil {
		os.Exit(1)
	}
}

func mutateComments(rl *fnsdk.ResourceList) error {
	for i := range rl.Items {
		lineComment, found, err := rl.Items[i].LineComment("metadata", "name")
		if err != nil {
			return err
		}
		if !found {
			return nil
		}

		if strings.TrimSpace(lineComment) == "" {
			lineComment = "bar-system"
		} else {
			lineComment = strings.Replace(lineComment, "foo", "bar", -1)
		}
		if err = rl.Items[i].SetLineComment(lineComment, "metadata", "name"); err != nil {
			return err
		}
	}
	return nil
}
