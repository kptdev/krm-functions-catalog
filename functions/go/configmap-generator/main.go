package main

import (
	"os"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/configmap-generator/fn"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/configmap-generator/generator"
)

func main() {
	if err := fn.AsMain(&generator.ConfigMapGenerator{}); err != nil {
		os.Exit(1)
	}
}
