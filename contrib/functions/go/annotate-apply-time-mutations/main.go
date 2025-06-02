package main

import (
	"fmt"
	"os"

	"github.com/kptdev/krm-functions-catalog/contrib/functions/go/annotate-apply-time-mutations/pkg"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
)

func main() {
	fn := pkg.Function{}
	cmd := command.Build(&fn, command.StandaloneEnabled, false)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
