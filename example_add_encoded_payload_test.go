// Copyright 2020 Adam Chalkley
//
// https://github.com/atc0005/go-nagios
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

package nagios_test

import (
	"log"

	"github.com/atc0005/go-nagios"
)

// Ignore this. This is just to satisfy the "whole file" example requirements
// per https://go.dev/blog/examples.
var _ = "https://github.com/atc0005/go-nagios"

// Example_addEncodedPayload demonstrates adding an encoded payload to plugin
// output (e.g., for potential later retrieval via the monitor systems' API).
func Example_addEncodedPayload() {
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

	// more stuff here involving performing the actual service check

	// This simple JSON structure represents a more detailed blob of data that
	// might need post-processing once retrieved from the monitoring system's
	// API, etc.
	//
	// Imagine this is something like a full certificate chain or a system's
	// state and not easily represented by performance data metrics (e.g.,
	// best represented in a structured format once later extracted and
	// decoded).
	sampleComplexData := `{"Age":17,"Interests":["books","games", "Crystal Stix"]}`

	if _, err := plugin.AddPayloadString(sampleComplexData); err != nil {
		log.Printf("failed to add encoded payload: %v", err)
		plugin.Errors = append(plugin.Errors, err)

		return
	}

	//nolint:goconst
	plugin.ServiceOutput = "one-line summary of plugin results "

	//nolint:goconst
	plugin.LongServiceOutput = "more detailed output from plugin here"

	// more stuff here involving wrapping up the service check
}
