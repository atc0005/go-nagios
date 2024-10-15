// Copyright 2020 Adam Chalkley
//
// https://github.com/atc0005/go-nagios
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

// Package nagios provides test coverage for unexported package functionality.
//
//nolint:dupl,gocognit // ignore "lines are duplicate of" and function complexity
package nagios

import (
	_ "embed"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// The specific format used by the test input files is VERY specific; trailing
// space + newline patterns are intentional. Because "format on save" editor
// functionality easily breaks this input it is stored in separate files to
// reduce test breakage due to editors "helping".
var (
	//go:embed testdata/payload/small_json_payload_unencoded.txt
	smallJSONPayloadUnencoded string

	// Earlier prototyping found that the stream encoding/decoding process
	// did not retain exclamation marks. I've not dug deep enough to
	// determine the root cause, but have observed that using the Decode
	// and Encode functions work reliably. Because a later refactoring
	// might switch back to using stream processing we explicitly test
	// using an exclamation point to guard against future breakage.
	//
	//go:embed testdata/payload/small_plaintext_payload_unencoded.txt
	smallPlaintextPayloadUnencoded string
)

const (
// Earlier prototyping found that the stream encoding/decoding process
// did not retain exclamation marks. I've not dug deep enough to
// determine the root cause, but have observed that using the Decode
// and Encode functions work reliably. Because a later refactoring
// might switch back to using stream processing we explicitly test
// using an exclamation point to guard against future breakage.
// smallPlaintextPayloadUnencoded string = "Hello, World!"
)

// TestServiceOutputIsNotInterpolated is intended to prevent further
// regressions of formatting being applied to literal/preformatted Service
// Output (aka, "one-line summary" output).
//
// See also:
//
// - https://github.com/atc0005/go-nagios/issues/139
// - https://github.com/atc0005/go-nagios/issues/58
func TestServiceOutputIsNotInterpolated(t *testing.T) {
	t.Parallel()

	// Setup Plugin value manually. This approach does not provide the
	// default time metric that would be provided when using the Plugin
	// constructor.
	var plugin = Plugin{
		LastError:      nil,
		ExitStatusCode: StateOKExitCode,
	}

	var output strings.Builder

	// If passed through fmt.Printf, fmt.Fprintf or fmt.Sprintf the '% o'
	// pattern is treated as a base 8 integer formatting verb. If passed
	// through fmt.Print or fmt.Fprint the pattern is ignored (as intended).
	testInput := "OK: Datastore HUSVM-DC1-vol6 space usage (0 VMs)" +
		" is 0.01% of 18.0TB with 18.0TB remaining" +
		" [WARNING: 90% , CRITICAL: 95%]"

	// The input from client code is expected to be passed as-is, without
	// formatting or interpretation of any kind.
	want := testInput
	plugin.ServiceOutput = testInput

	plugin.handleServiceOutputSection(&output)

	got := output.String()

	if want != got {
		t.Errorf("\nwant %q\ngot %q", want, got)
	}
}

// TestPerformanceDataIsNotDuplicated asserts that duplicate Performance Data
// metrics are not collected.
//
// See also:
//
// - https://github.com/atc0005/go-nagios/issues/157
func TestPerformanceDataIsNotDuplicated(t *testing.T) {
	t.Parallel()

	// Setup Plugin value manually. This approach does not provide the
	// default time metric that would be provided when using the Plugin
	// constructor.
	var plugin = Plugin{
		LastError:      nil,
		ExitStatusCode: StateOKExitCode,
	}

	// Collection of performance data with duplicate entries.
	pd := []PerformanceData{
		{ // first performance data entry
			Label: "test1",
			Value: "1",
		},
		{ // repeated
			Label: "test1",
			Value: "1",
		},
		{ // repeated with Label in all upper case
			Label: "TEST1",
			Value: "1",
		},
		{ // repeated with Label in mixed case
			Label: "teST1",
			Value: "1",
		},
		{ // first non-duplicate Label
			Label: "test2",
			Value: "1",
		},
	}

	if err := plugin.AddPerfData(false, pd...); err != nil {
		t.Errorf("failed to add initial performance data: %v", err)
	}

	// Standalone performance data metric, duplicate of entry from first
	// collection.
	pd = append(pd, PerformanceData{
		Label: "test1",
		Value: "1",
	})

	if err := plugin.AddPerfData(false, pd...); err != nil {
		t.Errorf("failed to append additional performance data: %v", err)
	}

	// Two unique labels, so should be just two performance data metrics.
	want := 2
	got := len(plugin.perfData)

	if got != want {
		t.Errorf(
			"\nwant %d performance data metrics\ngot %d performance data metrics",
			want,
			got,
		)
	} else {
		t.Logf(
			"OK: \nwant %d performance data metrics\ngot %d performance data metrics",
			want,
			got,
		)
	}

}

// TestEmptyServiceOutputProducesNoOutput asserts that an empty ServiceOutput
// field produces no output.
func TestEmptyServiceOutputProducesNoOutput(t *testing.T) {
	t.Parallel()

	// Setup Plugin value manually. This approach does not provide the
	// default time metric that would be provided when using the Plugin
	// constructor.
	var plugin = Plugin{
		LastError:      nil,
		ExitStatusCode: StateOKExitCode,
	}

	var outputBuffer strings.Builder

	// Explicitly indicate that the field is empty (default/zero value).
	plugin.ServiceOutput = ""

	plugin.handleServiceOutputSection(&outputBuffer)

	want := ""
	got := outputBuffer.String()

	if want != got {
		t.Errorf("\nwant %q\ngot %q", want, got)
	} else {
		t.Logf("OK: Empty service output field produces no output.")
	}

}

// TestEmptyEncodedPayloadWithDefaultDelimitersProducesNoOutput asserts that
// an empty payload buffer with default delimiters produces no output.
func TestEmptyEncodedPayloadWithDefaultDelimitersProducesNoOutput(t *testing.T) {
	t.Parallel()

	// Setup Plugin value manually. This approach does not provide the
	// default time metric that would be provided when using the Plugin
	// constructor.
	var plugin = Plugin{
		LastError:      nil,
		ExitStatusCode: StateOKExitCode,
	}

	var outputBuffer strings.Builder

	// At this point the encoded payload buffer is empty, so any attempt to
	// process it should result in no output.
	plugin.handleEncodedPayload(&outputBuffer)

	want := ""
	got := outputBuffer.String()

	if want != got {
		t.Errorf("\nwant %q\ngot %q", want, got)
	} else {
		t.Logf("OK: Empty payload buffer produces no output.")
	}
}

// TestEmptyEncodedPayloadWithCustomDelimitersProducesNoOutput asserts that an
// empty payload buffer with custom delimiters set produces no output.
func TestEmptyEncodedPayloadWithCustomDelimitersProducesNoOutput(t *testing.T) {
	t.Parallel()

	// Setup Plugin value manually. This approach does not provide the
	// default time metric that would be provided when using the Plugin
	// constructor.
	var plugin = Plugin{
		LastError:      nil,
		ExitStatusCode: StateOKExitCode,
	}

	var outputBuffer strings.Builder

	delimiterLeft := DefaultASCII85EncodingDelimiterLeft
	delimiterRight := DefaultASCII85EncodingDelimiterRight

	plugin.encodedPayloadDelimiterLeft = &delimiterLeft
	plugin.encodedPayloadDelimiterRight = &delimiterRight

	// At this point the encoded payload buffer is empty, so any attempt to
	// process it should result in no output.
	plugin.handleEncodedPayload(&outputBuffer)

	want := ""
	got := outputBuffer.String()

	if want != got {
		t.Errorf("\nwant %q\ngot %q", want, got)
	} else {
		t.Logf("OK: Empty payload buffer with custom delimiters produces no output.")
	}
}

// TestSetPayloadString_SetsInputSuccessfullyWhenCalledOnce asserts that a
// payload buffer populated via the `SetPayloadString` method (overwrite
// behavior) with non-repeating valid input produces valid output.
func TestSetPayloadString_SetsInputSuccessfullyWhenCalledOnce(t *testing.T) {
	t.Parallel()

	plugin := NewPlugin()

	tests := map[string]struct {
		// This represents data as given before it is written to the
		// payload buffer (unencoded).
		input string
	}{
		"simple JSON value": {
			input: smallJSONPayloadUnencoded,
		},
		"simple text value": {
			input: smallPlaintextPayloadUnencoded,
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
			t.Logf("Evaluating input %q", tt.input)

			plugin.encodedPayloadBuffer.Reset()

			want := tt.input
			written, err := plugin.SetPayloadString(tt.input)

			if err != nil {
				t.Fatalf("Failed to set payload buffer to given input: %v", err)
			} else {
				t.Logf("Successfully set payload buffer to %d bytes given input", written)
			}

			got := plugin.encodedPayloadBuffer.String()

			if d := cmp.Diff(want, got); d != "" {
				t.Errorf("(-want, +got)\n:%s", d)
			} else {
				t.Logf("OK: Payload buffer matches given input.")
			}
		})
	}
}

