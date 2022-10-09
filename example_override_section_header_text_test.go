// Copyright 2020 Adam Chalkley
//
// https://github.com/atc0005/go-nagios
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

package nagios_test

import (
	"github.com/atc0005/go-nagios"
)

// Ignore this. This is just to satisfy the "whole file" example requirements
// per https://go.dev/blog/examples.
var _ = "https://github.com/atc0005/go-nagios"

// ExampleOverrideSectionHeaders demonstrates overriding the default text with
// values that better fit our use case.
func Example_overrideSectionHeaders() {
	// First, create an instance of the ExitState type. Here we're
	// optimistic and we are going to assume that all will end well. If we do
	// not alter the exit status code later this is what will be reported to
	// Nagios when the plugin exits.
	var nagiosExitState = nagios.ExitState{
		LastError:      nil,
		ExitStatusCode: nagios.StateOKExitCode,
	}

	// Second, immediately defer ReturnCheckResults() so that it runs as the
	// last step in your client code. If you do not defer ReturnCheckResults()
	// immediately any other deferred functions in your client code will not
	// run.
	//
	// Avoid calling os.Exit() directly from your code. If you do, this
	// library is unable to function properly; this library expects that it
	// will handle calling os.Exit() with the required exit code (and
	// specifically formatted output).
	//
	// For handling error cases, the approach is roughly the same, only you
	// call return explicitly to end execution of the client code and allow
	// deferred functions to run.
	defer nagiosExitState.ReturnCheckResults()

	// more stuff here

	// Override default section headers with our custom values.
	nagiosExitState.SetErrorsLabel("VALIDATION ERRORS")
	nagiosExitState.SetDetailedInfoLabel("VALIDATION CHECKS REPORT")
}
