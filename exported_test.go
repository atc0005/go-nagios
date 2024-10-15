// Copyright 2020 Adam Chalkley
//
// https://github.com/atc0005/go-nagios
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

// Package nagios_test provides test coverage for exported package
// functionality.
//
//nolint:dupl,gocognit // ignore "lines are duplicate of" and function complexity
package nagios_test

import (
	_ "embed"
	"encoding/ascii85"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/atc0005/go-nagios"
	"github.com/google/go-cmp/cmp"
)

const (
	customEncodingDelimiterLeft  string = "CUSTOM_ENCODING_DELIMITER_LEFT"
	customEncodingDelimiterRight string = "CUSTOM_ENCODING_DELIMITER_RIGHT"
	customSectionHeader          string = "PAYLOAD FOR LATER API USE"
)

// The specific format used by the test input files is VERY specific; trailing
// space + newline patterns are intentional. Because "format on save" editor
// functionality easily breaks this input it is stored in separate files to
// reduce test breakage due to editors "helping".
var (
	//go:embed testdata/plugin-output-datastore-0001.txt
	pluginOutputDatastore0001 string

	//go:embed testdata/plugin-output-gh103-multi-line-with-perf-data.txt
	pluginOutputGH103MultiLineWithPerfData string

	//go:embed testdata/plugin-output-gh103-one-line-with-perf-data.txt
	pluginOutputGH103OneLineWithPerfData string

	//go:embed testdata/payload/small_json_payload_unencoded.txt
	smallJSONPayloadUnencoded string

	//go:embed testdata/payload/small_json_payload_encoded_with_default_delimiters.txt
	smallJSONPayloadEncodedWithDefaultDelimiters string

	//go:embed testdata/payload/small_json_payload_encoded_with_custom_delimiters.txt
	smallJSONPayloadEncodedWithCustomDelimiters string

	//go:embed testdata/payload/small_json_payload_encoded_without_delimiters.txt
	smallJSONPayloadEncodedWithNoDelimiters string

	// Earlier prototyping found that the stream encoding/decoding process
	// did not retain exclamation marks. I've not dug deep enough to
	// determine the root cause, but have observed that using the Decode
	// and Encode functions work reliably. Because a later refactoring
	// might switch back to using stream processing we explicitly test
	// using an exclamation point to guard against future breakage.
	//
	//go:embed testdata/payload/small_plaintext_payload_unencoded.txt
	smallPlaintextPayloadUnencoded string

	// Earlier prototyping found that the stream encoding/decoding process
	// did not retain exclamation marks. I've not dug deep enough to
	// determine the root cause, but have observed that using the Decode
	// and Encode functions work reliably. Because a later refactoring
	// might switch back to using stream processing we explicitly test
	// using an exclamation point to guard against future breakage.
	//
	//go:embed testdata/payload/small_plaintext_payload_encoded_with_default_delimiters.txt
	smallPlaintextPayloadEncodedWithDefaultDelimiters string

	// Earlier prototyping found that the stream encoding/decoding process
	// did not retain exclamation marks. I've not dug deep enough to
	// determine the root cause, but have observed that using the Decode
	// and Encode functions work reliably. Because a later refactoring
	// might switch back to using stream processing we explicitly test
	// using an exclamation point to guard against future breakage.
	//
	//go:embed testdata/payload/small_plaintext_payload_encoded_with_custom_delimiters.txt
	smallPlaintextPayloadEncodedWithCustomDelimiters string

	// Earlier prototyping found that the stream encoding/decoding process
	// did not retain exclamation marks. I've not dug deep enough to
	// determine the root cause, but have observed that using the Decode
	// and Encode functions work reliably. Because a later refactoring
	// might switch back to using stream processing we explicitly test
	// using an exclamation point to guard against future breakage.
	//
	//go:embed testdata/payload/small_plaintext_payload_encoded_without_delimiters.txt
	smallPlaintextPayloadEncodedWithNoDelimiters string

	//go:embed testdata/payload/large_payload_encoded_with_default_delimiters.txt
	largePayloadEncodedWithDefaultDelimiters string

	//go:embed testdata/payload/large_payload_encoded_with_custom_delimiters.txt
	largePayloadEncodedWithCustomDelimiters string

	//go:embed testdata/payload/large_payload_encoded_without_delimiters.txt
	largePayloadEncodedWithNoDelimiters string

	//go:embed testdata/payload/large_payload_unencoded.txt
	largePayloadUnencoded string

	//go:embed testdata/payload/plugin-output-gh251-custom-section-header-and-large-encoded-payload-with-custom-delimiters.txt
	pluginOutputCustomSectionHeadersAndLargeEncodedPayloadWithCustomDelimiters string

	//go:embed testdata/payload/plugin-output-gh251-custom-section-header-and-large-encoded-payload-with-default-delimiters.txt
	pluginOutputCustomSectionHeadersAndLargeEncodedPayloadWithDefaultDelimiters string

	//go:embed testdata/payload/plugin-output-gh251-custom-section-header-and-large-encoded-payload-with-no-delimiters.txt
	pluginOutputCustomSectionHeadersAndLargeEncodedPayloadWithNoDelimiters string

	//go:embed testdata/payload/plugin-output-gh251-custom-section-header-and-small-encoded-json-payload-with-custom-delimiters.txt
	pluginOutputCustomSectionHeadersAndSmallEncodedJSONPayloadWithCustomDelimiters string

	//go:embed testdata/payload/plugin-output-gh251-custom-section-header-and-small-encoded-json-payload-with-default-delimiters.txt
	pluginOutputCustomSectionHeadersAndSmallEncodedJSONPayloadWithDefaultDelimiters string

	//go:embed testdata/payload/plugin-output-gh251-custom-section-header-and-small-encoded-json-payload-with-no-delimiters.txt
	pluginOutputCustomSectionHeadersAndSmallEncodedJSONPayloadWithNoDelimiters string

	//go:embed testdata/payload/plugin-output-gh251-custom-section-header-and-small-encoded-plaintext-payload-with-custom-delimiters.txt
	pluginOutputCustomSectionHeadersAndSmallEncodedPlaintextPayloadWithCustomDelimiters string

	//go:embed testdata/payload/plugin-output-gh251-custom-section-header-and-small-encoded-plaintext-payload-with-default-delimiters.txt
	pluginOutputCustomSectionHeadersAndSmallEncodedPlaintextPayloadWithDefaultDelimiters string

	//go:embed testdata/payload/plugin-output-gh251-custom-section-header-and-small-encoded-plaintext-payload-with-no-delimiters.txt
	pluginOutputCustomSectionHeadersAndSmallEncodedPlaintextPayloadWithNoDelimiters string

	//go:embed testdata/payload/plugin-output-gh251-default-section-header-and-large-encoded-payload-with-custom-delimiters.txt
	pluginOutputDefaultSectionHeadersAndLargeEncodedPayloadWithCustomDelimiters string

	//go:embed testdata/payload/plugin-output-gh251-default-section-header-and-large-encoded-payload-with-default-delimiters.txt
	pluginOutputDefaultSectionHeadersAndLargeEncodedPayloadWithDefaultDelimiters string

	//go:embed testdata/payload/plugin-output-gh251-default-section-header-and-large-encoded-payload-with-no-delimiters.txt
	pluginOutputDefaultSectionHeadersAndLargeEncodedPayloadWithNoDelimiters string

	//go:embed testdata/payload/plugin-output-gh251-default-section-header-and-small-encoded-json-payload-with-custom-delimiters.txt
	pluginOutputDefaultSectionHeadersAndSmallEncodedJSONPayloadWithCustomDelimiters string

	//go:embed testdata/payload/plugin-output-gh251-default-section-header-and-small-encoded-json-payload-with-default-delimiters.txt
	pluginOutputDefaultSectionHeadersAndSmallEncodedJSONPayloadWithDefaultDelimiters string

	//go:embed testdata/payload/plugin-output-gh251-default-section-header-and-small-encoded-json-payload-with-no-delimiters.txt
	pluginOutputDefaultSectionHeadersAndSmallEncodedJSONPayloadWithNoDelimiters string

	//go:embed testdata/payload/plugin-output-gh251-default-section-header-and-small-encoded-plaintext-payload-with-custom-delimiters.txt
	pluginOutputDefaultSectionHeadersAndSmallEncodedPlaintextPayloadWithCustomDelimiters string

	//go:embed testdata/payload/plugin-output-gh251-default-section-header-and-small-encoded-plaintext-payload-with-default-delimiters.txt
	pluginOutputDefaultSectionHeadersAndSmallEncodedPlaintextPayloadWithDefaultDelimiters string

	//go:embed testdata/payload/plugin-output-gh251-default-section-header-and-small-encoded-plaintext-payload-with-no-delimiters.txt
	pluginOutputDefaultSectionHeadersAndSmallEncodedPlaintextPayloadWithNoDelimiters string
)

