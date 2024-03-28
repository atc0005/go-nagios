// Copyright 2023 Codeweavers Ltd
// Copyright 2023 Adam Chalkley
//
// https://github.com/atc0005/go-nagios
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

package nagios

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseRange asserts that the Threshold and Range parsing and evaluation
// functionality works as expected.
//
// See the [Nagios Plugin Dev Guidelines: Threshold and Ranges] definition for
// additional details.
//
// [Nagios Plugin Dev Guidelines: Threshold and Ranges]: https://nagios-plugins.org/doc/guidelines.html#THRESHOLDFORMAT
func TestParseRange(t *testing.T) {

	const pluginServiceOutput string = "CHECK-NT-REPLACEMENT"

	t.Run("Test 0 to N or alert", func(t *testing.T) {
		parsedThing := ParseRangeString("10")
		assert.Equal(t, parsedThing.End, 10.0)
		assert.Equal(t, parsedThing.CheckRange("54"), true)
		assert.Equal(t, parsedThing.CheckRange("-1"), true)
	})

	t.Run("Test N to infinity or alert", func(t *testing.T) {
		parsedThing := ParseRangeString("10:")
		assert.Equal(t, parsedThing.Start, 10.0)
		assert.Equal(t, parsedThing.EndInfinity, true)
		assert.Equal(t, parsedThing.CheckRange("10"), false)
		assert.Equal(t, parsedThing.CheckRange("9"), true)
		assert.Equal(t, parsedThing.CheckRange("-1"), true)
		assert.Equal(t, parsedThing.CheckRange("11"), false)
	})

	t.Run("Within a range involving -inf", func(t *testing.T) {
		parsedThing := ParseRangeString("~:30")
		assert.Equal(t, parsedThing.StartInfinity, true)
		assert.Equal(t, parsedThing.End, 30.0)
		assert.Equal(t, parsedThing.CheckRange("5"), false)
		assert.Equal(t, parsedThing.CheckRange("-10"), false)
		assert.Equal(t, parsedThing.CheckRange("-100"), false)
		assert.Equal(t, parsedThing.CheckRange("30"), false)
		assert.Equal(t, parsedThing.CheckRange("31"), true)
	})

	t.Run("Outside a defined range", func(t *testing.T) {
		parsedThing := ParseRangeString("5:33")
		assert.Equal(t, parsedThing.Start, 5.0)
		assert.Equal(t, parsedThing.End, 33.0)
		assert.Equal(t, parsedThing.CheckRange("33"), false)
		assert.Equal(t, parsedThing.CheckRange("34"), true)
		assert.Equal(t, parsedThing.CheckRange("4"), true)
		assert.Equal(t, parsedThing.CheckRange("5"), false)

	})

	t.Run("Within a defined range", func(t *testing.T) {
		parsedThing := ParseRangeString("10:200")
		assert.Equal(t, parsedThing.Start, 10.0)
		assert.Equal(t, parsedThing.End, 200.0)
		assert.Equal(t, parsedThing.CheckRange("54"), false)
		assert.Equal(t, parsedThing.CheckRange("10"), false)
		assert.Equal(t, parsedThing.CheckRange("9"), true)
		assert.Equal(t, parsedThing.CheckRange("200"), false)
		assert.Equal(t, parsedThing.CheckRange("201"), true)
	})

	t.Run("InsideRange", func(t *testing.T) {
		parsedThing := ParseRangeString("@32:64")
		assert.Equal(t, parsedThing.CheckRange("32"), true)
		assert.Equal(t, parsedThing.CheckRange("33"), true)
		assert.Equal(t, parsedThing.CheckRange("64"), true)
		assert.Equal(t, parsedThing.CheckRange("63"), true)
		assert.Equal(t, parsedThing.CheckRange("31"), false)
		assert.Equal(t, parsedThing.CheckRange("65"), false)
	})

	t.Run("If invalid range is provided (with positive infinity) parsing should return nil", func(t *testing.T) {
		parsedThing := ParseRangeString("50:~")
		assert.Nil(t, parsedThing)
	})

	t.Run("Alert in 0-32", func(t *testing.T) {
		parsedThing := ParseRangeString("@32")

		assert.Equal(t, parsedThing.CheckRange("32"), true)
		assert.Equal(t, parsedThing.CheckRange("31"), true)
		assert.Equal(t, parsedThing.CheckRange("0"), true)
		assert.Equal(t, parsedThing.CheckRange("33"), false)
		assert.Equal(t, parsedThing.CheckRange("-32"), false)
		assert.Equal(t, parsedThing.CheckRange("-1"), false)
	})

	t.Run("Alert on value 32", func(t *testing.T) {
		parsedThing := ParseRangeString("@32:32")

		assert.Equal(t, parsedThing.CheckRange("32"), true)
		assert.Equal(t, parsedThing.CheckRange("31"), false)
		assert.Equal(t, parsedThing.CheckRange("0"), false)
		assert.Equal(t, parsedThing.CheckRange("33"), false)
		assert.Equal(t, parsedThing.CheckRange("-32"), false)
		assert.Equal(t, parsedThing.CheckRange("-1"), false)
	})

	t.Run("Plugin should return exit code OK when value is within acceptable range", func(t *testing.T) {
		var plugin = Plugin{
			ExitStatusCode: StateOKExitCode,
		}
		plugin.ServiceOutput = pluginServiceOutput

		perfdata := PerformanceData{
			Label:             "perfdata label",
			Value:             "18.0",
			UnitOfMeasurement: "C",
			Warn:              "5:30",
			Crit:              "0:40",
		}
		assert.NoError(t, plugin.AddPerfData(false, perfdata))
		assert.NoError(t, plugin.EvaluateThreshold(perfdata))

		assert.Equal(t, StateOKExitCode, plugin.ExitStatusCode)
	})

	t.Run("Plugin should return exit code WARNING when value is within warning range", func(t *testing.T) {
		var plugin = Plugin{
			ExitStatusCode: StateOKExitCode,
		}
		plugin.ServiceOutput = pluginServiceOutput

		perfdata := PerformanceData{
			Label:             "perfdata label",
			Value:             "31.0",
			UnitOfMeasurement: "C",
			Warn:              "5:30",
			Crit:              "0:40",
		}
		assert.NoError(t, plugin.AddPerfData(false, perfdata))
		assert.NoError(t, plugin.EvaluateThreshold(perfdata))

		assert.Equal(t, StateWARNINGExitCode, plugin.ExitStatusCode)
	})

	t.Run("Plugin should return exit code WARNING when value is within warning range", func(t *testing.T) {
		var plugin = Plugin{
			ExitStatusCode: StateOKExitCode,
		}
		plugin.ServiceOutput = pluginServiceOutput

		perfdata := PerformanceData{
			Label:             "perfdata label",
			Value:             "4.0",
			UnitOfMeasurement: "C",
			Warn:              "5:30",
			Crit:              "0:40",
		}
		assert.NoError(t, plugin.AddPerfData(false, perfdata))
		assert.NoError(t, plugin.EvaluateThreshold(perfdata))

		assert.Equal(t, StateWARNINGExitCode, plugin.ExitStatusCode)
	})

	t.Run("Plugin should return exit code CRITICAL when value is within warning range", func(t *testing.T) {
		var plugin = Plugin{
			ExitStatusCode: StateOKExitCode,
		}
		plugin.ServiceOutput = pluginServiceOutput

		perfdata := PerformanceData{
			Label:             "perfdata label",
			Value:             "41.0",
			UnitOfMeasurement: "C",
			Warn:              "5:30",
			Crit:              "0:40",
		}
		assert.NoError(t, plugin.AddPerfData(false, perfdata))
		assert.NoError(t, plugin.EvaluateThreshold(perfdata))

		assert.Equal(t, StateCRITICALExitCode, plugin.ExitStatusCode)
	})

	t.Run("Plugin should return exit code Unknown when critical range is invalid", func(t *testing.T) {
		var plugin = Plugin{
			ExitStatusCode: StateOKExitCode,
		}
		plugin.ServiceOutput = pluginServiceOutput

		perfdata := PerformanceData{
			Label:             "perfdata label",
			Value:             "41.0",
			UnitOfMeasurement: "C",
			Warn:              "5:30",
			Crit:              "0:<=20", // invalid range as critical lower than warning
		}
		assert.NoError(t, plugin.AddPerfData(false, perfdata))
		assert.Error(t, plugin.EvaluateThreshold(perfdata))

		assert.Equal(t, StateUNKNOWNExitCode, plugin.ExitStatusCode)
	})

	t.Run("Plugin should return exit code CRITICAL when value is within warning range", func(t *testing.T) {
		var plugin = Plugin{
			ExitStatusCode: StateOKExitCode,
		}
		plugin.ServiceOutput = pluginServiceOutput

		perfdata := PerformanceData{
			Label:             "perfdata label",
			Value:             "-1.0",
			UnitOfMeasurement: "C",
			Warn:              "5:30",
			Crit:              "0:40",
		}
		assert.NoError(t, plugin.AddPerfData(false, perfdata))
		assert.NoError(t, plugin.EvaluateThreshold(perfdata))

		assert.Equal(t, StateCRITICALExitCode, plugin.ExitStatusCode)
	})

	t.Run("Plugin should return exit code UNKNOWN and error when warning range is invalid", func(t *testing.T) {
		var plugin = Plugin{
			ExitStatusCode: StateOKExitCode,
		}
		plugin.ServiceOutput = pluginServiceOutput

		perfdata := PerformanceData{
			Label:             "perfdata label",
			Value:             "-1.0",
			UnitOfMeasurement: "C",
			Warn:              "x20", // invalid range value
			Crit:              "",
		}
		assert.NoError(t, plugin.AddPerfData(false, perfdata))
		assert.Error(t, plugin.EvaluateThreshold(perfdata))

		assert.Equal(t, StateUNKNOWNExitCode, plugin.ExitStatusCode)
	})
}
