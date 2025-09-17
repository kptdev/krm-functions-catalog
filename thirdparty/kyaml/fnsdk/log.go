package fnsdk

import (
	"fmt"
	"os"
)

func Log(in ...interface{}) {
	fmt.Fprintln(os.Stderr, in...)
}

func Logf(format string, in ...interface{}) {
	fmt.Fprintf(os.Stderr, format, in...)
}
