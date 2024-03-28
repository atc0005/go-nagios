// Copyright 2023 Adam Chalkley
//
// https://github.com/atc0005/go-nagios
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

// Package nagios_test provides test coverage for exported package
// functionality.
package nagios_test

import (
	"testing"

	"github.com/atc0005/go-nagios"
	"github.com/google/go-cmp/cmp"
)

// TestParsePerfDataFailsForInvalidInput asserts that given invalid
// performance data metric strings (as displayed by Nagios, emitted by
// plugins) that parsing as valid PerformanceData values will fail.
func TestParsePerfDataFailsForInvalidInput(t *testing.T) {

	t.Parallel()

	tests := map[string]struct {
		// input is the performance data metrics for a plugin provided as a
		// single string.
		input string
	}{
		"unquoted labels containing spaces": {
			input: `load 1=0.260;5.000;10.000;0; load 5=0.320;4.000;6.000;0; load 15=0.300;3.000;4.000;0;`,
		},

		"value field with a non-numeric value": {
			input: `load1=xyz;5.000;10.000;0; load5=0.320;4.000;6.000;0; load15=0.300;3.000;4.000;0;`,
		},

		"extra semicolons": {
			input: `load1=0.260;5.000;10.000;0;;;;;; load5=0.320;4.000;6.000;0; load15=0.300;3.000;4.000;0;`,
		},

		"empty input": {
			input: "",
		},

		"missing label field": {
			input: `=1;5.000;10.000;0; load5=0.320;4.000;6.000;0; load15=0.300;3.000;4.000;0;`,
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

			result, err := nagios.ParsePerfData(tt.input)
			if err == nil {
				t.Logf("result: %+v:", result)
				t.Fatalf(
					"\nwant error when parsing invalid perfdata input\ngot successful parsing result",
				)
			} else {
				t.Logf("result: %+v:", result)
				t.Logf(
					"OK: \nwant error when parsing invalid perfdata input\ngot error as expected: %v", err,
				)
			}
		})
	}

}

