package cdk

import "github.com/odpf/salt/printer"

func WithSpinner(fn func(done func()) error) error {
	spinner := printer.Spin("")
	defer spinner.Stop()

	return fn(func() {
		spinner.Stop()
	})
}