// TestPluginOutputIsValid configures an Plugin value as client code would
// and then asserts that the generated output matches manually crafted test
// data.
func TestPluginOutputIsValid(t *testing.T) {
	t.Parallel()

	want := pluginOutputDatastore0001

	// Setup Plugin value manually. This approach does not provide the
	// default time metric that would be provided when using the Plugin
	// constructor.
	plugin := nagios.Plugin{
		LastError:      nil,
		ExitStatusCode: nagios.StateOKExitCode,
	}

	var outputBuffer strings.Builder
	plugin.SetOutputTarget(&outputBuffer)

	// os.Exit calls break tests
	plugin.SkipOSExit()

	pluginOutputWithLongServiceOutputSetup(t, &plugin)

	// Process exit state, emit output to our output buffer.
	plugin.ReturnCheckResults()

	// Retrieve the output buffer content so that we can compare actual output
	// against our expected output to assert we have a 1:1 match.
	got := outputBuffer.String()

	// if want  != got {
	// 	t.Error(cmp.Diff(want, got))
	// }
	if d := cmp.Diff(want, got); d != "" {
		t.Errorf("(-want, +got)\n:%s", d)
	}
}

// TestPerformanceDataIsOnSameLineAsServiceOutput asserts that performance
// data is emitted on the same line as the Service Output (aka, "one-line
// summary") if Long Service Output is empty.
//
// See also:
//
// - https://github.com/atc0005/go-nagios/issues/103
func TestPerformanceDataIsOnSameLineAsServiceOutput(t *testing.T) {
	t.Parallel()

	want := pluginOutputGH103OneLineWithPerfData

	// Setup Plugin value manually. This approach does not provide the
	// default time metric that would be provided when using the Plugin
	// constructor.
	plugin := nagios.Plugin{
		LastError:      nil,
		ExitStatusCode: nagios.StateOKExitCode,
	}

	var outputBuffer strings.Builder
	plugin.SetOutputTarget(&outputBuffer)

	// os.Exit calls break tests
	plugin.SkipOSExit()

	//nolint:goconst
	plugin.ServiceOutput =
		"OK: Datastore HUSVM-DC1-vol6 space usage (0 VMs)" +
			" is 0.01% of 18.0TB with 18.0TB remaining" +
			" [WARNING: 90% , CRITICAL: 95%]"

	pluginOutputWithLongServiceOutputMetrics(t, &plugin)

	// Process exit state, emit output to our output buffer.
	plugin.ReturnCheckResults()

	// Retrieve the output buffer content so that we can compare actual output
	// against our expected output to assert we have a 1:1 match.
	got := outputBuffer.String()

	if d := cmp.Diff(want, got); d != "" {
		t.Errorf("(-want, +got)\n:%s", d)
	}
}

// TestPerformanceDataIsAfterLongServiceOutput asserts that performance data
// is emitted after Long Service Output when that content is available.
//
// See also:
//
// - https://github.com/atc0005/go-nagios/issues/103
func TestPerformanceDataIsAfterLongServiceOutput(t *testing.T) {
	t.Parallel()

	want := pluginOutputGH103MultiLineWithPerfData

	var outputBuffer strings.Builder

	// Setup Plugin value manually. This approach does not provide the
	// default time metric that would be provided when using the Plugin
	// constructor.
	plugin := nagios.Plugin{
		LastError:      nil,
		ExitStatusCode: nagios.StateOKExitCode,
	}

	plugin.SetOutputTarget(&outputBuffer)

	// os.Exit calls break tests
	plugin.SkipOSExit()

	pluginOutputWithLongServiceOutputSetup(t, &plugin)
	pluginOutputWithLongServiceOutputMetrics(t, &plugin)

	// Process exit state, emit output to our output buffer.
	plugin.ReturnCheckResults()

	// Retrieve the output buffer content so that we can compare actual output
	// against our expected output to assert we have a 1:1 match.
	got := outputBuffer.String()

	if d := cmp.Diff(want, got); d != "" {
		t.Errorf("(-want, +got)\n:%s", d)
	}
}

// TestEmptyServiceOutputAndManuallyConstructedPluginProducesNoOutput
// asserts that an empty ServiceOutput field produces no output when manually
// constructing the Plugin value.
func TestEmptyServiceOutputAndManuallyConstructedPluginProducesNoOutput(t *testing.T) {
	t.Parallel()

	// Setup Plugin value manually. This approach does not provide the
	// default time metric that would be provided when using the Plugin
	// constructor.
	plugin := nagios.Plugin{
		LastError:      nil,
		ExitStatusCode: nagios.StateOKExitCode,
	}

	var outputBuffer strings.Builder

	// Explicitly indicate that the field is empty (default/zero value).
	plugin.ServiceOutput = ""

	plugin.SetOutputTarget(&outputBuffer)

	// os.Exit calls break tests
	plugin.SkipOSExit()

	// Process exit state, emit output to our output buffer.
	plugin.ReturnCheckResults()

	want := ""

	// Retrieve the output buffer content so that we can compare actual output
	// against our expected output to assert we have a 1:1 match.
	got := outputBuffer.String()

	if d := cmp.Diff(want, got); d != "" {
		t.Errorf("(-want, +got)\n:%s", d)
	} else {
		t.Logf("OK: Empty ServiceOutput field produces no output.")
	}

}

// TestEmptyServiceOutputAndConstructedPluginProducesNoOutput asserts that
// an empty ServiceOutput field produces no output. We provide a default time
// metric if client code does not specify one AND if there is ServiceOutput
// content to emit.
func TestEmptyServiceOutputAndConstructedPluginProducesNoOutput(t *testing.T) {
	t.Parallel()

	// Setup Plugin type the same way that client code using the
	// constructor would.
	plugin := nagios.NewPlugin()

	var outputBuffer strings.Builder

	// Explicitly indicate that the field is empty (default/zero value).
	plugin.ServiceOutput = ""

	plugin.SetOutputTarget(&outputBuffer)

	// os.Exit calls break tests
	plugin.SkipOSExit()

	// Process exit state, emit output to our output buffer.
	plugin.ReturnCheckResults()

	want := ""

	// Retrieve the output buffer content so that we can compare actual output
	// against our expected output to assert we have a 1:1 match.
	got := outputBuffer.String()

	if d := cmp.Diff(want, got); d != "" {
		t.Errorf("(-want, +got)\n:%s", d)
	} else {
		t.Logf("OK: Empty ServiceOutput field produces no output.")
	}

}

// TestEmptyClientPerfDataAndConstructedPluginProducesDefaultTimeMetric
// asserts that omitted performance data from client code produces a default
// time metric when using the Plugin constructor.
func TestEmptyClientPerfDataAndConstructedPluginProducesDefaultTimeMetric(t *testing.T) {
	t.Parallel()

	// Setup Plugin type the same way that client code using the
	// constructor would.
	plugin := nagios.NewPlugin()

	// Performance Data metrics are not emitted if we do not supply a
	// ServiceOutput value.
	plugin.ServiceOutput = "TacoTuesday"

	var outputBuffer strings.Builder

	plugin.SetOutputTarget(&outputBuffer)

	// os.Exit calls break tests
	plugin.SkipOSExit()

	// Process exit state, emit output to our output buffer.
	plugin.ReturnCheckResults()

	want := fmt.Sprintf(
		"%s | %s",
		plugin.ServiceOutput,
		"'time'=",
	)

	got := outputBuffer.String()

	if !strings.Contains(got, want) {
		t.Errorf("ERROR: Plugin output does not contain the expected time metric")
		t.Errorf("\nwant %q\ngot %q", want, got)
	} else {
		t.Logf("OK: Emitted performance data contains the expected time metric.")
	}
}

