// Copyright 2020 Adam Chalkley
//
// https://github.com/atc0005/go-nagios
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

package nagios_test

import (
	"fmt"
	"os"

	"github.com/atc0005/go-nagios"
)

// Ignore this. This is just to satisfy the "whole file" example requirements
// per https://go.dev/blog/examples.
var _ = "https://github.com/atc0005/go-nagios"

// ExampleUsingOnlyTheProvidedConstants is a simple example that illustrates
// using only the provided constants from this package. After you've imported
// this library, reference the exported data types as you would from any other
// package.
func Example_usingOnlyTheProvidedConstants() {
	// In this example, we reference a specific exit code for the OK state:
	fmt.Println("OK: All checks have passed")
	os.Exit(nagios.StateOKExitCode)

	// You can also use the provided state "labels" to avoid using literal
	// string state values (recommended):
	fmt.Printf(
		"%s: All checks have passed%s",
		nagios.StateOKLabel,
		nagios.CheckOutputEOL,
	)

	os.Exit(nagios.StateOKExitCode)

}