// TestSetPayloadString_SetsInputSuccessfullyWhenCalledMultipleTimes asserts
// that a payload buffer populated via the `SetPayloadString` method
// (overwrite behavior) with repeating valid input produces valid output.
func TestSetPayloadString_SetsInputSuccessfullyWhenCalledMultipleTimes(t *testing.T) {
	t.Parallel()

	plugin := NewPlugin()

	tests := map[string]struct {
		// This represents data as given before it is written to the
		// payload buffer (unencoded).
		input string
	}{
		"simple JSON value": {
			input: smallJSONPayloadUnencoded,
		},
		"simple text value": {
			input: smallPlaintextPayloadUnencoded,
		},
	}

	for name, tt := range tests {
		// Guard against referencing the loop iterator variable directly.
		//
		// https://stackoverflow.com/questions/68559574/using-the-variable-on-range-scope-x-in-function-literal-scopelint
		// https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		tt := tt

		t.Run(name, func(t *testing.T) {
			t.Logf("Evaluating input %q", tt.input)

			plugin.encodedPayloadBuffer.Reset()

			repeat := 2
			for i := 0; i < repeat+1; i++ {
				written, err := plugin.SetPayloadString(tt.input)

				if err != nil {
					t.Fatalf("Failed to set payload buffer to given input: %v", err)
				} else {
					t.Logf("Successfully set payload buffer to %d bytes given input", written)
				}
			}

			// Repeat function call (with overwrite behavior) should produce
			// non-repeating output.
			want := tt.input

			got := plugin.encodedPayloadBuffer.String()

			if d := cmp.Diff(want, got); d != "" {
				t.Errorf("(-want, +got)\n:%s", d)
			} else {
				t.Logf("OK: Payload buffer matches given input.")
			}
		})
	}
}

