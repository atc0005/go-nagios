// Copyright 2024 Adam Chalkley
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
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPlugin_SetDebugLoggingOutputTarget_IsValidWithValidInput(t *testing.T) {
	t.Parallel()

	plugin := NewPlugin()

	// Assert that log output sink is still unset
	if plugin.logOutputSink != nil {
		t.Fatal("ERROR: plugin logOutputSink is not at the expected default unset value.")
	} else {
		t.Log("OK: plugin logOutputSink is at the expected default unset value.")
	}

	// Assert that logger is still unset
	if plugin.logger != nil {
		t.Fatal("ERROR: plugin logger is not at the expected default unset value.")
	} else {
		t.Log("OK: plugin logger is at the expected default unset value.")
	}

	var outputBuffer strings.Builder

	plugin.SetDebugLoggingOutputTarget(&outputBuffer)

	// All debug logging options were previously enabled after setting a debug
	// logging output target. This behavior has changed and now debug logging
	// must be explicitly enabled.
	//
	// assertAllDebugLoggingOptionsAreEnabled(plugin, t)

	// Assert that plugin.outputSink is set as expected.
	switch {
	case plugin.logOutputSink == nil:
		t.Fatal("ERROR: plugin logOutputSink is unset instead of the given custom value.")
	case plugin.logOutputSink == defaultPluginDebugLoggingOutputTarget():
		t.Fatal("ERROR: plugin logOutputSink is set to the default/fallback value instead of the expected custom value.")
	case plugin.logOutputSink != &outputBuffer:
		t.Error("ERROR: logOutputSink is not set to custom output target")
		// t.Logf("plugin.logOutputSink address: %p", plugin.logOutputSink)
		// t.Logf("&outputBuffer address: %p", &outputBuffer)

		d := cmp.Diff(&outputBuffer, plugin.logOutputSink)
		t.Fatalf("(-want, +got)\n:%s", d)
	default:
		t.Log("OK: plugin logOutputSink is at the expected custom value.")
	}

	assertLoggerIsConfiguredProperlyAfterSettingDebugLoggingOutputTarget(plugin, t)
}

func TestPlugin_SetDebugLoggingOutputTarget_CorrectlySetsFallbackLoggingTargetWithInvalidInput(t *testing.T) {
	t.Parallel()

	plugin := NewPlugin()

	// By calling this method we implicitly enable debug logging.
	//
	// By setting an invalid output target a log message is emitted to the
	// default debug log output target.
	plugin.SetDebugLoggingOutputTarget(nil)

	// Assert that plugin.outputSink is set as expected.
	want := defaultPluginDebugLoggingOutputTarget()
	got := plugin.logOutputSink

	switch {
	case got == nil:
		t.Error("ERROR: plugin debug log output target is unset instead of the default/fallback value.")
	case got != want:
		t.Error("ERROR: plugin debug log output target is not set to the default/fallback value.")
		d := cmp.Diff(want, got)
		t.Fatalf("(-want, +got)\n:%s", d)
	default:
		t.Log("OK: plugin debug log output target is at the expected default/fallback value.")
	}
}

func TestPlugin_setupLogger_CorrectlySetsDefaultLoggerTargetWhenDebugLogOutputSinkIsUnset(t *testing.T) {
	t.Parallel()

	plugin := NewPlugin()

	switch {
	case plugin.logger != nil:
		t.Fatal("ERROR: plugin logger is not at the expected default unset value.")
	default:
		t.Log("OK: plugin logger is at the expected default unset value.")
	}

	plugin.setupLogger()

	switch {
	case plugin.logger == nil:
		t.Fatal("ERROR: plugin logger is unset instead of being configured for use.")
	default:
		t.Log("OK: plugin logger is set as expected.")
	}

	loggerTarget := plugin.logger.Writer()
	switch {
	case loggerTarget == nil:
		t.Fatal("ERROR: plugin logger target is unset instead of being configured for use.")
	case loggerTarget != defaultPluginDebugLoggerTarget():
		t.Fatal("ERROR: plugin logger target is not set to use default logger target as expected.")
	default:
		t.Logf("OK: plugin logger target is set to default logger target ('%#v') as expected.", defaultPluginDebugLoggerTarget())
	}
}