// TestPluginWithEncodedPayloadWithValidInputProducesValidOutput asserts that
// a populated payload buffer with valid delimiters and section header
// produces valid output.
func TestPluginWithEncodedPayloadWithValidInputProducesValidOutput(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		unencodedPayloadInput string
		expectedOutput        string
		delimiterLeft         string
		delimiterRight        string
		sectionHeader         string
	}{
		"custom section header and large encoded payload and custom delimiters": {
			unencodedPayloadInput: largePayloadUnencoded,
			expectedOutput:        pluginOutputCustomSectionHeadersAndLargeEncodedPayloadWithCustomDelimiters,
			delimiterLeft:         customEncodingDelimiterLeft,
			delimiterRight:        customEncodingDelimiterRight,
			sectionHeader:         customSectionHeader,
		},
		"custom section header and large encoded payload and default delimiters": {
			unencodedPayloadInput: largePayloadUnencoded,
			expectedOutput:        pluginOutputCustomSectionHeadersAndLargeEncodedPayloadWithDefaultDelimiters,
			delimiterLeft:         nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight:        nagios.DefaultASCII85EncodingDelimiterRight,
			sectionHeader:         customSectionHeader,
		},
		"custom section header and large encoded payload and no delimiters": {
			unencodedPayloadInput: largePayloadUnencoded,
			expectedOutput:        pluginOutputCustomSectionHeadersAndLargeEncodedPayloadWithNoDelimiters,
			delimiterLeft:         "",
			delimiterRight:        "",
			sectionHeader:         customSectionHeader,
		},
		"custom section header and small encoded json payload and custom delimiters": {
			unencodedPayloadInput: smallJSONPayloadUnencoded,
			expectedOutput:        pluginOutputCustomSectionHeadersAndSmallEncodedJSONPayloadWithCustomDelimiters,
			delimiterLeft:         customEncodingDelimiterLeft,
			delimiterRight:        customEncodingDelimiterRight,
			sectionHeader:         customSectionHeader,
		},
		"custom section header and small encoded json payload and default delimiters": {
			unencodedPayloadInput: smallJSONPayloadUnencoded,
			expectedOutput:        pluginOutputCustomSectionHeadersAndSmallEncodedJSONPayloadWithDefaultDelimiters,
			delimiterLeft:         nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight:        nagios.DefaultASCII85EncodingDelimiterRight,
			sectionHeader:         customSectionHeader,
		},
		"custom section header and small encoded json payload and no delimiters": {
			unencodedPayloadInput: smallJSONPayloadUnencoded,
			expectedOutput:        pluginOutputCustomSectionHeadersAndSmallEncodedJSONPayloadWithNoDelimiters,
			delimiterLeft:         "",
			delimiterRight:        "",
			sectionHeader:         customSectionHeader,
		},
		"custom section header and small encoded plaintext payload and custom delimiters": {
			unencodedPayloadInput: smallPlaintextPayloadUnencoded,
			expectedOutput:        pluginOutputCustomSectionHeadersAndSmallEncodedPlaintextPayloadWithCustomDelimiters,
			delimiterLeft:         customEncodingDelimiterLeft,
			delimiterRight:        customEncodingDelimiterRight,
			sectionHeader:         customSectionHeader,
		},
		"custom section header and small encoded plaintext payload and default delimiters": {
			unencodedPayloadInput: smallPlaintextPayloadUnencoded,
			expectedOutput:        pluginOutputCustomSectionHeadersAndSmallEncodedPlaintextPayloadWithDefaultDelimiters,
			delimiterLeft:         nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight:        nagios.DefaultASCII85EncodingDelimiterRight,
			sectionHeader:         customSectionHeader,
		},
		"custom section header and small encoded plaintext payload and no delimiters": {
			unencodedPayloadInput: smallPlaintextPayloadUnencoded,
			expectedOutput:        pluginOutputCustomSectionHeadersAndSmallEncodedPlaintextPayloadWithNoDelimiters,
			delimiterLeft:         "",
			delimiterRight:        "",
			sectionHeader:         customSectionHeader,
		},
		"default section header and large encoded payload and custom delimiters": {
			unencodedPayloadInput: largePayloadUnencoded,
			expectedOutput:        pluginOutputDefaultSectionHeadersAndLargeEncodedPayloadWithCustomDelimiters,
			delimiterLeft:         customEncodingDelimiterLeft,
			delimiterRight:        customEncodingDelimiterRight,
			sectionHeader:         "",
		},
		"default section header and large encoded payload and default delimiters": {
			unencodedPayloadInput: largePayloadUnencoded,
			expectedOutput:        pluginOutputDefaultSectionHeadersAndLargeEncodedPayloadWithDefaultDelimiters,
			delimiterLeft:         nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight:        nagios.DefaultASCII85EncodingDelimiterRight,
			sectionHeader:         "",
		},
		"default section header and large encoded payload and no delimiters": {
			unencodedPayloadInput: largePayloadUnencoded,
			expectedOutput:        pluginOutputDefaultSectionHeadersAndLargeEncodedPayloadWithNoDelimiters,
			delimiterLeft:         "",
			delimiterRight:        "",
			sectionHeader:         "",
		},
		"default section header and small encoded json payload and custom delimiters": {
			unencodedPayloadInput: smallJSONPayloadUnencoded,
			expectedOutput:        pluginOutputDefaultSectionHeadersAndSmallEncodedJSONPayloadWithCustomDelimiters,
			delimiterLeft:         customEncodingDelimiterLeft,
			delimiterRight:        customEncodingDelimiterRight,
			sectionHeader:         "",
		},
		"default section header and small encoded json payload and default delimiters": {
			unencodedPayloadInput: smallJSONPayloadUnencoded,
			expectedOutput:        pluginOutputDefaultSectionHeadersAndSmallEncodedJSONPayloadWithDefaultDelimiters,
			delimiterLeft:         nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight:        nagios.DefaultASCII85EncodingDelimiterRight,
			sectionHeader:         "",
		},
		"default section header and small encoded json payload and no delimiters": {
			unencodedPayloadInput: smallJSONPayloadUnencoded,
			expectedOutput:        pluginOutputDefaultSectionHeadersAndSmallEncodedJSONPayloadWithNoDelimiters,
			delimiterLeft:         "",
			delimiterRight:        "",
			sectionHeader:         "",
		},
		"default section header and small encoded plaintext payload and custom delimiters": {
			unencodedPayloadInput: smallPlaintextPayloadUnencoded,
			expectedOutput:        pluginOutputDefaultSectionHeadersAndSmallEncodedPlaintextPayloadWithCustomDelimiters,
			delimiterLeft:         customEncodingDelimiterLeft,
			delimiterRight:        customEncodingDelimiterRight,
			sectionHeader:         "",
		},
		"default section header and small encoded plaintext payload and default delimiters": {
			unencodedPayloadInput: smallPlaintextPayloadUnencoded,
			expectedOutput:        pluginOutputDefaultSectionHeadersAndSmallEncodedPlaintextPayloadWithDefaultDelimiters,
			delimiterLeft:         nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight:        nagios.DefaultASCII85EncodingDelimiterRight,
			sectionHeader:         "",
		},
		"default section header and small encoded plaintext payload and no delimiters": {
			unencodedPayloadInput: smallPlaintextPayloadUnencoded,
			expectedOutput:        pluginOutputDefaultSectionHeadersAndSmallEncodedPlaintextPayloadWithNoDelimiters,
			delimiterLeft:         "",
			delimiterRight:        "",
			sectionHeader:         "",
		},
	}

	for name, tt := range tests {
		// Guard against referencing the loop iterator variable directly.
		//
		// https://stackoverflow.com/questions/68559574/using-the-variable-on-range-scope-x-in-function-literal-scopelint
		// https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		tt := tt

		//
		t.Run(name, func(t *testing.T) {
			// Setup Plugin type the same way that client code using the
			// constructor would.
			plugin := nagios.NewPlugin()

			var outputBuffer strings.Builder

			plugin.SetOutputTarget(&outputBuffer)

			// os.Exit calls break tests
			plugin.SkipOSExit()

			pluginOutputWithLongServiceOutputSetup(t, plugin)

			// Current logic is that you need to specify blank delimiters to
			// skip including delimiters in the encoded output. Because of how
			// we're setting up the test cases we treat an empty string as if
			// the `SetEncodedPayloadDelimiterLeft` or
			// `SetEncodedPayloadDelimiterRight` methods are called with an
			// empty string.
			plugin.SetEncodedPayloadDelimiterLeft(tt.delimiterLeft)
			plugin.SetEncodedPayloadDelimiterRight(tt.delimiterRight)

			if tt.sectionHeader != "" {
				plugin.SetEncodedPayloadLabel(tt.sectionHeader)
			}

			written, err := plugin.AddPayloadString(tt.unencodedPayloadInput)
			if err != nil {
				t.Fatalf("Failed to append given input to payload buffer: %v", err)
			} else {
				t.Logf("Successfully appended %d bytes given input to payload buffer", written)
			}

			pluginOutputWithLongServiceOutputMetrics(t, plugin)

			// Process exit state, emit output to our output buffer.
			plugin.ReturnCheckResults()

			want := tt.expectedOutput
			got := outputBuffer.String()

			if d := cmp.Diff(want, got); d != "" {
				t.Errorf("(-want, +got)\n:%s", d)
			} else {
				t.Logf("OK: Plugin output matches expected format.")
			}
		})
	}
}