// TestSetPayloadBytes_SetsInputSuccessfullyWhenCalledOnce asserts that a
// payload buffer populated via the `SetPayloadBytes` method (overwrite
// behavior) with non-repeating valid input produces valid output.
func TestSetPayloadBytes_SetsInputSuccessfullyWhenCalledOnce(t *testing.T) {
	t.Parallel()

	plugin := NewPlugin()

	tests := map[string]struct {
		// This represents data as given before it is written to the
		// payload buffer (unencoded).
		input []byte
	}{
		"simple JSON value": {
			input: []byte(smallJSONPayloadUnencoded),
		},
		"simple text value": {
			input: []byte(smallPlaintextPayloadUnencoded),
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
			t.Logf("Evaluating input %q", tt.input)

			plugin.encodedPayloadBuffer.Reset()

			want := string(tt.input)
			written, err := plugin.SetPayloadBytes(tt.input)

			if err != nil {
				t.Fatalf("Failed to set payload buffer to given input: %v", err)
			} else {
				t.Logf("Successfully set payload buffer to %d bytes given input", written)
			}

			got := plugin.encodedPayloadBuffer.String()

			if d := cmp.Diff(want, got); d != "" {
				t.Errorf("(-want, +got)\n:%s", d)
			} else {
				t.Logf("OK: Payload buffer matches given input.")
			}
		})
	}
}

