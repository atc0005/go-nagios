// Copyright 2020 Adam Chalkley
//
// https://github.com/atc0005/go-nagios
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

// Package nagios_test provides test coverage for exported package
// functionality.
package nagios_test

import (
	_ "embed"
	"fmt"
	"strings"
	"testing"

	"github.com/atc0005/go-nagios"
	"github.com/google/go-cmp/cmp"
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
)

// TestPluginOutputIsValid configures an ExitState value as client code would
// and then asserts that the generated output matches manually crafted test
// data.
func TestPluginOutputIsValid(t *testing.T) {
	t.Parallel()

	want := pluginOutputDatastore0001

	// Setup ExitState value manually. This approach does not provide the
	// default time metric that would be provided when using the ExitState
	// constructor.
	nagiosExitState := nagios.ExitState{
		LastError:      nil,
		ExitStatusCode: nagios.StateOKExitCode,
	}

	var outputBuffer strings.Builder
	nagiosExitState.SetOutputTarget(&outputBuffer)

	// os.Exit calls break tests
	nagiosExitState.SkipOSExit()

	nagiosExitState.CriticalThreshold = fmt.Sprintf(
		"%d%% datastore usage",
		95,
	)

	nagiosExitState.WarningThreshold = fmt.Sprintf(
		"%d%% datastore usage",
		90,
	)

	nagiosExitState.ServiceOutput =
		"OK: Datastore HUSVM-DC1-vol6 space usage (0 VMs)" +
			" is 0.01% of 18.0TB with 18.0TB remaining" +
			" [WARNING: 90% , CRITICAL: 95%]"

	var longServiceOutputReport strings.Builder

	fmt.Fprintf(
		&longServiceOutputReport,
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

	fmt.Fprintf(
		&longServiceOutputReport,
		"%s---%s%s",
		nagios.CheckOutputEOL,
		nagios.CheckOutputEOL,
		nagios.CheckOutputEOL,
	)

	fmt.Fprintf(
		&longServiceOutputReport,
		"* vSphere environment: %s%s",
		"https://vc1.example.com:443/sdk",
		nagios.CheckOutputEOL,
	)

	fmt.Fprintf(
		&longServiceOutputReport,
		"* Plugin User Agent: %s%s",
		"check-vmware/v0.30.6-0-g25fdcdc",
		nagios.CheckOutputEOL,
	)

	nagiosExitState.LongServiceOutput = longServiceOutputReport.String()

	// Process exit state, emit output to our output buffer.
	nagiosExitState.ReturnCheckResults()

	// Retrieve the output buffer content so that we can compare actual output
	// against our expected output to assert we have a 1:1 match.
	got := outputBuffer.String()

	// if want != got {
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

	// Setup ExitState value manually. This approach does not provide the
	// default time metric that would be provided when using the ExitState
	// constructor.
	nagiosExitState := nagios.ExitState{
		LastError:      nil,
		ExitStatusCode: nagios.StateOKExitCode,
	}

	var outputBuffer strings.Builder
	nagiosExitState.SetOutputTarget(&outputBuffer)

	// os.Exit calls break tests
	nagiosExitState.SkipOSExit()

	nagiosExitState.ServiceOutput =
		"OK: Datastore HUSVM-DC1-vol6 space usage (0 VMs)" +
			" is 0.01% of 18.0TB with 18.0TB remaining" +
			" [WARNING: 90% , CRITICAL: 95%]"

	pd := nagios.PerformanceData{
		Label: "time",
		Value: "874ms",
	}

	if err := nagiosExitState.AddPerfData(false, pd); err != nil {
		t.Errorf("failed to add performance data: %v", err)
	}

	// Process exit state, emit output to our output buffer.
	nagiosExitState.ReturnCheckResults()

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

	// Setup ExitState value manually. This approach does not provide the
	// default time metric that would be provided when using the ExitState
	// constructor.
	nagiosExitState := nagios.ExitState{
		LastError:      nil,
		ExitStatusCode: nagios.StateOKExitCode,
	}

	nagiosExitState.SetOutputTarget(&outputBuffer)

	// os.Exit calls break tests
	nagiosExitState.SkipOSExit()

	nagiosExitState.CriticalThreshold = fmt.Sprintf(
		"%d%% datastore usage",
		95,
	)

	nagiosExitState.WarningThreshold = fmt.Sprintf(
		"%d%% datastore usage",
		90,
	)

	nagiosExitState.ServiceOutput =
		"OK: Datastore HUSVM-DC1-vol6 space usage (0 VMs)" +
			" is 0.01% of 18.0TB with 18.0TB remaining" +
			" [WARNING: 90% , CRITICAL: 95%]"

	var longServiceOutputReport strings.Builder

	fmt.Fprintf(
		&longServiceOutputReport,
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

	fmt.Fprintf(
		&longServiceOutputReport,
		"%s---%s%s",
		nagios.CheckOutputEOL,
		nagios.CheckOutputEOL,
		nagios.CheckOutputEOL,
	)

	fmt.Fprintf(
		&longServiceOutputReport,
		"* vSphere environment: %s%s",
		"https://vc1.example.com:443/sdk",
		nagios.CheckOutputEOL,
	)

	fmt.Fprintf(
		&longServiceOutputReport,
		"* Plugin User Agent: %s%s",
		"check-vmware/v0.30.6-0-g25fdcdc",
		nagios.CheckOutputEOL,
	)

	nagiosExitState.LongServiceOutput = longServiceOutputReport.String()

	pd := nagios.PerformanceData{
		Label: "time",
		Value: "874ms",
	}

	if err := nagiosExitState.AddPerfData(false, pd); err != nil {
		t.Errorf("failed to add performance data: %v", err)
	}

	// Process exit state, emit output to our output buffer.
	nagiosExitState.ReturnCheckResults()

	// Retrieve the output buffer content so that we can compare actual output
	// against our expected output to assert we have a 1:1 match.
	got := outputBuffer.String()

	if d := cmp.Diff(want, got); d != "" {
		t.Errorf("(-want, +got)\n:%s", d)
	}
}

// TestEmptyServiceOutputAndManuallyConstructedExitStateProducesNoOutput
// asserts that an empty ServiceOutput field produces no output when manually
// constructing the ExitState value.
func TestEmptyServiceOutputAndManuallyConstructedExitStateProducesNoOutput(t *testing.T) {
	t.Parallel()

	// Setup ExitState value manually. This approach does not provide the
	// default time metric that would be provided when using the ExitState
	// constructor.
	nagiosExitState := nagios.ExitState{
		LastError:      nil,
		ExitStatusCode: nagios.StateOKExitCode,
	}

	var outputBuffer strings.Builder

	// Explicitly indicate that the field is empty (default/zero value).
	nagiosExitState.ServiceOutput = ""

	nagiosExitState.SetOutputTarget(&outputBuffer)

	// os.Exit calls break tests
	nagiosExitState.SkipOSExit()

	// Process exit state, emit output to our output buffer.
	nagiosExitState.ReturnCheckResults()

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

// TestEmptyServiceOutputAndConstructedExitStateProducesNoOutput asserts that
// an empty ServiceOutput field produces no output. We provide a default time
// metric if client code does not specify one AND if there is ServiceOutput
// content to emit.
func TestEmptyServiceOutputAndConstructedExitStateProducesNoOutput(t *testing.T) {
	t.Parallel()

	// Setup ExitState type the same way that client code using the
	// constructor would.
	nagiosExitState := nagios.New()

	var outputBuffer strings.Builder

	// Explicitly indicate that the field is empty (default/zero value).
	nagiosExitState.ServiceOutput = ""

	nagiosExitState.SetOutputTarget(&outputBuffer)

	// os.Exit calls break tests
	nagiosExitState.SkipOSExit()

	// Process exit state, emit output to our output buffer.
	nagiosExitState.ReturnCheckResults()

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

// TestEmptyClientPerfDataAndConstructedExitStateProducesDefaultTimeMetric
// asserts that omitted performance data from client code produces a default
// time metric when using the ExitState constructor.
func TestEmptyClientPerfDataAndConstructedExitStateProducesDefaultTimeMetric(t *testing.T) {
	t.Parallel()

	// Setup ExitState type the same way that client code using the
	// constructor would.
	nagiosExitState := nagios.New()

	// Performance Data metrics are not emitted if we do not supply a
	// ServiceOutput value.
	nagiosExitState.ServiceOutput = "TacoTuesday"

	var outputBuffer strings.Builder

	nagiosExitState.SetOutputTarget(&outputBuffer)

	// os.Exit calls break tests
	nagiosExitState.SkipOSExit()

	// Process exit state, emit output to our output buffer.
	nagiosExitState.ReturnCheckResults()

	want := fmt.Sprintf(
		"%s | %s",
		nagiosExitState.ServiceOutput,
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