func TestEncodeASCII85Payload_SuccessfullyEncodesPayloadWithValidInput(t *testing.T) {
	t.Parallel()

	type testCase struct {
		unencodedPayloadInput []byte
		delimiterLeft         string
		delimiterRight        string
		expectedOutput        string
	}

	tests := map[string]testCase{
		"small unencoded json payload with default delimiters added": {
			unencodedPayloadInput: []byte(smallJSONPayloadUnencoded),
			delimiterLeft:         nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight:        nagios.DefaultASCII85EncodingDelimiterRight,
			expectedOutput:        smallJSONPayloadEncodedWithDefaultDelimiters,
		},
		"small unencoded json payload with custom delimiters added": {
			unencodedPayloadInput: []byte(smallJSONPayloadUnencoded),
			delimiterLeft:         customEncodingDelimiterLeft,
			delimiterRight:        customEncodingDelimiterRight,
			expectedOutput:        smallJSONPayloadEncodedWithCustomDelimiters,
		},
		"small unencoded json payload with no delimiters added": {
			unencodedPayloadInput: []byte(smallJSONPayloadUnencoded),
			delimiterLeft:         "",
			delimiterRight:        "",
			expectedOutput:        smallJSONPayloadEncodedWithNoDelimiters,
		},
		"small unencoded plaintext payload with default delimiters added": {
			unencodedPayloadInput: []byte(smallPlaintextPayloadUnencoded),
			delimiterLeft:         nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight:        nagios.DefaultASCII85EncodingDelimiterRight,
			expectedOutput:        smallPlaintextPayloadEncodedWithDefaultDelimiters,
		},
		"small unencoded plaintext payload with custom delimiters added": {
			unencodedPayloadInput: []byte(smallPlaintextPayloadUnencoded),
			delimiterLeft:         customEncodingDelimiterLeft,
			delimiterRight:        customEncodingDelimiterRight,
			expectedOutput:        smallPlaintextPayloadEncodedWithCustomDelimiters,
		},
		"small unencoded plaintext payload with no delimiters added": {
			unencodedPayloadInput: []byte(smallPlaintextPayloadUnencoded),
			delimiterLeft:         "",
			delimiterRight:        "",
			expectedOutput:        smallPlaintextPayloadEncodedWithNoDelimiters,
		},
		"large unencoded plaintext payload with default delimiters added": {
			unencodedPayloadInput: []byte(largePayloadUnencoded),
			delimiterLeft:         nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight:        nagios.DefaultASCII85EncodingDelimiterRight,
			expectedOutput:        largePayloadEncodedWithDefaultDelimiters,
		},
		"large unencoded plaintext payload with custom delimiters added": {
			unencodedPayloadInput: []byte(largePayloadUnencoded),
			delimiterLeft:         customEncodingDelimiterLeft,
			delimiterRight:        customEncodingDelimiterRight,
			expectedOutput:        largePayloadEncodedWithCustomDelimiters,
		},
		"large unencoded plaintext payload with no delimiters added": {
			unencodedPayloadInput: []byte(largePayloadUnencoded),
			delimiterLeft:         "",
			delimiterRight:        "",
			expectedOutput:        largePayloadEncodedWithNoDelimiters,
		},
	}

	for name, tt := range tests {
		// Guard against referencing the loop iterator variable directly.
		//
		// https://stackoverflow.com/questions/68559574/using-the-variable-on-range-scope-x-in-function-literal-scopelint
		// https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		tt := tt

		//
		t.Run(name, func(t *testing.T) {
			want := tt.expectedOutput

			got := nagios.EncodeASCII85Payload(
				tt.unencodedPayloadInput,
				tt.delimiterLeft,
				tt.delimiterRight,
			)

			if d := cmp.Diff(want, got); d != "" {
				t.Errorf("(-want, +got)\n:%s", d)
			} else {
				t.Logf("OK: Encoded payload matches expected output.")
			}
		})
	}
}

func TestEncodeASCII85Payload_FailsToEncodePayloadWithInvalidInput(t *testing.T) {
	t.Parallel()

	type testCase struct {
		unencodedPayloadInput []byte
		delimiterLeft         string
		delimiterRight        string
		expectedOutput        string
	}

	tests := map[string]testCase{
		"empty input with default delimiters": {
			unencodedPayloadInput: []byte(""),
			delimiterLeft:         nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight:        nagios.DefaultASCII85EncodingDelimiterRight,
			expectedOutput:        "",
		},
		"empty input with custom delimiters": {
			unencodedPayloadInput: []byte(""),
			delimiterLeft:         customEncodingDelimiterLeft,
			delimiterRight:        customEncodingDelimiterRight,
			expectedOutput:        "",
		},
		"empty input with no delimiters": {
			unencodedPayloadInput: []byte(""),
			delimiterLeft:         "",
			delimiterRight:        "",
			expectedOutput:        "",
		},
	}

	for name, tt := range tests {
		// Guard against referencing the loop iterator variable directly.
		//
		// https://stackoverflow.com/questions/68559574/using-the-variable-on-range-scope-x-in-function-literal-scopelint
		// https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		tt := tt

		//
		t.Run(name, func(t *testing.T) {
			want := tt.expectedOutput

			got := nagios.EncodeASCII85Payload(
				tt.unencodedPayloadInput,
				tt.delimiterLeft,
				tt.delimiterRight,
			)

			if d := cmp.Diff(want, got); d != "" {
				t.Errorf("(-want, +got)\n:%s", d)
			} else {
				t.Logf("OK: Encoded payload matches expected output.")
			}
		})
	}
}

