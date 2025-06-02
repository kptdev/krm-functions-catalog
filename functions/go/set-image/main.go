package main

import (
	"os"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"github.com/kptdev/krm-functions-catalog/functions/go/set-image/transformer"
)

func main() {
	if err := fn.AsMain(&transformer.SetImage{}); err != nil {
		os.Exit(1)
	}
}