// TestParsePerfDataSucceedsForValidInput asserts that given valid performance
// data metric strings (as displayed by Nagios, emitted by plugins) that
// parsing as valid PerformanceData values will succeed as expected.
// Additionally, the resulting PerformanceData values are compared against
// expected PerformanceData values to assert all fields are as expected.
func TestParsePerfDataSucceedsForValidInput(t *testing.T) {

	t.Parallel()

	tests := map[string]struct {
		// input is the performance data metrics for a plugin provided as a
		// single string.
		input string

		// result provides both the literal expected values generated from
		// parsing the input performance data string and the implicit overall
		// results count.
		result []nagios.PerformanceData
	}{
		"Load averages double quoted": {
			// https://github.com/nagios-plugins/nagios-plugins/blob/12446aea1d353d891cd6291ba8086a0f5247c93d/plugins/check_load.c#L206-L210
			input: `"load1=0.260;5.000;10.000;0; load5=0.320;4.000;6.000;0; load15=0.300;3.000;4.000;0;"`,
			result: []nagios.PerformanceData{
				{
					Label:             "load1",
					Value:             "0.260",
					UnitOfMeasurement: "",
					Warn:              "5.000",
					Crit:              "10.000",
					Min:               "0",
					Max:               "",
				},
				{
					Label:             "load5",
					Value:             "0.320",
					UnitOfMeasurement: "",
					Warn:              "4.000",
					Crit:              "6.000",
					Min:               "0",
					Max:               "",
				},
				{
					Label:             "load15",
					Value:             "0.300",
					UnitOfMeasurement: "",
					Warn:              "3.000",
					Crit:              "4.000",
					Min:               "0",
					Max:               "",
				},
			},
		},

		"Load averages unquoted": {
			// https://github.com/nagios-plugins/nagios-plugins/blob/12446aea1d353d891cd6291ba8086a0f5247c93d/plugins/check_load.c#L206-L210
			input: `load1=0.260;5.000;10.000;0; load5=0.320;4.000;6.000;0; load15=0.300;3.000;4.000;0;`,
			result: []nagios.PerformanceData{
				{
					Label:             "load1",
					Value:             "0.260",
					UnitOfMeasurement: "",
					Warn:              "5.000",
					Crit:              "10.000",
					Min:               "0",
					Max:               "",
				},
				{
					Label:             "load5",
					Value:             "0.320",
					UnitOfMeasurement: "",
					Warn:              "4.000",
					Crit:              "6.000",
					Min:               "0",
					Max:               "",
				},
				{
					Label:             "load15",
					Value:             "0.300",
					UnitOfMeasurement: "",
					Warn:              "3.000",
					Crit:              "4.000",
					Min:               "0",
					Max:               "",
				},
			},
		},

		"Single quoted time metric label with all semicolon separators": {
			input: `'time'=49ms;;;;`,
			result: []nagios.PerformanceData{
				{
					Label:             "time",
					Value:             "49",
					UnitOfMeasurement: "ms",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
			},
		},

		"Single quoted time metric label with one trailing semicolon separator": {
			input: `'time'=49ms;`,
			result: []nagios.PerformanceData{
				{
					Label:             "time",
					Value:             "49",
					UnitOfMeasurement: "ms",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
			},
		},

		"Single quoted time metric label without semicolon separators": {
			input: `'time'=49ms`,
			result: []nagios.PerformanceData{
				{
					Label:             "time",
					Value:             "49",
					UnitOfMeasurement: "ms",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
			},
		},

		"Disk usage labels single quoted": {
			input: `'/'=7826MB;28621;30211;0;31802 '/dev/shm'=0MB;3542;3739;0;3936 '/boot'=40MB;428;452;0;476`,
			result: []nagios.PerformanceData{
				{
					Label:             "/",
					Value:             "7826",
					UnitOfMeasurement: "MB",
					Warn:              "28621",
					Crit:              "30211",
					Min:               "0",
					Max:               "31802",
				},
				{
					Label:             "/dev/shm",
					Value:             "0",
					UnitOfMeasurement: "MB",
					Warn:              "3542",
					Crit:              "3739",
					Min:               "0",
					Max:               "3936",
				},
				{
					Label:             "/boot",
					Value:             "40",
					UnitOfMeasurement: "MB",
					Warn:              "428",
					Crit:              "452",
					Min:               "0",
					Max:               "476",
				},
			},
		},

		"Disk usage unquoted": {
			input: `/=7826MB;28621;30211;0;31802 /dev/shm=0MB;3542;3739;0;3936 /boot=40MB;428;452;0;476`,
			result: []nagios.PerformanceData{
				{
					Label:             "/",
					Value:             "7826",
					UnitOfMeasurement: "MB",
					Warn:              "28621",
					Crit:              "30211",
					Min:               "0",
					Max:               "31802",
				},
				{
					Label:             "/dev/shm",
					Value:             "0",
					UnitOfMeasurement: "MB",
					Warn:              "3542",
					Crit:              "3739",
					Min:               "0",
					Max:               "3936",
				},
				{
					Label:             "/boot",
					Value:             "40",
					UnitOfMeasurement: "MB",
					Warn:              "428",
					Crit:              "452",
					Min:               "0",
					Max:               "476",
				},
			},
		},

		"Processes unquoted": {
			input: `procs=7307;450;600;0;`,
			result: []nagios.PerformanceData{
				{
					Label:             "procs",
					Value:             "7307",
					UnitOfMeasurement: "",
					Warn:              "450",
					Crit:              "600",
					Min:               "0",
					Max:               "",
				},
			},
		},

		"VMware Snapshots Age single quoted with all semicolon separators": {
			input: `'critical_snapshots'=1;;;; 'resource_pools_evaluated'=29;;;; 'resource_pools_excluded'=0;;;; 'resource_pools_included'=0;;;; 'snapshots'=11;;;; 'time'=1720ms;;;; 'vms'=495;;;; 'vms_with_critical_snapshots'=1;;;; 'vms_with_warning_snapshots'=0;;;; 'warning_snapshots'=0;;;;`,
			result: []nagios.PerformanceData{
				{
					Label:             "critical_snapshots",
					Value:             "1",
					UnitOfMeasurement: "",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "resource_pools_evaluated",
					Value:             "29",
					UnitOfMeasurement: "",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "resource_pools_excluded",
					Value:             "0",
					UnitOfMeasurement: "",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "resource_pools_included",
					Value:             "0",
					UnitOfMeasurement: "",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "snapshots",
					Value:             "11",
					UnitOfMeasurement: "",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "time",
					Value:             "1720",
					UnitOfMeasurement: "ms",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "vms",
					Value:             "495",
					UnitOfMeasurement: "",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "vms_with_critical_snapshots",
					Value:             "1",
					UnitOfMeasurement: "",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "vms_with_warning_snapshots",
					Value:             "0",
					UnitOfMeasurement: "",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "warning_snapshots",
					Value:             "0",
					UnitOfMeasurement: "",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
			},
		},

		"All Statuspage components single quoted with all semicolon separators": {
			// NOTE: While the specific values from this test case are off
			// (components in a warning state compared to all problem
			// components), the metrics collection is still in a valid format.
			input: `'all_component_groups'=13;;;; 'all_components'=333;;;; 'all_components_critical'=0;;;; 'all_components_ok'=326;;;; 'all_components_unknown'=0;;;; 'all_components_warning'=7;;;; 'all_problem_components'=4;;;; 'excluded_problem_components'=0;;;; 'remaining_components_critical'=0;;;; 'remaining_components_ok'=326;;;; 'remaining_components_unknown'=0;;;; 'remaining_components_warning'=7;;;; 'remaining_problem_components'=4;;;; 'time'=283ms;;;;`,
			result: []nagios.PerformanceData{
				{
					Label:             "all_component_groups",
					Value:             "13",
					UnitOfMeasurement: "",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "all_components",
					Value:             "333",
					UnitOfMeasurement: "",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "all_components_critical",
					Value:             "0",
					UnitOfMeasurement: "",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "all_components_ok",
					Value:             "326",
					UnitOfMeasurement: "",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "all_components_unknown",
					Value:             "0",
					UnitOfMeasurement: "",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "all_components_warning",
					Value:             "7",
					UnitOfMeasurement: "",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "all_problem_components",
					Value:             "4",
					UnitOfMeasurement: "",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "excluded_problem_components",
					Value:             "0",
					UnitOfMeasurement: "",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "remaining_components_critical",
					Value:             "0",
					UnitOfMeasurement: "",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "remaining_components_ok",
					Value:             "326",
					UnitOfMeasurement: "",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "remaining_components_unknown",
					Value:             "0",
					UnitOfMeasurement: "",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "remaining_components_warning",
					Value:             "7",
					UnitOfMeasurement: "",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "remaining_problem_components",
					Value:             "4",
					UnitOfMeasurement: "",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "time",
					Value:             "283",
					UnitOfMeasurement: "ms",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
			},
		},

		"check_cert plugin metrics single quoted with all semicolon separators": {
			input: `'certs_present_intermediate'=2;;;; 'certs_present_leaf'=1;;;; 'certs_present_root'=0;;;; 'certs_present_unknown'=0;;;; 'expires_intermediate'=1703d;30;15;; 'expires_leaf'=62d;30;15;; 'time'=41ms;;;;`,
			result: []nagios.PerformanceData{
				{
					Label:             "certs_present_intermediate",
					Value:             "2",
					UnitOfMeasurement: "",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "certs_present_leaf",
					Value:             "1",
					UnitOfMeasurement: "",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "certs_present_root",
					Value:             "0",
					UnitOfMeasurement: "",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "certs_present_unknown",
					Value:             "0",
					UnitOfMeasurement: "",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "expires_intermediate",
					Value:             "1703",
					UnitOfMeasurement: "d",
					Warn:              "30",
					Crit:              "15",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "expires_leaf",
					Value:             "62",
					UnitOfMeasurement: "d",
					Warn:              "30",
					Crit:              "15",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "time",
					Value:             "41",
					UnitOfMeasurement: "ms",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
			},
		},
		"check_cert plugin lifetime metrics single quoted with all semicolon separators": {
			input: `'life_remaining_intermediate'=48%;;;; 'life_remaining_leaf'=32%;;;;`,
			result: []nagios.PerformanceData{
				{
					Label:             "life_remaining_intermediate",
					Value:             "48",
					UnitOfMeasurement: "%",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
				{
					Label:             "life_remaining_leaf",
					Value:             "32",
					UnitOfMeasurement: "%",
					Warn:              "",
					Crit:              "",
					Min:               "",
					Max:               "",
				},
			},
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

			perfDataResults, err := nagios.ParsePerfData(tt.input)
			if err != nil {
				t.Fatalf("failed to parse perfdata input: %v", err)
			}

			t.Logf("perfDataResults: %v", perfDataResults)

			testParsePerfDataCollection(t, perfDataResults, tt.result)

		})
	}

}

func testParsePerfDataCollection(
	t *testing.T,
	expected []nagios.PerformanceData,
	results []nagios.PerformanceData,
) {
	t.Helper()

	// Start with asserting that the perfdata metrics count is as expected
	// before performing more specific checks.
	switch {
	case len(results) != len(expected):
		want := len(expected)
		got := len(results)
		t.Fatalf(
			"\nwant %d perfdata metrics from input"+
				"\ngot %d perfdata metrics from input", want, got,
		)
	default:
		got := len(results)
		t.Logf("OK: got %d perfdata metrics from input", got)
	}

	// Since the parsed performance data metrics and the expected metrics
	// collection are in the same order we use the index from looping through
	// the parsed results to retrieve the expected result for comparison.
	for i := range results {
		want := expected[i]
		got := results[i]

		switch d := cmp.Diff(want, got); {
		case d != "":
			t.Errorf("ERROR: Parsed perfdata result does not match expected result")
			t.Errorf("(-want, +got)\n:%s", d)
		default:
			t.Logf("OK: Parsed perfdata result matches expected result")
			t.Log(got.String())
		}
	}
}