func TestDecodeASCII85Payload_SuccessfullyDecodesPayloadWithValidInput(t *testing.T) {
	t.Parallel()

	type testCase struct {
		encodedPayloadInput []byte
		delimiterLeft       string
		delimiterRight      string
		expectedOutput      string
	}

	tests := map[string]testCase{
		"small encoded json payload with default delimiters": {
			encodedPayloadInput: []byte(smallJSONPayloadEncodedWithDefaultDelimiters),
			delimiterLeft:       nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight:      nagios.DefaultASCII85EncodingDelimiterRight,
			expectedOutput:      smallJSONPayloadUnencoded,
		},
		"small encoded json payload with custom delimiters": {
			encodedPayloadInput: []byte(smallJSONPayloadEncodedWithCustomDelimiters),
			delimiterLeft:       customEncodingDelimiterLeft,
			delimiterRight:      customEncodingDelimiterRight,
			expectedOutput:      smallJSONPayloadUnencoded,
		},
		"small encoded json payload with no delimiters": {
			encodedPayloadInput: []byte(smallJSONPayloadEncodedWithNoDelimiters),
			delimiterLeft:       "",
			delimiterRight:      "",
			expectedOutput:      smallJSONPayloadUnencoded,
		},
		"small encoded plaintext payload with default delimiters": {
			encodedPayloadInput: []byte(smallPlaintextPayloadEncodedWithDefaultDelimiters),
			delimiterLeft:       nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight:      nagios.DefaultASCII85EncodingDelimiterRight,
			expectedOutput:      smallPlaintextPayloadUnencoded,
		},
		"small encoded plaintext payload with custom delimiters": {
			encodedPayloadInput: []byte(smallPlaintextPayloadEncodedWithCustomDelimiters),
			delimiterLeft:       customEncodingDelimiterLeft,
			delimiterRight:      customEncodingDelimiterRight,
			expectedOutput:      smallPlaintextPayloadUnencoded,
		},
		"small encoded plaintext payload with no delimiters": {
			encodedPayloadInput: []byte(smallPlaintextPayloadEncodedWithNoDelimiters),
			delimiterLeft:       "",
			delimiterRight:      "",
			expectedOutput:      smallPlaintextPayloadUnencoded,
		},
		"large encoded plaintext payload with default delimiters": {
			encodedPayloadInput: []byte(largePayloadEncodedWithDefaultDelimiters),
			delimiterLeft:       nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight:      nagios.DefaultASCII85EncodingDelimiterRight,
			expectedOutput:      largePayloadUnencoded,
		},
		"large encoded plaintext payload with custom delimiters": {
			encodedPayloadInput: []byte(largePayloadEncodedWithCustomDelimiters),
			delimiterLeft:       customEncodingDelimiterLeft,
			delimiterRight:      customEncodingDelimiterRight,
			expectedOutput:      largePayloadUnencoded,
		},
		"large encoded plaintext payload with no delimiters": {
			encodedPayloadInput: []byte(largePayloadEncodedWithNoDelimiters),
			delimiterLeft:       "",
			delimiterRight:      "",
			expectedOutput:      largePayloadUnencoded,
		},
	}

	for name, tt := range tests {
		// Guard against referencing the loop iterator variable directly.
		//
		// https://stackoverflow.com/questions/68559574/using-the-variable-on-range-scope-x-in-function-literal-scopelint
		// https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		tt := tt

		//
		t.Run(name, func(t *testing.T) {
			want := tt.expectedOutput

			result, err := nagios.DecodeASCII85Payload(
				tt.encodedPayloadInput,
				tt.delimiterLeft,
				tt.delimiterRight,
			)

			got := string(result)

			if err != nil {
				t.Fatalf("Failed to decode encoded payload: %v", err)
			} else {
				t.Logf("Successfully decoded encoded payload")
			}

			if d := cmp.Diff(want, got); d != "" {
				t.Errorf("(-want, +got)\n:%s", d)
			} else {
				t.Logf("OK: Encoded payload matches expected output.")
			}
		})
	}
}

func TestDecodeASCII85Payload_FailsToDecodePayloadWithInvalidInput(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input          []byte
		delimiterLeft  string
		delimiterRight string
		expectedErr    error
	}

	tests := map[string]testCase{
		"empty input with default delimiters": {
			input:          []byte(""),
			delimiterLeft:  nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight: nagios.DefaultASCII85EncodingDelimiterRight,
			expectedErr:    nagios.ErrMissingValue,
		},
		"empty input with custom delimiters": {
			input:          []byte(""),
			delimiterLeft:  customEncodingDelimiterLeft,
			delimiterRight: customEncodingDelimiterRight,
			expectedErr:    nagios.ErrMissingValue,
		},
		"empty input with no delimiters": {
			input:          []byte(""),
			delimiterLeft:  "",
			delimiterRight: "",
			expectedErr:    nagios.ErrMissingValue,
		},
		"unencoded input with default delimiters": {
			input:          []byte("!z!!!!!!!!!"), // borrowed from TestDecodeCorrupt testcase within ascii85_test.go
			delimiterLeft:  nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight: nagios.DefaultASCII85EncodingDelimiterRight,
			expectedErr:    ascii85.CorruptInputError(1), // illegal ascii85 data at input byte 1
		},
		"unencoded input with custom delimiters": {
			input:          []byte("!z!!!!!!!!!"), // borrowed from TestDecodeCorrupt testcase within ascii85_test.go
			delimiterLeft:  customEncodingDelimiterLeft,
			delimiterRight: customEncodingDelimiterRight,
			expectedErr:    ascii85.CorruptInputError(1), // illegal ascii85 data at input byte 1
		},
		"unencoded input with no delimiters": {
			input:          []byte("!z!!!!!!!!!"), // borrowed from TestDecodeCorrupt testcase within ascii85_test.go
			delimiterLeft:  "",
			delimiterRight: "",
			expectedErr:    ascii85.CorruptInputError(1), // illegal ascii85 data at input byte 1
		},
	}

	for name, tt := range tests {
		// Guard against referencing the loop iterator variable directly.
		//
		// https://stackoverflow.com/questions/68559574/using-the-variable-on-range-scope-x-in-function-literal-scopelint
		// https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		tt := tt

		//
		t.Run(name, func(t *testing.T) {

			// When successful a byte slice is returned, but when an error
			// occurs a nil value is returned. Because we're providing invalid
			// input, we expect  a nil value.
			var want []byte

			got, err := nagios.DecodeASCII85Payload(
				tt.input,
				tt.delimiterLeft,
				tt.delimiterRight,
			)

			if !errors.Is(err, tt.expectedErr) {
				t.Logf("Error is of type: %T", err)
				t.Fatalf("Decoding attempt did not fail with expected error: %v", err)
			} else {
				t.Logf("Error is of type: %T", err)
				t.Logf("Decoding attempt failed with expected error for invalid input: %v", err)
			}

			if d := cmp.Diff(want, got); d != "" {
				t.Logf("(-want, +got)\n:%s", d)
			} else {
				t.Logf("OK: Decoding invalid input produced no output.")
			}
		})
	}
}

