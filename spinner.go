package main

import (
	"fmt"
	"os"

	"github.com/kovetskiy/spinner-go"
)

func withSpinner(status string, method func() error) error {
	if debugMode {
		fmt.Fprintln(os.Stderr, status)
		return method()
	}

	return spinner.SetStatus(status + " ").Call(method)
}