func TestPlugin_setupLogger_CorrectlySetsLoggerTargetWhenLogOutputSinkIsSet(t *testing.T) {
	t.Parallel()

	plugin := NewPlugin()

	// Configure log output sink to a custom target.
	var outputBuffer strings.Builder
	plugin.logOutputSink = &outputBuffer

	switch {
	case plugin.logger != nil:
		t.Fatal("ERROR: plugin logger is not at the expected default unset value.")
	default:
		t.Log("OK: plugin logger is at the expected default unset value.")
	}

	plugin.setupLogger()

	switch {
	case plugin.logger == nil:
		t.Fatal("ERROR: plugin logger is unset instead of being configured for use.")
	default:
		t.Log("OK: plugin logger is set as expected.")
	}

	loggerTarget := plugin.logger.Writer()
	switch {
	case loggerTarget == nil:
		t.Fatal("ERROR: plugin logger target is unset instead of being configured for use.")
	case loggerTarget != plugin.logOutputSink:
		t.Fatal("ERROR: plugin logger target is not set to custom output sink as expected.")
	default:
		t.Log("OK: plugin logger target is set as expected.")
	}
}

func TestPlugin_DebugLoggingEnableAll_CorrectlyConfiguresLogTargetAndLoggerWithFallbackValues(t *testing.T) {
	t.Parallel()

	plugin := NewPlugin()

	// Assert that log output sink is still unset
	if plugin.logOutputSink != nil {
		t.Fatal("ERROR: plugin logOutputSink is not at the expected default unset value.")
	} else {
		t.Log("OK: plugin logOutputSink is at the expected default unset value.")
	}

	// Assert that logger is still unset
	if plugin.logger != nil {
		t.Fatal("ERROR: plugin logger is not at the expected default unset value.")
	} else {
		t.Log("OK: plugin logger is at the expected default unset value.")
	}

	// Expected results of calling this function:
	//
	// - the fallback debug log target is set
	// - the logger is setup
	plugin.DebugLoggingEnableAll()

	switch {
	case plugin.logger == nil:
		t.Fatal("ERROR: plugin logger is unset instead of being configured for use.")
	default:
		t.Log("OK: plugin logger is set as expected.")
	}

	loggerTarget := plugin.logger.Writer()
	switch {
	case loggerTarget == nil:
		t.Fatal("ERROR: plugin logger target is unset instead of being configured for use.")
	case loggerTarget != plugin.logOutputSink:
		t.Fatal("ERROR: plugin logger target is not set to match debug log output target as expected.")
	default:
		t.Log("OK: plugin logger target is set as expected.")
	}
}

func TestPlugin_DebugLoggingEnableAll_CorrectlyEnablesAllDebugLoggingOptions(t *testing.T) {
	t.Parallel()

	plugin := NewPlugin()

	plugin.DebugLoggingEnableAll()

	assertAllDebugLoggingOptionsAreEnabled(plugin, t)
}

func TestPlugin_DebugLoggingDisableAll_CorrectlyLeavesLogTargetAndLoggerUnmodified(t *testing.T) {
	t.Parallel()

	plugin := NewPlugin()

	// Assert that log output sink is still unset
	if plugin.logOutputSink != nil {
		t.Fatal("ERROR: plugin logOutputSink is not at the expected default unset value.")
	} else {
		t.Log("OK: plugin logOutputSink is at the expected default unset value.")
	}

	// Configure log output sink to a custom target.
	var outputBuffer strings.Builder
	plugin.logOutputSink = &outputBuffer

	switch {
	case plugin.logger != nil:
		t.Fatal("ERROR: plugin logger is not at the expected default unset value.")
	default:
		t.Log("OK: plugin logger is at the expected default unset value.")
	}

	plugin.setupLogger()

	switch {
	case plugin.logger == nil:
		t.Fatal("ERROR: plugin logger is unset instead of being configured for use.")
	default:
		t.Log("OK: plugin logger is set as expected.")
	}

	pluginLoggerBeforeDisablingLogging := plugin.logger
	pluginLogTargetBeforeDisablingLogging := plugin.logOutputSink

	plugin.DebugLoggingDisableAll()

	// Assert that the debug log target remains untouched.
	switch {
	case plugin.logOutputSink != pluginLogTargetBeforeDisablingLogging:
		t.Errorf("ERROR: plugin debug log target is not set to same value before logging was disabled.")
		d := cmp.Diff(
			plugin.logOutputSink,
			pluginLogTargetBeforeDisablingLogging,
			// https://stackoverflow.com/questions/73476661/cmp-equal-gives-panic-message-cannot-handle-unexported-field-at
			cmp.AllowUnexported(strings.Builder{}),
		)

		t.Errorf("(-want, +got)\n:%s", d)

	default:
		t.Log("OK: plugin debug log target is set to same value before logging was disabled.")
	}

	// Assert that the logger remains untouched.
	switch {
	case plugin.logger != pluginLoggerBeforeDisablingLogging:
		t.Fatal("ERROR: plugin logger is not set to same value before logging was disabled.")
	default:
		t.Log("OK: plugin logger is set to same value before logging was disabled.")
	}
}

