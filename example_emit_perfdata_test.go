// Copyright 2020 Adam Chalkley
//
// https://github.com/atc0005/go-nagios
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

package nagios_test

import (
	"fmt"
	"log"
	"time"

	"github.com/atc0005/go-nagios"
)

// Ignore this. This is just to satisfy the "whole file" example requirements
// per https://go.dev/blog/examples.
var _ = "https://github.com/atc0005/go-nagios"

// ExampleEmitPerformanceData demonstrates providing multiple plugin
// performance data values explicitly.
//
// NOTE: While this example illustrates providing a time metric, this metric
// is provided for you if using the nagios.NewPlugin constructor. If
// specifying this value ourselves, *our* value takes precedence and the
// default value is ignored.
func Example_emitPerformanceData() {
	// Start the timer. We'll use this to emit the plugin runtime as a
	// performance data metric.
	//
	// NOTE: While this example illustrates providing a time metric, this
	// metric is provided for you if using the nagios.NewPlugin constructor.
	// If specifying this value ourselves, *our* value takes precedence and
	// the default value is ignored.
	pluginStart := time.Now()

	// First, create an instance of the Plugin type. Here we're opting to
	// manually construct the Plugin value instead of using the constructor
	// (mostly for contrast).
	//
	// We set the ExitStatusCode value to reflect a successful plugin
	// execution. If we do not alter the exit status code later this is what
	// will be reported to Nagios when the plugin exits.
	var plugin = nagios.Plugin{
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
	defer plugin.ReturnCheckResults()

	pd := []nagios.PerformanceData{
		{
			// NOTE: This metric is provided by default if using the provided
			// nagios.NewPlugin constructor.
			//
			// If we specify this value ourselves, *our* value takes
			// precedence and the default value is ignored.
			Label: "time",
			Value: fmt.Sprintf("%dms", time.Since(pluginStart).Milliseconds()),
		},
		{
			Label: "datacenters",
			Value: fmt.Sprintf("%d", 2),
		},
		{
			Label: "triggered_alarms",
			Value: fmt.Sprintf("%d", 14),
		},
	}

	if err := plugin.AddPerfData(false, pd...); err != nil {
		log.Printf("failed to add performance data metrics: %v", err)
		plugin.Errors = append(plugin.Errors, err)

		// NOTE: You might wish to make this a "best effort" scenario and
		// proceed with plugin execution. In that case, don't return yet until
		// further data has been gathered.
		return
	}

	// more stuff here

	plugin.ServiceOutput = "one-line summary of plugin results "

	plugin.LongServiceOutput = "more detailed output from plugin here"
}
