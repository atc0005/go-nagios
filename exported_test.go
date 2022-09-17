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

// The specific format used by the test input file is VERY specific; trailing
// space + newline patterns are intentional. Because "format on save" editor
// functionality easily breaks this input it is stored in a separate file to
// reduce test breakage due to editors "helping".
//
//go:embed testdata/plugin-output-datastore-0001.txt
var pluginOutputDatastore0001 string

// TestPluginOutputIsValid configures an ExitState value as client code would
// and then asserts that the generated output matches manually crafted test
// data.
func TestPluginOutputIsValid(t *testing.T) {
	t.Parallel()

	want := pluginOutputDatastore0001

	// Setup ExitState type the same way that client code would.
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