func TestPlugin_DebugLoggingDisableAll_CorrectlyDisablesAllDebugLoggingOptions(t *testing.T) {
	t.Parallel()

	plugin := NewPlugin()

	plugin.DebugLoggingDisableAll()

	assertAllDebugLoggingOptionsAreDisabled(plugin, t)
}

func TestPlugin_DebugLoggingEnableActions_CorrectlyConfiguresLogTargetAndLoggerWithFallbackValues(t *testing.T) {
	t.Parallel()

	plugin := NewPlugin()

	// Assert that log output sink is still unset
	if plugin.logOutputSink != nil {
		t.Fatal("ERROR: plugin logOutputSink is not at the expected default unset value.")
	} else {
		t.Log("OK: plugin logOutputSink is at the expected default unset value.")
	}

	// Assert that logger is still unset
	if plugin.logger != nil {
		t.Fatal("ERROR: plugin logger is not at the expected default unset value.")
	} else {
		t.Log("OK: plugin logger is at the expected default unset value.")
	}

	// Expected results of calling this function:
	//
	// - the fallback debug log target is set
	// - the logger is setup
	plugin.DebugLoggingEnableActions()

	switch {
	case plugin.logger == nil:
		t.Fatal("ERROR: plugin logger is unset instead of being configured for use.")
	default:
		t.Log("OK: plugin logger is set as expected.")
	}

	loggerTarget := plugin.logger.Writer()
	switch {
	case loggerTarget == nil:
		t.Fatal("ERROR: plugin logger target is unset instead of being configured for use.")
	case loggerTarget != plugin.logOutputSink:
		t.Fatal("ERROR: plugin logger target is not set to match debug log output target as expected.")
	default:
		t.Log("OK: plugin logger target is set as expected.")
	}
}

func TestPlugin_DebugLoggingEnableActions_CorrectlyEnablesOnlyDebugLoggingActionsOption(t *testing.T) {
	t.Parallel()

	plugin := NewPlugin()

	// Flip everything off to start with so we can selectively enable just the
	// debug logging option we're interested in.
	plugin.debugLogging = allDebugLoggingOptionsDisabled()

	plugin.DebugLoggingEnableActions()

	selectDebugLoggingOptionsEnabled := allDebugLoggingOptionsDisabled()
	selectDebugLoggingOptionsEnabled.actions = true

	if !cmp.Equal(
		selectDebugLoggingOptionsEnabled,

		// https://stackoverflow.com/questions/73476661/cmp-equal-gives-panic-message-cannot-handle-unexported-field-at
		plugin.debugLogging, cmp.AllowUnexported(debugLoggingOptions{}),
	) {
		d := cmp.Diff(
			selectDebugLoggingOptionsEnabled,
			plugin.debugLogging,

			// https://stackoverflow.com/questions/73476661/cmp-equal-gives-panic-message-cannot-handle-unexported-field-at
			cmp.AllowUnexported(debugLoggingOptions{}),
		)

		t.Errorf("(-want, +got)\n:%s", d)
	}
}

func TestPlugin_DebugLoggingDisableActions_CorrectlyDisablesOnlyDebugLoggingActionsOption(t *testing.T) {
	t.Parallel()

	plugin := NewPlugin()

	// Flip everything on to start with so we can selectively disable specific
	// debug logging options.
	plugin.debugLogging = allDebugLoggingOptionsEnabled()

	plugin.DebugLoggingDisableActions()

	selectDebugLoggingOptionsDisabled := allDebugLoggingOptionsEnabled()
	selectDebugLoggingOptionsDisabled.actions = false

	if !cmp.Equal(
		selectDebugLoggingOptionsDisabled,

		// https://stackoverflow.com/questions/73476661/cmp-equal-gives-panic-message-cannot-handle-unexported-field-at
		plugin.debugLogging, cmp.AllowUnexported(debugLoggingOptions{}),
	) {
		d := cmp.Diff(
			selectDebugLoggingOptionsDisabled,
			plugin.debugLogging,

			// https://stackoverflow.com/questions/73476661/cmp-equal-gives-panic-message-cannot-handle-unexported-field-at
			cmp.AllowUnexported(debugLoggingOptions{}),
		)

		t.Errorf("(-want, +got)\n:%s", d)
	}
}