func TestExtractEncodedASCII85Payload_SuccessfullyExtractsEncodedPayloadWithValidInput(t *testing.T) {
	t.Parallel()

	type testCase struct {
		inputWithEncodedPayload string
		delimiterLeft           string
		delimiterRight          string
		expectedOutput          string
	}

	// The extraction process retains the payload in encoded form but attempts
	// to remove any specified delimiters. When the input does not contain any
	// delimiters the extraction process is *VERY* unreliable.
	tests := map[string]testCase{
		"custom section header and large encoded payload and custom delimiters": {
			inputWithEncodedPayload: pluginOutputCustomSectionHeadersAndLargeEncodedPayloadWithCustomDelimiters,
			delimiterLeft:           customEncodingDelimiterLeft,
			delimiterRight:          customEncodingDelimiterRight,
			expectedOutput:          largePayloadEncodedWithNoDelimiters,
		},
		"custom section header and large encoded payload and default delimiters": {
			inputWithEncodedPayload: pluginOutputCustomSectionHeadersAndLargeEncodedPayloadWithDefaultDelimiters,
			delimiterLeft:           nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight:          nagios.DefaultASCII85EncodingDelimiterRight,
			expectedOutput:          largePayloadEncodedWithNoDelimiters,
		},
		"custom section header and small json encoded payload and custom delimiters": {
			inputWithEncodedPayload: pluginOutputCustomSectionHeadersAndSmallEncodedJSONPayloadWithCustomDelimiters,
			delimiterLeft:           customEncodingDelimiterLeft,
			delimiterRight:          customEncodingDelimiterRight,
			expectedOutput:          smallJSONPayloadEncodedWithNoDelimiters,
		},
		"custom section header and small json encoded payload and default delimiters": {
			inputWithEncodedPayload: pluginOutputCustomSectionHeadersAndSmallEncodedJSONPayloadWithDefaultDelimiters,
			delimiterLeft:           nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight:          nagios.DefaultASCII85EncodingDelimiterRight,
			expectedOutput:          smallJSONPayloadEncodedWithNoDelimiters,
		},
		"custom section header and small plaintext encoded payload and custom delimiters": {
			inputWithEncodedPayload: pluginOutputCustomSectionHeadersAndSmallEncodedPlaintextPayloadWithCustomDelimiters,
			delimiterLeft:           customEncodingDelimiterLeft,
			delimiterRight:          customEncodingDelimiterRight,
			expectedOutput:          smallPlaintextPayloadEncodedWithNoDelimiters,
		},
		"custom section header and small plaintext encoded payload and default delimiters": {
			inputWithEncodedPayload: pluginOutputCustomSectionHeadersAndSmallEncodedPlaintextPayloadWithDefaultDelimiters,
			delimiterLeft:           nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight:          nagios.DefaultASCII85EncodingDelimiterRight,
			expectedOutput:          smallPlaintextPayloadEncodedWithNoDelimiters,
		},
		"default section header and small encoded json payload and custom delimiters": {
			inputWithEncodedPayload: pluginOutputDefaultSectionHeadersAndSmallEncodedJSONPayloadWithCustomDelimiters,
			delimiterLeft:           customEncodingDelimiterLeft,
			delimiterRight:          customEncodingDelimiterRight,
			expectedOutput:          smallJSONPayloadEncodedWithNoDelimiters,
		},
		"default section header and small encoded json payload and default delimiters": {
			inputWithEncodedPayload: pluginOutputDefaultSectionHeadersAndSmallEncodedJSONPayloadWithDefaultDelimiters,
			delimiterLeft:           nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight:          nagios.DefaultASCII85EncodingDelimiterRight,
			expectedOutput:          smallJSONPayloadEncodedWithNoDelimiters,
		},
		"default section header and small encoded plaintext payload and custom delimiters": {
			inputWithEncodedPayload: pluginOutputDefaultSectionHeadersAndSmallEncodedPlaintextPayloadWithCustomDelimiters,
			delimiterLeft:           customEncodingDelimiterLeft,
			delimiterRight:          customEncodingDelimiterRight,
			expectedOutput:          smallPlaintextPayloadEncodedWithNoDelimiters,
		},
		"default section header and small encoded plaintext payload and default delimiters": {
			inputWithEncodedPayload: pluginOutputDefaultSectionHeadersAndSmallEncodedPlaintextPayloadWithDefaultDelimiters,
			delimiterLeft:           nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight:          nagios.DefaultASCII85EncodingDelimiterRight,
			expectedOutput:          smallPlaintextPayloadEncodedWithNoDelimiters,
		},
	}

	for name, tt := range tests {
		// Guard against referencing the loop iterator variable directly.
		//
		// https://stackoverflow.com/questions/68559574/using-the-variable-on-range-scope-x-in-function-literal-scopelint
		// https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		tt := tt

		//
		t.Run(name, func(t *testing.T) {
			want := tt.expectedOutput

			got, err := nagios.ExtractEncodedASCII85Payload(
				tt.inputWithEncodedPayload,
				"", // custom regex option; we will use default by not providing value here
				tt.delimiterLeft,
				tt.delimiterRight,
			)

			if err != nil {
				t.Fatalf("Failed to extract encoded payload: %v", err)
			} else {
				t.Logf("Successfully extracted encoded payload")
			}

			if d := cmp.Diff(want, got); d != "" {
				t.Errorf("(-want, +got)\n:%s", d)
			} else {
				t.Logf("OK: Extracted encoded payload matches expected output.")
			}
		})
	}
}

func TestExtractEncodedASCII85Payload_FailsToExtractEncodedPayloadWithInvalidInput(t *testing.T) {
	t.Parallel()

	type testCase struct {
		inputWithEncodedPayload string
		delimiterLeft           string
		delimiterRight          string
		expectedOutput          string
	}

	// Not providing delimiters when attempting to extract payloads is *VERY*
	// unreliable.
	//
	// Most scenarios fail with the extraction attempt matching the first line
	// of the plugin output (which is input in our case):
	//
	// `OK: Datastore HUSVM-DC1-`
	//
	// instead of matching the intended payload towards the bottom of the
	// sample input. For our purposes, we consider this to be a known problem;
	// if encoding the payload *without* delimiters and then later attempting
	// to extract the payload, failure is the only "reliable" outcome.
	tests := map[string]testCase{
		"empty input and custom delimiters": {
			inputWithEncodedPayload: "",
			delimiterLeft:           customEncodingDelimiterLeft,
			delimiterRight:          customEncodingDelimiterRight,
			expectedOutput:          "",
		},
		"empty input and default delimiters": {
			inputWithEncodedPayload: "",
			delimiterLeft:           nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight:          nagios.DefaultASCII85EncodingDelimiterRight,
			expectedOutput:          "",
		},
		"empty input and no delimiters": {
			inputWithEncodedPayload: "",
			delimiterLeft:           "",
			delimiterRight:          "",
			expectedOutput:          "",
		},
		"custom section header and large encoded payload and no delimiters": {
			inputWithEncodedPayload: pluginOutputCustomSectionHeadersAndLargeEncodedPayloadWithNoDelimiters,
			delimiterLeft:           "",
			delimiterRight:          "",
			expectedOutput:          "OK: Datastore HUSVM-DC1-",
		},
		"custom section header and small json encoded payload and no delimiters": {
			inputWithEncodedPayload: pluginOutputCustomSectionHeadersAndSmallEncodedJSONPayloadWithNoDelimiters,
			delimiterLeft:           "",
			delimiterRight:          "",
			expectedOutput:          "OK: Datastore HUSVM-DC1-",
		},
		"custom section header and small plaintext encoded payload and no delimiters": {
			inputWithEncodedPayload: pluginOutputCustomSectionHeadersAndSmallEncodedPlaintextPayloadWithNoDelimiters,
			delimiterLeft:           "",
			delimiterRight:          "",
			expectedOutput:          "OK: Datastore HUSVM-DC1-",
		},
		"default section header and small encoded json payload and no delimiters": {
			inputWithEncodedPayload: pluginOutputDefaultSectionHeadersAndSmallEncodedJSONPayloadWithNoDelimiters,
			delimiterLeft:           "",
			delimiterRight:          "",
			expectedOutput:          "OK: Datastore HUSVM-DC1-",
		},
		"default section header and large encoded payload and no delimiters": {
			inputWithEncodedPayload: pluginOutputDefaultSectionHeadersAndLargeEncodedPayloadWithNoDelimiters,
			delimiterLeft:           "",
			delimiterRight:          "",
			expectedOutput:          "OK: Datastore HUSVM-DC1-",
		},
		"default section header and small encoded plaintext payload and no delimiters": {
			inputWithEncodedPayload: pluginOutputDefaultSectionHeadersAndSmallEncodedPlaintextPayloadWithNoDelimiters,
			delimiterLeft:           "",
			delimiterRight:          "",
			expectedOutput:          "OK: Datastore HUSVM-DC1-",
		},
	}

	for name, tt := range tests {
		// Guard against referencing the loop iterator variable directly.
		//
		// https://stackoverflow.com/questions/68559574/using-the-variable-on-range-scope-x-in-function-literal-scopelint
		// https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		tt := tt

		t.Run(name, func(t *testing.T) {
			want := tt.expectedOutput

			got, err := nagios.ExtractEncodedASCII85Payload(
				tt.inputWithEncodedPayload,
				"",
				tt.delimiterLeft,
				tt.delimiterRight,
			)

			// If we're expecting output, then we're not expecting an error.
			switch {
			case len(tt.expectedOutput) > 0:
				if d := cmp.Diff(want, got); d != "" {
					t.Errorf("(-want, +got)\n:%s", d)
				} else {
					t.Logf("OK: Encoded payload matches expected output.")
				}

			// If we're not expecting a match, then we're expecting an error.
			default:
				if err == nil {
					t.Errorf("Extraction attempt did not fail with expected result")
				} else {
					t.Logf("Extraction attempt failed as expected: %v", err)
				}
			}

		})
	}
}

