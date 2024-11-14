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
	"os"

	"github.com/atc0005/go-nagios"
)

// Ignore this. This is just to satisfy the "whole file" example requirements
// per https://go.dev/blog/examples.
var _ = "https://github.com/atc0005/go-nagios"

// Example_extractEncodedPayload represents a sample client application that
// extracts a previously encoded payload from plugin output (e.g., retrieved
// via the monitoring system's API or a log file).
func Example_extractEncodedPayload() {
	// This represents the data before it was encoded and added to the plugin
	// output.
	origData := `{"Age":17,"Interests":["books","games", "Crystal Stix"]}`

	// This represents the encoded data that was previously added to the
	// plugin output.
	encodedData := `<~HQkagAKj/i2_6.EDKKH1ATMs7,!&pP@W-1#F!<.ZB45XgF!<.X,"$BrF*(i,+B*ArGTpFA~>`

	sampleServiceOutput := "one-line summary of plugin results "
	sampleLongServiceOutput := "detailed text line1\ndetailed text line2\ndetailed text line3"
	sampleMetricsOutput := `| 'time'=874ms;;;;`

	// This represents the original plugin output captured by the monitoring
	// system which we retrieved via the monitoring system's API, a log file,
	// etc.
	originalPluginOutput := fmt.Sprintf(
		"%s%s%s%s%s%s%s",
		sampleServiceOutput,
		sampleLongServiceOutput,
		nagios.CheckOutputEOL,
		encodedData,
		nagios.CheckOutputEOL,
		sampleMetricsOutput,
		nagios.CheckOutputEOL,
	)

	decodedPayload, err := nagios.ExtractAndDecodePayload(
		originalPluginOutput,
		"",
		nagios.DefaultASCII85EncodingDelimiterLeft,
		nagios.DefaultASCII85EncodingDelimiterRight,
	)

	if err != nil {
		log.Println("Failed to extract and decode payload from original plugin output", err)

		os.Exit(1)
	}

	// We compare here for illustration purposes, but in many cases you may
	// not have access to the data in its original form as collected by the
	// monitoring plugin (e.g., it was generated dynamically or retrieved from
	// a remote system's state and not stored long-term).
	if decodedPayload != origData {
		log.Println("Extracted & decoded payload data does not match original data")

		os.Exit(1)
	}

	fmt.Println("Original data:", decodedPayload)

	// Output:
	//
	// Original data: {"Age":17,"Interests":["books","games", "Crystal Stix"]}
}