func TestPlugin_DebugLoggingEnablePluginOutputSize_CorrectlyConfiguresLogTargetAndLoggerWithFallbackValues(t *testing.T) {
	t.Parallel()

	plugin := NewPlugin()

	// Assert that log output sink is still unset
	if plugin.logOutputSink != nil {
		t.Fatal("ERROR: plugin logOutputSink is not at the expected default unset value.")
	} else {
		t.Log("OK: plugin logOutputSink is at the expected default unset value.")
	}

	// Assert that logger is still unset
	if plugin.logger != nil {
		t.Fatal("ERROR: plugin logger is not at the expected default unset value.")
	} else {
		t.Log("OK: plugin logger is at the expected default unset value.")
	}

	// Expected results of calling this function:
	//
	// - the fallback debug log target is set
	// - the logger is setup
	plugin.DebugLoggingEnablePluginOutputSize()

	switch {
	case plugin.logger == nil:
		t.Fatal("ERROR: plugin logger is unset instead of being configured for use.")
	default:
		t.Log("OK: plugin logger is set as expected.")
	}

	loggerTarget := plugin.logger.Writer()
	switch {
	case loggerTarget == nil:
		t.Fatal("ERROR: plugin logger target is unset instead of being configured for use.")
	case loggerTarget != plugin.logOutputSink:
		t.Fatal("ERROR: plugin logger target is not set to match debug log output target as expected.")
	default:
		t.Log("OK: plugin logger target is set as expected.")
	}
}

func TestPlugin_DebugLoggingEnablePluginOutputSize_CorrectlyEnablesOnlyDebugLoggingOutputSizeOption(t *testing.T) {
	t.Parallel()

	plugin := NewPlugin()

	// Flip everything off to start with so we can selectively enable just the
	// debug logging option we're interested in.
	plugin.debugLogging = allDebugLoggingOptionsDisabled()

	plugin.DebugLoggingEnablePluginOutputSize()

	selectDebugLoggingOptionsEnabled := allDebugLoggingOptionsDisabled()
	selectDebugLoggingOptionsEnabled.pluginOutputSize = true

	if !cmp.Equal(
		selectDebugLoggingOptionsEnabled,

		// https://stackoverflow.com/questions/73476661/cmp-equal-gives-panic-message-cannot-handle-unexported-field-at
		plugin.debugLogging, cmp.AllowUnexported(debugLoggingOptions{}),
	) {
		d := cmp.Diff(
			selectDebugLoggingOptionsEnabled,
			plugin.debugLogging,

			// https://stackoverflow.com/questions/73476661/cmp-equal-gives-panic-message-cannot-handle-unexported-field-at
			cmp.AllowUnexported(debugLoggingOptions{}),
		)

		t.Errorf("(-want, +got)\n:%s", d)
	}
}

func TestPlugin_DebugLoggingDisablePluginOutputSize_CorrectlyDisablesOnlyDebugLoggingOutputSizeOption(t *testing.T) {
	t.Parallel()

	plugin := NewPlugin()

	// Flip everything on to start with so we can selectively disable specific
	// debug logging options.
	plugin.debugLogging = allDebugLoggingOptionsEnabled()

	plugin.DebugLoggingDisablePluginOutputSize()

	selectDebugLoggingOptionsDisabled := allDebugLoggingOptionsEnabled()
	selectDebugLoggingOptionsDisabled.pluginOutputSize = false

	if !cmp.Equal(
		selectDebugLoggingOptionsDisabled,

		// https://stackoverflow.com/questions/73476661/cmp-equal-gives-panic-message-cannot-handle-unexported-field-at
		plugin.debugLogging, cmp.AllowUnexported(debugLoggingOptions{}),
	) {
		d := cmp.Diff(
			selectDebugLoggingOptionsDisabled,
			plugin.debugLogging,

			// https://stackoverflow.com/questions/73476661/cmp-equal-gives-panic-message-cannot-handle-unexported-field-at
			cmp.AllowUnexported(debugLoggingOptions{}),
		)

		t.Errorf("(-want, +got)\n:%s", d)
	}

}