// TestSetPayloadBytes_SetsInputSuccessfullyWhenCalledMultipleTimes asserts that a
// payload buffer populated via the `SetPayloadBytes` method (overwrite
// behavior) with repeating valid input produces valid output.
func TestSetPayloadBytes_SetsInputSuccessfullyWhenCalledMultipleTimes(t *testing.T) {
	t.Parallel()

	plugin := NewPlugin()

	tests := map[string]struct {
		// This represents data as given before it is written to the
		// payload buffer (unencoded).
		input []byte
	}{
		"simple JSON value": {
			input: []byte(smallJSONPayloadUnencoded),
		},
		"simple text value": {
			input: []byte(smallPlaintextPayloadUnencoded),
		},
	}

	for name, tt := range tests {
		// Guard against referencing the loop iterator variable directly.
		//
		// https://stackoverflow.com/questions/68559574/using-the-variable-on-range-scope-x-in-function-literal-scopelint
		// https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		tt := tt

		t.Run(name, func(t *testing.T) {
			t.Logf("Evaluating input %q", tt.input)

			plugin.encodedPayloadBuffer.Reset()

			repeat := 2
			for i := 0; i < repeat+1; i++ {
				written, err := plugin.SetPayloadBytes(tt.input)

				if err != nil {
					t.Fatalf("Failed to set payload buffer to given input: %v", err)
				} else {
					t.Logf("Successfully set payload buffer to %d bytes given input", written)
				}
			}

			// Repeat function call (with overwrite behavior) should produce
			// non-repeating output.
			want := string(tt.input)

			got := plugin.encodedPayloadBuffer.String()

			if d := cmp.Diff(want, got); d != "" {
				t.Errorf("(-want, +got)\n:%s", d)
			} else {
				t.Logf("OK: Payload buffer matches given input.")
			}
		})
	}
}

// TestAddPayloadString_AppendsInputSuccessfullyWhenCalledOnce asserts that a
// payload buffer populated via the `AddPayloadString` method with
// non-repeating valid input produces valid output.
func TestAddPayloadString_AppendsInputSuccessfullyWhenCalledOnce(t *testing.T) {
	t.Parallel()

	plugin := NewPlugin()

	tests := map[string]struct {
		// This represents data as given before it is written to the
		// payload buffer (unencoded).
		input string
	}{
		"simple JSON value": {
			input: smallJSONPayloadUnencoded,
		},
		"simple text value": {
			input: smallPlaintextPayloadUnencoded,
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
			t.Logf("Evaluating input %q", tt.input)

			plugin.encodedPayloadBuffer.Reset()

			want := tt.input
			written, err := plugin.AddPayloadString(tt.input)

			if err != nil {
				t.Fatalf("Failed to add given input to payload buffer: %v", err)
			} else {
				t.Logf("Successfully added %d bytes given input to payload buffer", written)
			}

			got := plugin.encodedPayloadBuffer.String()

			if d := cmp.Diff(want, got); d != "" {
				t.Errorf("(-want, +got)\n:%s", d)
			} else {
				t.Logf("OK: Payload buffer matches given input.")
			}
		})
	}
}

// TestAddPayloadString_AppendsInputSuccessfullyWhenCalledMultipleTimes
// asserts that a payload buffer populated via the `AddPayloadString` method
// with repeating valid input produces valid output.
func TestAddPayloadString_AppendsInputSuccessfullyWhenCalledMultipleTimes(t *testing.T) {
	t.Parallel()

	plugin := NewPlugin()

	tests := map[string]struct {
		// This represents data as given before it is written to the
		// payload buffer (unencoded).
		input string
	}{
		"simple JSON value": {
			input: smallJSONPayloadUnencoded,
		},
		"simple text value": {
			input: smallPlaintextPayloadUnencoded,
		},
	}

	for name, tt := range tests {
		// Guard against referencing the loop iterator variable directly.
		//
		// https://stackoverflow.com/questions/68559574/using-the-variable-on-range-scope-x-in-function-literal-scopelint
		// https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		tt := tt

		t.Run(name, func(t *testing.T) {
			t.Logf("Evaluating input %q", tt.input)

			plugin.encodedPayloadBuffer.Reset()

			repeat := 2
			for i := 0; i < repeat+1; i++ {
				written, err := plugin.AddPayloadString(tt.input)

				if err != nil {
					t.Fatalf("Failed to add given input to payload buffer: %v", err)
				} else {
					t.Logf("Successfully added %d bytes given input to payload buffer", written)
				}
			}

			// Repeat function call should produce appended output.
			var want string
			for i := 0; i < repeat+1; i++ {
				want += tt.input
			}

			got := plugin.encodedPayloadBuffer.String()

			if d := cmp.Diff(want, got); d != "" {
				t.Errorf("(-want, +got)\n:%s", d)
			} else {
				t.Logf("OK: Payload buffer matches given input.")
			}
		})
	}
}