func TestExtractAndDecodeASCII85Payload_SuccessfullyExtractsAndDecodesPayloadWithValidInput(t *testing.T) {
	t.Parallel()

	type testCase struct {
		inputWithEncodedPayload string
		delimiterLeft           string
		delimiterRight          string
		expectedOutput          string
	}

	// The extraction & decoding process attempts to remove any specified
	// delimiters. When the input does not contain any delimiters the
	// extraction process is *VERY* unreliable.
	tests := map[string]testCase{
		"custom section header and large encoded payload and custom delimiters": {
			inputWithEncodedPayload: pluginOutputCustomSectionHeadersAndLargeEncodedPayloadWithCustomDelimiters,
			delimiterLeft:           customEncodingDelimiterLeft,
			delimiterRight:          customEncodingDelimiterRight,
			expectedOutput:          largePayloadUnencoded,
		},
		"custom section header and large encoded payload and default delimiters": {
			inputWithEncodedPayload: pluginOutputCustomSectionHeadersAndLargeEncodedPayloadWithDefaultDelimiters,
			delimiterLeft:           nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight:          nagios.DefaultASCII85EncodingDelimiterRight,
			expectedOutput:          largePayloadUnencoded,
		},
		"custom section header and small json encoded payload and custom delimiters": {
			inputWithEncodedPayload: pluginOutputCustomSectionHeadersAndSmallEncodedJSONPayloadWithCustomDelimiters,
			delimiterLeft:           customEncodingDelimiterLeft,
			delimiterRight:          customEncodingDelimiterRight,
			expectedOutput:          smallJSONPayloadUnencoded,
		},
		"custom section header and small json encoded payload and default delimiters": {
			inputWithEncodedPayload: pluginOutputCustomSectionHeadersAndSmallEncodedJSONPayloadWithDefaultDelimiters,
			delimiterLeft:           nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight:          nagios.DefaultASCII85EncodingDelimiterRight,
			expectedOutput:          smallJSONPayloadUnencoded,
		},
		"custom section header and small plaintext encoded payload and custom delimiters": {
			inputWithEncodedPayload: pluginOutputCustomSectionHeadersAndSmallEncodedPlaintextPayloadWithCustomDelimiters,
			delimiterLeft:           customEncodingDelimiterLeft,
			delimiterRight:          customEncodingDelimiterRight,
			expectedOutput:          smallPlaintextPayloadUnencoded,
		},
		"custom section header and small plaintext encoded payload and default delimiters": {
			inputWithEncodedPayload: pluginOutputCustomSectionHeadersAndSmallEncodedPlaintextPayloadWithDefaultDelimiters,
			delimiterLeft:           nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight:          nagios.DefaultASCII85EncodingDelimiterRight,
			expectedOutput:          smallPlaintextPayloadUnencoded,
		},
		"default section header and small encoded json payload and custom delimiters": {
			inputWithEncodedPayload: pluginOutputDefaultSectionHeadersAndSmallEncodedJSONPayloadWithCustomDelimiters,
			delimiterLeft:           customEncodingDelimiterLeft,
			delimiterRight:          customEncodingDelimiterRight,
			expectedOutput:          smallJSONPayloadUnencoded,
		},
		"default section header and small encoded json payload and default delimiters": {
			inputWithEncodedPayload: pluginOutputDefaultSectionHeadersAndSmallEncodedJSONPayloadWithDefaultDelimiters,
			delimiterLeft:           nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight:          nagios.DefaultASCII85EncodingDelimiterRight,
			expectedOutput:          smallJSONPayloadUnencoded,
		},
		"default section header and small encoded plaintext payload and custom delimiters": {
			inputWithEncodedPayload: pluginOutputDefaultSectionHeadersAndSmallEncodedPlaintextPayloadWithCustomDelimiters,
			delimiterLeft:           customEncodingDelimiterLeft,
			delimiterRight:          customEncodingDelimiterRight,
			expectedOutput:          smallPlaintextPayloadUnencoded,
		},
		"default section header and small encoded plaintext payload and default delimiters": {
			inputWithEncodedPayload: pluginOutputDefaultSectionHeadersAndSmallEncodedPlaintextPayloadWithDefaultDelimiters,
			delimiterLeft:           nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight:          nagios.DefaultASCII85EncodingDelimiterRight,
			expectedOutput:          smallPlaintextPayloadUnencoded,
		},
	}

	for name, tt := range tests {
		// Guard against referencing the loop iterator variable directly.
		//
		// https://stackoverflow.com/questions/68559574/using-the-variable-on-range-scope-x-in-function-literal-scopelint
		// https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		tt := tt

		//
		t.Run(name, func(t *testing.T) {
			want := tt.expectedOutput

			got, err := nagios.ExtractAndDecodeASCII85Payload(
				tt.inputWithEncodedPayload,
				"", // custom regex option; we will use default by not providing value here
				tt.delimiterLeft,
				tt.delimiterRight,
			)

			if err != nil {
				t.Fatalf("Failed to extract and decode payload: %v", err)
			} else {
				t.Logf("Successfully extracted and decoded payload")
			}

			if d := cmp.Diff(want, got); d != "" {
				t.Errorf("(-want, +got)\n:%s", d)
			} else {
				t.Logf("OK: Extracted & decoded payload matches expected output.")
			}
		})
	}
}