func TestPlugin_log_CorrectlyProducesNoOutputWhenLoggerIsUnset(t *testing.T) {

	// credit: Modified version of Google Gemini example code (via Search)
	//
	// prompt: "golang using os.Pipe() to change os.Stderr and os.Stdout for tests"

	// Intentionally *not* running this test in parallel since we're going to
	// monkey patch io.StdOut and io.StdErr briefly.
	//
	// t.Parallel()

	var (
		origStdErr = os.Stderr
		origStdOut = os.Stdout
	)

	// Ensure that no matter what, we restore the original values to prevent
	// affecting other tests.
	t.Cleanup(func() {
		os.Stderr = origStdErr
		os.Stdout = origStdOut
	})

	plugin := NewPlugin()

	switch {
	case plugin.logOutputSink != nil:
		t.Fatal("ERROR: plugin logOutputSink is not at the expected default unset value.")
	default:
		t.Log("OK: plugin logOutputSink is at the expected default unset value.")
	}

	switch {
	case plugin.logger != nil:
		t.Fatal("ERROR: plugin logger is not at the expected default unset value.")
	default:
		t.Log("OK: plugin logger is at the expected default unset value.")
	}

	// Create a new pipe for capturing standard output
	r, w, stdOutOsPipeErr := os.Pipe()
	if stdOutOsPipeErr != nil {
		t.Fatal("Error creating pipe to emulate os.Stdout as part of test setup")
	}
	os.Stdout = w

	// Create a new pipe for capturing standard error
	rErr, wErr, stdErrOsPipeErr := os.Pipe()
	if stdErrOsPipeErr != nil {
		t.Fatal("Error creating pipe to emulate os.Stderr as part of test setup")
	}
	os.Stderr = wErr

	// This shouldn't go anywhere.
	plugin.log("Testing")

	// Close the write ends of the pipes to signal that we're done writing
	if err := w.Close(); err != nil {
		t.Fatalf("Error closing stdout pipe: %v", err)
	}

	if err := wErr.Close(); err != nil {
		t.Fatalf("Error closing stdout pipe: %v", err)
	}

	// Restore original stdout and stderr values.
	os.Stderr = origStdErr
	os.Stdout = origStdOut

	// Read the output from the pipes
	var stdOutBuffer, stdErrBuffer bytes.Buffer

	var written int64
	var ioCopyErr error

	written, ioCopyErr = io.Copy(&stdOutBuffer, r)
	switch {
	case ioCopyErr != nil:
		t.Fatalf("ERROR: Failed to copy stdout pipe content to stdout buffer for evaluation: %v", ioCopyErr)
	case written != 0:
		t.Errorf("ERROR: Copied %d bytes of unexpected content to stdout buffer", written)
	default:
		t.Log("OK: io.Copy operation on stdout pipe found no content but also encountered no errors")
	}

	written, ioCopyErr = io.Copy(&stdErrBuffer, rErr)
	switch {
	case ioCopyErr != nil:
		t.Fatalf("ERROR: Failed to copy stderr pipe content to stderr buffer for evaluation: %v", ioCopyErr)
	case written != 0:
		t.Errorf("ERROR: Copied %d bytes of unexpected content to stderr buffer", written)
	default:
		t.Log("OK: io.Copy operation on stderr pipe found no content but also encountered no errors")
	}

	capturedStdOut := stdOutBuffer.String()
	switch {
	case capturedStdOut != "":
		want := ""
		got := capturedStdOut
		d := cmp.Diff(want, got)
		t.Fatalf("(-want, +got)\n:%s", d)
	default:
		t.Log("OK: No output logged to stdout as expected.")
	}

	capturedStdErr := stdErrBuffer.String()
	switch {
	case capturedStdErr != "":
		want := ""
		got := capturedStdErr
		d := cmp.Diff(want, got)
		t.Fatalf("(-want, +got)\n:%s", d)
	default:
		t.Log("OK: No output logged to stderr as expected.")
	}
}

func TestPlugin_logAction_CorrectlyProducesNoOutputWhenDebugLoggingActionsOptionIsDisabled(t *testing.T) {
	plugin := NewPlugin()

	var outputBuffer strings.Builder

	plugin.SetDebugLoggingOutputTarget(&outputBuffer)
	plugin.debugLogging.actions = false

	// This shouldn't go anywhere.
	testMsg := "Test action entry"
	plugin.logAction(testMsg)

	capturedDebugLogOutput := outputBuffer.String()
	switch {
	case strings.Contains(capturedDebugLogOutput, testMsg):
		want := removeEntry(capturedDebugLogOutput, testMsg, CheckOutputEOL)
		got := capturedDebugLogOutput
		d := cmp.Diff(want, got)
		t.Fatalf("(-want, +got)\n:%s", d)
	default:
		t.Log("OK: No debug logging output captured as expected.")
	}
}

