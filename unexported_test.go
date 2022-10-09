// Copyright 2020 Adam Chalkley
//
// https://github.com/atc0005/go-nagios
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

// Package nagios provides test coverage for unexported package functionality.
package nagios

import (
	"strings"
	"testing"
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

	// Setup ExitState type the same way that client code would.
	var nagiosExitState = ExitState{
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
	nagiosExitState.ServiceOutput = testInput

	nagiosExitState.handleServiceOutputSection(&output)

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

	// Setup ExitState type the same way that client code would.
	var nagiosExitState = ExitState{
		LastError:      nil,
		ExitStatusCode: StateOKExitCode,
	}

	// Collection of performance data with duplicate entries.
	pd := []PerformanceData{
		{
			Label: "test1",
			Value: "first performance data entry",
		},
		{
			Label: "test1",
			Value: "first performance data entry, repeated",
		},
		{
			Label: "TEST1",
			Value: "first performance data entry, repeated with all upper case",
		},
		{
			Label: "teST1",
			Value: "first performance data entry, repeated with mixed case",
		},
		{
			Label: "test2",
			Value: "not a duplicate",
		},
	}

	if err := nagiosExitState.AddPerfData(false, pd...); err != nil {
		t.Errorf("failed to add initial performance data: %v", err)
	}

	// Standalone performance data metric, duplicate of entry from first
	// collection.
	pd = append(pd, PerformanceData{
		Label: "test1",
		Value: "first performance data entry, repeated by itself",
	})

	if err := nagiosExitState.AddPerfData(false, pd...); err != nil {
		t.Errorf("failed to append additional performance data: %v", err)
	}

	// Two unique labels, so should be just two performance data metrics.
	want := 2
	got := len(nagiosExitState.perfData)

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
