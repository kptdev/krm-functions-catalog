module github.com/kptdev/krm-functions-catalog/functions/go/generate-kpt-pkg-docs

go 1.24.3

require (
	github.com/GoogleContainerTools/kpt-functions-sdk/go v0.0.0-20211015223526-2e6d9522928b
	github.com/kptdev/krm-functions-catalog/functions/go/list-setters v0.0.0-00010101000000-000000000000
	github.com/olekukonko/tablewriter v0.0.5
	github.com/stretchr/testify v1.7.0
	sigs.k8s.io/kustomize/kyaml v0.13.0
)

replace github.com/kptdev/krm-functions-catalog/functions/go/list-setters => ../list-setters

require (
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-errors/errors v1.0.1 // indirect
	github.com/go-openapi/jsonpointer v0.19.3 // indirect
	github.com/go-openapi/jsonreference v0.19.3 // indirect
	github.com/go-openapi/swag v0.19.5 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.0 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/spf13/cobra v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/xlab/treeprint v0.0.0-20181112141820-a009c3971eca // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/kube-openapi v0.0.0-20210421082810-95288971da7e // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)
