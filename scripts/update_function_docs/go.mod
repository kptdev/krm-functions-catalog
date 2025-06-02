module github.com/kptdev/krm-functions-catalog/scripts/update_function_docs

go 1.24.3

require (
	github.com/GoogleContainerTools/kpt-functions-catalog/scripts/patch_reader v0.0.0
	gopkg.in/yaml.v2 v2.4.0
)

require golang.org/x/mod v0.4.1 // indirect

replace github.com/GoogleContainerTools/kpt-functions-catalog/scripts/patch_reader v0.0.0 => ../patch_reader
