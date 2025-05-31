package main

import (
	"os"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/configmap-generator/generator"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

func main() {
	if err := fn.AsMain(&generator.ConfigMapGenerator{}); err != nil {
		os.Exit(1)
	}
}