func TestExtractAndDecodeASCII85Payload_FailsToExtractAndDecodePayloadWithInvalidInput(t *testing.T) {
	t.Parallel()

	type testCase struct {
		inputWithEncodedPayload string
		delimiterLeft           string
		delimiterRight          string
		chosenRegex             string
		expectedOutput          string
	}

	/*
		The ExtractAndDecodeASCII85Payload function (just like the
		ExtractEncodedASCII85Payload function) produces (reliably) undesirable
		results when delimiters are not used.

		When we attempt to decode `OK: Datastore HUSVM-DC1-` as Ascii85
		encoded text it generates "noise".

		encodedPayload, err := ExtractEncodedASCII85Payload(text, customRegex, leftDelimiter, rightDelimiter)

			contents of `encodedPayload`:
				OK: Datastore HUSVM-DC1-

		decodedPayload, err := decodeASCII85([]byte(encodedPayload))

			contents of `decodedPayload` (byte slice):
				[144 172 56 32 4 159 230 194 254 135 145 26 166 133 39 206 50]

			when attempting to convert to a string via `string(decodedPayload)`
			for troubleshooting purposes:
				��8 ������␦��'�2

		when the byte slice "noise" (undesirable decoding) is converted to a
		string and returned:

		return string(decodedPayload), nil

			this value is what we're left with:
				\x90\xac8 \x04\x9f\xe6\xc2\xfe\x87\x91\x1a\xa6\x85'\xce2

		This in no way represents the encoded payload nor the original
		extracted & decoded payload we would expect to see.
	*/

	// Not providing delimiters when attempting to extract payloads is *VERY*
	// unreliable.
	//
	// Most scenarios fail with the extraction attempt matching the first line
	// of the plugin output (which is input in our case):
	//
	// `OK: Datastore HUSVM-DC1-`
	//
	// instead of matching the intended payload towards the bottom of the
	// sample input. For our purposes, we consider this to be a known problem;
	// if encoding the payload *without* delimiters and then later attempting
	// to extract the payload, failure is the only "reliable" outcome.
	tests := map[string]testCase{
		"empty input and custom delimiters and default regex": {
			inputWithEncodedPayload: "",
			delimiterLeft:           customEncodingDelimiterLeft,
			delimiterRight:          customEncodingDelimiterRight,
			chosenRegex:             nagios.DefaultASCII85EncodingPatternRegex,
			expectedOutput:          "",
		},
		"empty input and default delimiters and default regex": {
			inputWithEncodedPayload: "",
			delimiterLeft:           nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight:          nagios.DefaultASCII85EncodingDelimiterRight,
			chosenRegex:             nagios.DefaultASCII85EncodingPatternRegex,
			expectedOutput:          "",
		},
		"empty input and no delimiters and default regex": {
			inputWithEncodedPayload: "",
			delimiterLeft:           "",
			delimiterRight:          "",
			chosenRegex:             nagios.DefaultASCII85EncodingPatternRegex,
			expectedOutput:          "",
		},
		"custom section header and large encoded payload and no delimiters and default regex": {
			inputWithEncodedPayload: pluginOutputCustomSectionHeadersAndLargeEncodedPayloadWithNoDelimiters,
			delimiterLeft:           "",
			delimiterRight:          "",
			chosenRegex:             nagios.DefaultASCII85EncodingPatternRegex,
			expectedOutput:          "\x90\xac8 \x04\x9f\xe6\xc2\xfe\x87\x91\x1a\xa6\x85'\xce2",
		},
		"custom section header and small json encoded payload and no delimiters and default regex": {
			inputWithEncodedPayload: pluginOutputCustomSectionHeadersAndSmallEncodedJSONPayloadWithNoDelimiters,
			delimiterLeft:           "",
			delimiterRight:          "",
			chosenRegex:             nagios.DefaultASCII85EncodingPatternRegex,
			expectedOutput:          "\x90\xac8 \x04\x9f\xe6\xc2\xfe\x87\x91\x1a\xa6\x85'\xce2",
		},
		"custom section header and small plaintext encoded payload and no delimiters and default regex": {
			inputWithEncodedPayload: pluginOutputCustomSectionHeadersAndSmallEncodedPlaintextPayloadWithNoDelimiters,
			delimiterLeft:           "",
			delimiterRight:          "",
			chosenRegex:             nagios.DefaultASCII85EncodingPatternRegex,
			expectedOutput:          "\x90\xac8 \x04\x9f\xe6\xc2\xfe\x87\x91\x1a\xa6\x85'\xce2",
		},
		"default section header and small encoded json payload and no delimiters and default regex": {
			inputWithEncodedPayload: pluginOutputDefaultSectionHeadersAndSmallEncodedJSONPayloadWithNoDelimiters,
			delimiterLeft:           "",
			delimiterRight:          "",
			chosenRegex:             nagios.DefaultASCII85EncodingPatternRegex,
			expectedOutput:          "\x90\xac8 \x04\x9f\xe6\xc2\xfe\x87\x91\x1a\xa6\x85'\xce2",
		},
		"default section header and small encoded plaintext payload and no delimiters and default regex": {
			inputWithEncodedPayload: pluginOutputDefaultSectionHeadersAndSmallEncodedPlaintextPayloadWithNoDelimiters,
			delimiterLeft:           "",
			delimiterRight:          "",
			chosenRegex:             nagios.DefaultASCII85EncodingPatternRegex,
			expectedOutput:          "\x90\xac8 \x04\x9f\xe6\xc2\xfe\x87\x91\x1a\xa6\x85'\xce2",
		},
		"custom section header and small json encoded payload and default delimiters and custom regex": {
			inputWithEncodedPayload: pluginOutputCustomSectionHeadersAndSmallEncodedJSONPayloadWithDefaultDelimiters,
			delimiterLeft:           nagios.DefaultASCII85EncodingDelimiterLeft,
			delimiterRight:          nagios.DefaultASCII85EncodingDelimiterRight,
			chosenRegex:             `[[`,
			expectedOutput:          "",
		},
		"default section header and small encoded plaintext payload and custom delimiters and custom regex": {
			inputWithEncodedPayload: pluginOutputDefaultSectionHeadersAndSmallEncodedPlaintextPayloadWithCustomDelimiters,
			delimiterLeft:           customEncodingDelimiterLeft,
			delimiterRight:          customEncodingDelimiterRight,
			chosenRegex:             `[][][][][][][]`,
			expectedOutput:          "",
		},
		"custom section header and large encoded payload and default delimiters and custom regex": {
			inputWithEncodedPayload: pluginOutputCustomSectionHeadersAndLargeEncodedPayloadWithDefaultDelimiters,
			delimiterLeft:           "",
			delimiterRight:          "",
			chosenRegex:             `[[[`,
			expectedOutput:          "",
		},
	}

	for name, tt := range tests {
		// Guard against referencing the loop iterator variable directly.
		//
		// https://stackoverflow.com/questions/68559574/using-the-variable-on-range-scope-x-in-function-literal-scopelint
		// https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		tt := tt

		t.Run(name, func(t *testing.T) {
			want := tt.expectedOutput

			got, err := nagios.ExtractAndDecodeASCII85Payload(
				tt.inputWithEncodedPayload,
				tt.chosenRegex,
				tt.delimiterLeft,
				tt.delimiterRight,
			)

			// If we're expecting output, then we're not expecting an error.
			switch {
			case len(tt.expectedOutput) > 0:
				if d := cmp.Diff(want, got); d != "" {
					t.Errorf("(-want, +got)\n:%s", d)
				} else {
					t.Logf("OK: Decoded payload matches expected output.")
				}

			// If we're not expecting a match, then we're expecting an error.
			default:
				if err == nil {
					t.Errorf("Extraction and decode attempt did not fail with expected result")
				} else {
					t.Logf("Extraction and decode attempt failed as expected: %v", err)
				}
			}

		})
	}
}

func pluginOutputWithLongServiceOutputSetup(t *testing.T, plugin *nagios.Plugin) {
	t.Helper()

	// os.Exit calls break tests. Potentially duplicated by caller, but
	// effectively a NOOP if repeated so not an issue.
	plugin.SkipOSExit()

	//nolint:goconst
	plugin.ServiceOutput =
		"OK: Datastore HUSVM-DC1-vol6 space usage (0 VMs)" +
			" is 0.01% of 18.0TB with 18.0TB remaining" +
			" [WARNING: 90% , CRITICAL: 95%]"

	plugin.CriticalThreshold = fmt.Sprintf(
		"%d%% datastore usage",
		95,
	)

	plugin.WarningThreshold = fmt.Sprintf(
		"%d%% datastore usage",
		90,
	)

	var longServiceOutputBuffer strings.Builder

	_, _ = fmt.Fprintf(
		&longServiceOutputBuffer,
		"Datastore Space Summary:%s%s"+
			"* Name: %s%s"+
			"* Space Used: %v (%.2f%%)%s"+
			"* Space Remaining: %v (%.2f%%)%s"+
			"* VMs: %v %s%s",
		nagios.CheckOutputEOL,
		nagios.CheckOutputEOL,
		"HUSVM-DC1-vol6",
		nagios.CheckOutputEOL,
		"2.3GB",
		0.01,
		nagios.CheckOutputEOL,
		"18.0TB",
		99.99,
		nagios.CheckOutputEOL,
		0,
		nagios.CheckOutputEOL,
		nagios.CheckOutputEOL,
	)

	_, _ = fmt.Fprintf(
		&longServiceOutputBuffer,
		"%s---%s%s",
		nagios.CheckOutputEOL,
		nagios.CheckOutputEOL,
		nagios.CheckOutputEOL,
	)

	_, _ = fmt.Fprintf(
		&longServiceOutputBuffer,
		"* vSphere environment: %s%s",
		"https://vc1.example.com:443/sdk",
		nagios.CheckOutputEOL,
	)

	_, _ = fmt.Fprintf(
		&longServiceOutputBuffer,
		"* Plugin User Agent: %s%s",
		"check-vmware/v0.30.6-0-g25fdcdc",
		nagios.CheckOutputEOL,
	)

	plugin.LongServiceOutput += longServiceOutputBuffer.String()
}

func pluginOutputWithLongServiceOutputMetrics(t *testing.T, plugin *nagios.Plugin) {
	t.Helper()

	// os.Exit calls break tests. Potentially duplicated by caller, but
	// effectively a NOOP if repeated so not an issue.
	plugin.SkipOSExit()

	pd := nagios.PerformanceData{
		Label: "time",
		Value: "874ms",
	}

	if err := plugin.AddPerfData(false, pd); err != nil {
		t.Errorf("failed to add performance data: %v", err)
	}
}