// TestAddPayloadBytes_AppendsInputSuccessfullyWhenCalledOnce asserts that a
// payload buffer populated via the `AddPayloadBytes` method with
// non-repeating valid input produces valid output.
func TestAddPayloadBytes_AppendsInputSuccessfullyWhenCalledOnce(t *testing.T) {
	t.Parallel()

	plugin := NewPlugin()

	tests := map[string]struct {
		// This represents data as given before it is written to the
		// payload buffer (unencoded).
		input []byte
	}{
		"simple JSON value": {
			input: []byte(smallJSONPayloadUnencoded),
		},
		"simple text value": {
			input: []byte(smallPlaintextPayloadUnencoded),
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
			t.Logf("Evaluating input %q", tt.input)

			plugin.encodedPayloadBuffer.Reset()

			want := string(tt.input)
			written, err := plugin.AddPayloadBytes(tt.input)

			if err != nil {
				t.Fatalf("Failed to add given input to payload buffer: %v", err)
			} else {
				t.Logf("Successfully added %d bytes given input to payload buffer", written)
			}

			got := plugin.encodedPayloadBuffer.String()

			if d := cmp.Diff(want, got); d != "" {
				t.Errorf("(-want, +got)\n:%s", d)
			} else {
				t.Logf("OK: Payload buffer matches given input.")
			}
		})
	}
}

// TestAddPayloadBytes_AppendsInputSuccessfullyWhenCalledMultipleTimes
// asserts that a payload buffer populated via the `AddPayloadBytes` method
// with repeating valid input produces valid output.
func TestAddPayloadBytes_AppendsInputSuccessfullyWhenCalledMultipleTimes(t *testing.T) {
	t.Parallel()

	plugin := NewPlugin()

	tests := map[string]struct {
		// This represents data as given before it is written to the
		// payload buffer (unencoded).
		input []byte
	}{
		"simple JSON value": {
			input: []byte(smallJSONPayloadUnencoded),
		},
		"simple text value": {
			input: []byte(smallPlaintextPayloadUnencoded),
		},
	}

	for name, tt := range tests {
		// Guard against referencing the loop iterator variable directly.
		//
		// https://stackoverflow.com/questions/68559574/using-the-variable-on-range-scope-x-in-function-literal-scopelint
		// https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		tt := tt

		t.Run(name, func(t *testing.T) {
			t.Logf("Evaluating input %q", tt.input)

			plugin.encodedPayloadBuffer.Reset()

			repeat := 2
			for i := 0; i < repeat+1; i++ {
				written, err := plugin.AddPayloadBytes(tt.input)

				if err != nil {
					t.Fatalf("Failed to add given input to payload buffer: %v", err)
				} else {
					t.Logf("Successfully added %d bytes given input to payload buffer", written)
				}
			}

			// Repeat function call should produce appended output.
			var want string
			for i := 0; i < repeat+1; i++ {
				want += string(tt.input)
			}

			got := plugin.encodedPayloadBuffer.String()

			if d := cmp.Diff(want, got); d != "" {
				t.Errorf("(-want, +got)\n:%s", d)
			} else {
				t.Logf("OK: Payload buffer matches given input.")
			}
		})
	}
}

// TestEmptyPerfDataAndEmptyServiceOutputProducesNoOutput asserts that an
// empty Performance Data metrics collection AND empty ServiceOutput produces
// no output.
func TestEmptyPerfDataAndEmptyServiceOutputProducesNoOutput(t *testing.T) {
	t.Parallel()

	// Setup Plugin value manually. This approach does not provide the
	// default time metric that would be provided when using the Plugin
	// constructor.
	var plugin = Plugin{
		LastError:      nil,
		ExitStatusCode: StateOKExitCode,
	}

	var outputBuffer strings.Builder

	// No output should be produced since we don't have anything in the
	// ServiceOutput field.
	plugin.handleServiceOutputSection(&outputBuffer)

	// At this point the collected performance data collection is empty, the
	// field used to hold the entries is nil. An attempt to process the empty
	// collection should result in no output.
	plugin.handlePerformanceData(&outputBuffer)

	want := ""
	got := outputBuffer.String()

	if want != got {
		t.Errorf("\nwant %q\ngot %q", want, got)
	} else {
		t.Logf("OK: Empty performance data collection produces no output.")
	}

}

