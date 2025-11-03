package testutil

import (
	"io/ioutil"
	"testing"

	"github.com/kptdev/krm-functions-catalog/archived/functions/go/export-terraform/thirdparty/kyaml/fnsdk"
	"github.com/stretchr/testify/assert"
)

func TestResourceListFromFile(t *testing.T) {
	rl, err := ResourceListFromFile("testdata/resourcelist.yaml")
	assert.NoError(t, err)

	err = setAnnotationFn(rl)
	assert.NoError(t, err)

	actual, err := rl.ToYAML()
	assert.NoError(t, err)

	rlExpected, err := ResourceListFromFile("testdata/expected_resourcelist.yaml")
	assert.NoError(t, err)
	expected, err := rlExpected.ToYAML()
	assert.Equal(t, string(expected), string(actual))
}

func TestResourceListFromToDirectory(t *testing.T) {
	rl, err := ResourceListFromDirectory("testdata/resources/input", "testdata/resources/input/fn-config.yaml")
	assert.NoError(t, err)

	err = setAnnotationFn(rl)
	assert.NoError(t, err)

	tempDir, err := ioutil.TempDir("", "kyaml-fn-sdk-*")
	assert.NoError(t, err)

	err = ResourceListToDirectory(rl, tempDir)
	assert.NoError(t, err)

	rlActual, err := ResourceListFromDirectory(tempDir, "")
	assert.NoError(t, err)
	actual, err := rlActual.ToYAML()

	rlExpected, err := ResourceListFromDirectory("testdata/resources/expected", "")
	assert.NoError(t, err)
	expected, err := rlExpected.ToYAML()
	assert.Equal(t, string(expected), string(actual))
}

func setAnnotationFn(rl *fnsdk.ResourceList) error {
	for _, item := range rl.Items {
		item.SetAnnotation("foo", "bar")
	}
	return nil
}
