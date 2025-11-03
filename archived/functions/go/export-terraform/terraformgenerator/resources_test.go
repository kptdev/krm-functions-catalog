package terraformgenerator

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/kptdev/krm-functions-catalog/archived/functions/go/export-terraform/thirdparty/kyaml/fnsdk"
)

func TestGetParentRef(t *testing.T) {
	require := require.New(t)
	contents, err := os.ReadFile(path.Join("../testdata/iam/input/folder_external.yaml"))
	require.NoErrorf(err, "Loads test content")

	item, err := sdk.ParseKubeObject(contents)
	require.NoErrorf(err, "Parses test content")

	resource := &terraformResource{
		Item: item,
	}

	kind, name, err := resource.getParentRef()
	require.NoErrorf(err, "Finds parent resource")
	require.Equalf("Folder", kind, "kind to match")
	require.Equalf("335620346181", name, "name to match")
}