// TestEmptyClientPerfDataAndConstructedPluginProducesDefaultTimeMetric
// asserts that an empty Performance Data metrics collection AND a constructed
// Plugin value produces a default time metric in the output.
func TestEmptyClientPerfDataAndConstructedPluginProducesDefaultTimeMetric(t *testing.T) {
	t.Parallel()

	// Setup Plugin type the same way that client code using the
	// constructor would.
	plugin := NewPlugin()

	// Performance Data metrics are not emitted if we do not supply a
	// ServiceOutput value.
	plugin.ServiceOutput = "TacoTuesday"

	var outputBuffer strings.Builder

	plugin.handleServiceOutputSection(&outputBuffer)

	// At this point the collected performance data collection is empty, the
	// field used to hold the entries is nil. The default time metric is
	// inserted since client code has not specified this metric.
	plugin.handlePerformanceData(&outputBuffer)

	// Assert that the metric is present.
	defaultTimePerfData, ok := plugin.perfData[defaultTimeMetricLabel]
	if !ok {
		t.Fatal("Default time performance data metric not present when client code omits metrics")
	}

	want := fmt.Sprintf(
		"%s |%s%s",
		plugin.ServiceOutput,
		defaultTimePerfData.String(),
		CheckOutputEOL,
	)
	got := outputBuffer.String()

	if d := cmp.Diff(want, got); d != "" {
		t.Errorf("ERROR: Emitted performance data missing default time metric.")
		t.Errorf("(-want, +got)\n:%s", d)
	} else {
		t.Logf("OK: Emitted performance data contains default time metric.")
	}

}

// TestNonEmptyClientPerfDataAndConstructedPluginRetainsExistingTimeMetric
// asserts that an existing time Performance Data metric is retained when
// using a constructed Plugin value (which emits a default time metric in
// the output if NOT specified by client code).
func TestNonEmptyClientPerfDataAndConstructedPluginRetainsExistingTimeMetric(t *testing.T) {
	t.Parallel()

	// Setup Plugin type the same way that client code using the
	// constructor would.
	plugin := NewPlugin()

	// Performance Data metrics are not emitted if we do not supply a
	// ServiceOutput value.
	plugin.ServiceOutput = "TacoTuesday"

	var outputBuffer strings.Builder

	plugin.handleServiceOutputSection(&outputBuffer)

	// Emulate client code specifying a time metric. This value should not be
	// overwritten with the default time metric.
	clientRuntimeMetric := addTestTimeMetric(t, plugin)

	// Assert that the metric is present.
	_, ok := plugin.perfData[strings.ToLower(clientRuntimeMetric.Label)]
	if !ok {
		t.Fatal("Expected performance data metric from client code is missing")
	}

	// At this point the collected performance data collection is non-empty
	// since client code has specified a time value. The default time metric
	// should NOT be inserted since client code has specified this metric.
	plugin.handlePerformanceData(&outputBuffer)

	want := fmt.Sprintf(
		"%s |%s%s",
		plugin.ServiceOutput,
		clientRuntimeMetric.String(),
		CheckOutputEOL,
	)

	got := outputBuffer.String()

	if d := cmp.Diff(want, got); d != "" {
		t.Errorf("ERROR: Emitted performance data missing client-provided time metric.")
		t.Errorf("(-want, +got)\n:%s", d)
	} else {
		t.Logf("OK: Emitted performance data retains client-provided time metric.")
	}
}

// addTestTimeMetric attaches a test `time` performance data metric regardless
// of whether an existing value is present in the collection. The test metric
// is also returned as a convenience.
func addTestTimeMetric(t *testing.T, p *Plugin) PerformanceData {
	t.Helper()

	const runtimeMetricTestVal = 9000
	runtimeMetric := PerformanceData{
		Label:             defaultTimeMetricLabel,
		Value:             fmt.Sprintf("%d", runtimeMetricTestVal),
		UnitOfMeasurement: defaultTimeMetricUnitOfMeasurement,
	}

	if p.perfData == nil {
		p.perfData = make(map[string]PerformanceData)
	}

	p.perfData[defaultTimeMetricLabel] = runtimeMetric

	return runtimeMetric
}
