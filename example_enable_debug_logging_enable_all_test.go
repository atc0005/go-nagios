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

// This example demonstrates enabling debug logging for all plugin activity
// types.
func Example_debugLoggingEnableAll() {
	// First, create an instance of the Plugin type. By default this value is
	// configured to indicate a successful execution. This should be
	// overridden by client code to indicate the final plugin state to Nagios
	// when the plugin exits.
	var plugin = nagios.NewPlugin()

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
	defer plugin.ReturnCheckResults()

	// Enable debug logging for all activity types. This will produce the most
	// verbose output but provides a comprehensive view of this library's
	// behavior.
	//
	// By default, debug logging output is sent to stderr but this can be
	// overridden as needed by setting a custom debug logging output target.
	plugin.DebugLoggingEnableAll()

	// more stuff here involving performing the actual service check

	plugin.ServiceOutput = "one-line summary of plugin results "       //nolint:goconst
	plugin.LongServiceOutput = "more detailed output from plugin here" //nolint:goconst

	// more stuff here involving wrapping up the service check
}
