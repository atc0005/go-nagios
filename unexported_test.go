// Copyright 2020 Adam Chalkley
//
// https://github.com/atc0005/go-nagios
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

// Package nagios provides test coverage for unexported package functionality.
package nagios

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
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