func TestPlugin_logPluginOutputSize_CorrectlyProducesNoOutputWhenDebugLoggingOutputSizeOptionIsDisabled(t *testing.T) {
	plugin := NewPlugin()

	var outputBuffer strings.Builder

	plugin.SetDebugLoggingOutputTarget(&outputBuffer)
	plugin.debugLogging.pluginOutputSize = false

	// This shouldn't go anywhere.
	testMsg := "Test output size entry"
	plugin.logPluginOutputSize(testMsg)

	capturedDebugLogOutput := outputBuffer.String()
	switch {
	case strings.Contains(capturedDebugLogOutput, testMsg):
		want := removeEntry(capturedDebugLogOutput, testMsg, CheckOutputEOL)
		got := capturedDebugLogOutput
		d := cmp.Diff(want, got)
		t.Fatalf("(-want, +got)\n:%s", d)
	default:
		t.Log("OK: No debug logging output captured as expected.")
	}
}

func assertLoggerIsConfiguredProperlyAfterSettingDebugLoggingOutputTarget(plugin *Plugin, t *testing.T) {
	t.Helper()

	// Assert that plugin.logger is set as expected.
	switch {
	case plugin.logger == nil:
		t.Fatal("ERROR: plugin logger is unset instead of being configured for use.")
	default:
		t.Log("OK: plugin logger is set as expected.")
	}

	// Assert that plugin.logger prefix is set as expected.
	actualLoggerPrefix := plugin.logger.Prefix()
	switch {
	case actualLoggerPrefix != logMsgPrefix:
		t.Error("ERROR: plugin logger prefix not set to the expected value.")
		d := cmp.Diff(logMsgPrefix, actualLoggerPrefix)
		t.Fatalf("(-want, +got)\n:%s", d)
	default:
		t.Logf("OK: plugin logger prefix is at the expected value %s", actualLoggerPrefix)
	}

	// Assert that plugin.logger flags is set as expected.
	actualLoggerFlags := plugin.logger.Flags()
	switch {
	case actualLoggerFlags != logFlags:
		t.Error("ERROR: plugin logger flags are not set to the expected value.")
		d := cmp.Diff(logFlags, actualLoggerFlags)
		t.Fatalf("(-want, +got)\n:%s", d)
	default:
		t.Logf("OK: plugin logger flags is set to the expected value %d", actualLoggerFlags)
	}
}

func assertAllDebugLoggingOptionsAreEnabled(plugin *Plugin, t *testing.T) {
	t.Helper()

	// Assert that debug logging is enabled by requiring that all fields are
	// set.
	allDebugLoggingOptionsEnabled := allDebugLoggingOptionsEnabled()

	if !cmp.Equal(
		allDebugLoggingOptionsEnabled,

		// https://stackoverflow.com/questions/73476661/cmp-equal-gives-panic-message-cannot-handle-unexported-field-at
		plugin.debugLogging, cmp.AllowUnexported(debugLoggingOptions{}),
	) {
		d := cmp.Diff(
			allDebugLoggingOptionsEnabled,
			plugin.debugLogging,

			// https://stackoverflow.com/questions/73476661/cmp-equal-gives-panic-message-cannot-handle-unexported-field-at
			cmp.AllowUnexported(debugLoggingOptions{}),
		)

		t.Errorf("(-want, +got)\n:%s", d)
	}
}

func assertAllDebugLoggingOptionsAreDisabled(plugin *Plugin, t *testing.T) {
	t.Helper()

	// Assert that debug logging is enabled by requiring that all fields are
	// set.
	allDebugLoggingOptionsDisabled := allDebugLoggingOptionsDisabled()

	if !cmp.Equal(
		allDebugLoggingOptionsDisabled,

		// https://stackoverflow.com/questions/73476661/cmp-equal-gives-panic-message-cannot-handle-unexported-field-at
		plugin.debugLogging, cmp.AllowUnexported(debugLoggingOptions{}),
	) {
		d := cmp.Diff(
			allDebugLoggingOptionsDisabled,
			plugin.debugLogging,

			// https://stackoverflow.com/questions/73476661/cmp-equal-gives-panic-message-cannot-handle-unexported-field-at
			cmp.AllowUnexported(debugLoggingOptions{}),
		)

		t.Errorf("(-want, +got)\n:%s", d)
	}
}
