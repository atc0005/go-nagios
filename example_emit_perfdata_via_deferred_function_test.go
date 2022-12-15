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

// ExampleEmitPerformanceDataViaDeferredAnonymousFunc demonstrates emitting
// plugin performance data provided via a deferred anonymous function.
//
// NOTE: While this example illustrates providing a time metric, this metric
// is provided for you if using the nagios.NewPlugin constructor. If specifying
// this value ourselves, *our* value takes precedence and the default value is
// ignored.
func Example_emitPerformanceDataViaDeferredAnonymousFunc() {

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

	// Collect last minute details just before ending plugin execution.
	defer func(plugin *nagios.Plugin, start time.Time) {

		// Record plugin runtime, emit this metric regardless of exit
		// point/cause.
		runtimeMetric := nagios.PerformanceData{
			// NOTE: This metric is provided by default if using the provided
			// nagios.NewPlugin constructor.
			//
			// If we specify this value ourselves, *our* value takes
			// precedence and the default value is ignored.
			Label: "time",
			Value: fmt.Sprintf("%dms", time.Since(start).Milliseconds()),
		}
		if err := plugin.AddPerfData(false, runtimeMetric); err != nil {
			log.Printf("failed to add time (runtime) performance data metric: %v", err)
			plugin.Errors = append(plugin.Errors, err)
		}

	}(&plugin, pluginStart)

	// more stuff here

	//nolint:goconst
	plugin.ServiceOutput = "one-line summary text here"

	//nolint:goconst
	plugin.LongServiceOutput = "more detailed output here"
}
