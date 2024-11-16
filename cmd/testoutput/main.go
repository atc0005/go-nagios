// Copyright 2024 Adam Chalkley
//
// https://github.com/atc0005/go-nagios
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"

	"github.com/atc0005/go-nagios"
)

type testdata struct {
	input string
	name  string
}

var (
	//go:embed small_json_payload_unencoded.txt
	smallJSONPayloadUnencoded string

	// Earlier prototyping found that the stream encoding/decoding process
	// did not retain exclamation marks. I've not dug deep enough to
	// determine the root cause, but have observed that using the Decode
	// and Encode functions work reliably. Because a later refactoring
	// might switch back to using stream processing we explicitly test
	// using an exclamation point to guard against future breakage.
	//
	//go:embed small_plaintext_payload_unencoded.txt
	smallPlaintextPayloadUnencoded string

	//go:embed large_payload_unencoded.txt
	largePayloadUnencoded string
)

func main() {

	unencodedInputData := []testdata{
		{
			input: smallJSONPayloadUnencoded,
			name:  "smallJSONPayloadUnencoded",
		},
		{
			input: smallPlaintextPayloadUnencoded,
			name:  "smallPlaintextPayloadUnencoded",
		},
		{
			input: largePayloadUnencoded,
			name:  "largePayloadUnencoded",
		},
	}

	plugin := nagios.NewPlugin()

	outputFile, err := os.Create("scratch/i301-test-output.txt")
	if err != nil {
		fmt.Println("failed to create output file")

		return
	}

	defer func() {
		if err := outputFile.Close(); err != nil {
			fmt.Println("error occurred closing output file")
		}
	}()

	// We're going to short circuit the normal plugin exit behavior and block
	// os.Exit calls; we'll handle the exit behavior manually for our testing.
	plugin.SkipOSExit()

	// Setup output buffer
	// var buffer strings.Builder

	// Emit output here instead of stdout so that we can programmatically
	// inspect the results.
	plugin.SetOutputTarget(outputFile)

	// This is the default exit status code if not overridden. We just set it
	// explicitly to make that clear.
	plugin.ExitStatusCode = nagios.StateOKExitCode

	plugin.ServiceOutput = fmt.Sprintf(
		"%s: one-line service check summary here",
		nagios.StateOKLabel,
	)

	plugin.SetEncodedPayloadLabel("PAYLOAD FOR LATER API USE")

	for i, testData := range unencodedInputData {
		plugin.LongServiceOutput = fmt.Sprintf(
			"Sample detailed plugin output here for testData %q", testData.name)

		fmt.Println()

		fmt.Printf("Processing new testData entry %q (%d)\n", testData.name, i)

		if _, err := plugin.SetPayloadString(testData.input); err != nil {
			log.Printf("failed to add encoded payload for %q: %v", testData.name, err)
			plugin.Errors = append(plugin.Errors, err)

			return
		}

		// fmt.Println("Debugging unencoded payload content")
		// fmt.Print(plugin.UnencodedPayload())

		// This is usually deferred, but we're calling it explicitly because we're
		// going to evaluate the output buffer we just created.
		plugin.ReturnCheckResults()

		// Debugging
		// fmt.Print(buffer.String())
	}
}
