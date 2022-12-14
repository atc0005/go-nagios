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

// ExampleBasicPluginStructure demonstrates the basic structure for a
// monitoring plugin that uses this package.
func Example_basicPluginStructure() {
	// First, create an instance of the ExitState type. By default the
	// ExitState value is configured to indicate plugin success. This should
	// be overridden by client code to indicate the final plugin state to
	// Nagios when the plugin exits.
	var nagiosExitState = nagios.New()

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

	nagiosExitState.ServiceOutput = "one-line summary of plugin results "

	nagiosExitState.LongServiceOutput = "more detailed output from plugin here"
}
