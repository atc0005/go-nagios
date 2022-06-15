// Copyright 2020 Adam Chalkley
//
// https://github.com/atc0005/go-nagios
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

package nagios

import (
	"fmt"
	"os"
	"runtime/debug"
)

// Nagios plugin/service check states. These constants replicate the values
// from utils.sh which is normally found at one of these two locations,
// depending on which Linux distribution you're using:
//
//     /usr/lib/nagios/plugins/utils.sh
//     /usr/local/nagios/libexec/utils.sh
//
// See also http://nagios-plugins.org/doc/guidelines.html
const (
	StateOKExitCode        int = 0
	StateWARNINGExitCode   int = 1
	StateCRITICALExitCode  int = 2
	StateUNKNOWNExitCode   int = 3
	StateDEPENDENTExitCode int = 4
)

// Nagios plugin/service check state "labels". These constants are provided as
// an alternative to using literal state strings throughout client application
// code.
const (
	StateOKLabel        string = "OK"
	StateWARNINGLabel   string = "WARNING"
	StateCRITICALLabel  string = "CRITICAL"
	StateUNKNOWNLabel   string = "UNKNOWN"
	StateDEPENDENTLabel string = "DEPENDENT"
)

// CheckOutputEOL is the newline character(s) used with formatted service and
// host check output. Based on previous testing, Nagios treats LF newlines
// (without a leading space) within the `$LONGSERVICEOUTPUT$` macro as literal
// values instead of parsing them for display purposes.
//
// Using DOS EOL values with `fmt.Printf()` gave expected formatting results
// in the Nagios Core web UI, but resulted in double newlines in Nagios XI
// output (see GH-109). Switching back to a UNIX EOL with a single leading
// space appears to give the intended results for both Nagios Core and Nagios
// XI.
const CheckOutputEOL string = " \n"

// Default header text for various sections of the output if not overridden.
const (
	defaultThresholdsLabel   string = "THRESHOLDS"
	defaultErrorsLabel       string = "ERRORS"
	defaultDetailedInfoLabel string = "DETAILED INFO"
)

// ServiceState represents the status label and exit code for a service check.
type ServiceState struct {

	// Label maps directly to one of the supported Nagios state labels.
	Label string

	// ExitCode is the exit or exit status code associated with a Nagios
	// service check.
	ExitCode int
}

// PerformanceData represents the performance data generated by a Nagios
// plugin.
//
// Plugin performance data is external data specific to the plugin used to
// perform the host or service check. Plugin-specific data can include things
// like percent packet loss, free disk space, processor load, number of
// current users, etc. - basically any type of metric that the plugin is
// measuring when it executes.
type PerformanceData struct {

	// Label is the single quoted text string used as a label for a specific
	// performance data point. The label length is arbitrary, but ideally the
	// first 19 characters are unique due to a limitation in RRD. There is
	// also a limitation in the amount of data that NRPE returns to Nagios.
	//
	// The popular convention used by plugin authors (and official
	// documentation) is to use underscores for separating multiple words. For
	// example, 'percent_packet_loss' instead of 'percent packet loss',
	// 'percentPacketLoss' or 'percent-packet-loss.
	Label string

	// Value is the data point associated with the performance data label.
	//
	// Value is in class [-0-9.] and must be the same UOM as Min and Max UOM.
	// Value may be a literal "U" instead, this would indicate that the actual
	// value couldn't be determined.
	Value string

	// UnitOfMeasurement is an optional unit of measurement (UOM). If
	// provided, consists of a string of zero or more characters. Numbers,
	// semicolons or quotes are not permitted.
	//
	// Examples:
	//
	// 1) no unit specified - assume a number (int or float) of things (eg,
	// users, processes, load averages)
	// 2) s - seconds (also us, ms)
	// 3) % - percentage
	// 4) B - bytes (also KB, MB, TB)
	// 5) c - a continuous counter (such as bytes transmitted on an interface)
	UnitOfMeasurement string

	// Warn is in the range format (see the Section called Threshold and
	// Ranges). Must be the same UOM as Crit. An empty string is permitted.
	//
	// https://nagios-plugins.org/doc/guidelines.html#THRESHOLDFORMAT
	Warn string

	// Crit is in the range format (see the Section called Threshold and
	// Ranges). Must be the same UOM as Warn. An empty string is permitted.
	//
	// https://nagios-plugins.org/doc/guidelines.html#THRESHOLDFORMAT
	Crit string

	// Min is in class [-0-9.] and must be the same UOM as Value and Max. Min
	// is not required if UOM=%. An empty string is permitted.
	Min string

	// Max is in class [-0-9.] and must be the same UOM as Value and Min. Max
	// is not required if UOM=%. An empty string is permitted.
	Max string
}

// Validate performs basic validation of PerformanceData. An error is returned
// for any validation failures.
func (pd PerformanceData) Validate() error {

	// Validate fields
	switch {
	case pd.Label == "":
		return fmt.Errorf("provided performance data missing required label")
	case pd.Value == "":
		return fmt.Errorf("provided performance data missing required value")

	// TODO: Expand validation
	// https://nagios-plugins.org/doc/guidelines.html
	default:
		return nil

	}
}

// ExitCallBackFunc represents a function that is called as a final step
// before application termination so that branding information can be emitted
// for inclusion in the notification. This helps identify which specific
// application (and its version) that is responsible for the notification.
type ExitCallBackFunc func() string

// ExitState represents the last known execution state of the
// application, including the most recent error and the final intended plugin
// state.
type ExitState struct {

	// LastError is the last error encountered which should be reported as
	// part of ending the service check (e.g., "Failed to connect to XYZ to
	// check contents of Inbox").
	//
	// Deprecated: Use Errors field or AddError method instead.
	LastError error

	// Errors is a collection of one or more recorded errors to be displayed
	// in LongServiceOutput as a list when ending the service check.
	Errors []error

	// ExitStatusCode is the exit or exit status code provided to the Nagios
	// instance that calls this service check. These status codes indicate to
	// Nagios "state" the service is considered to be in. The most common
	// states are OK (0), WARNING (1) and CRITICAL (2).
	ExitStatusCode int

	// ServiceOutput is the first line of text output from the last service
	// check (i.e. "Ping OK").
	ServiceOutput string

	// LongServiceOutput is the full text output (aside from the first line)
	// from the last service check.
	LongServiceOutput string

	// perfData is the collection of zero or more PerformanceData values
	// generated by a Nagios plugin.
	perfData []PerformanceData

	// WarningThreshold is the value used to determine when the service check
	// has crossed between an existing state into a WARNING state. This value
	// is used for display purposes.
	WarningThreshold string

	// CriticalThreshold is the value used to determine when the service check
	// has crossed between an existing state into a CRITICAL state. This value
	// is used for display purposes.
	CriticalThreshold string

	// thresholdLabel is an optional custom label used in place of the
	// standard text prior to a list of threshold values.
	thresholdsLabel string

	// errorsLabel is an optional custom label used in place of the standard
	// text prior to a list of recorded error values.
	errorsLabel string

	// detailedInfoLabel is an optional custom label used in place of the
	// standard text prior to emitting LongServiceOutput.
	detailedInfoLabel string

	// BrandingCallback is a function that is called before application
	// termination to emit branding details at the end of the notification.
	// See also ExitCallBackFunc.
	BrandingCallback ExitCallBackFunc
}

// ReturnCheckResults is intended to provide a reliable way to return a
// desired exit code from applications used as Nagios plugins. In most cases,
// this method should be registered as the first deferred function in client
// code. See remarks regarding "masking" or "swallowing" application panics.
//
// Since Nagios relies on plugin exit codes to determine success/failure of
// checks, the approach that is most often used with other languages is to use
// something like Using os.Exit() directly and force an early exit of the
// application with an explicit exit code. Using os.Exit() directly in Go does
// not run deferred functions. Go-based plugins that do not rely on deferring
// function calls may be able to use os.Exit(), but introducing new
// dependencies later could introduce problems if those dependencies rely on
// deferring functions.
//
// Before calling this method, client code should first set appropriate field
// values on the receiver. When called, this method will process them and exit
// with the desired exit code and status output.
//
// To repeat, if scheduled via defer, this method should be registered first;
// because this method calls os.Exit to set the intended plugin exit state, no
// other deferred functions will have an opportunity to run, so register this
// method first so that when deferred, it will be run last (FILO).
//
// Because this method is (or should be) deferred first within client code, it
// will run after all other deferred functions. It will also run before a
// panic in client code forces the application to exit. As already noted, this
// method calls os.Exit to set the plugin exit state. Because os.Exit forces
// the application to terminate immediately without running other deferred
// functions or processing panics, this "masks", "swallows" or "blocks" panics
// from client code from surfacing. This method checks for unhandled panics
// and if found, overrides exit state details from client code and surfaces
// details from the panic instead as a CRITICAL state.
func (es *ExitState) ReturnCheckResults() {

	// Check for unhandled panic in client code. If present, override
	// ExitState and make clear that the client code/plugin crashed.
	if err := recover(); err != nil {

		es.AddError(fmt.Errorf("plugin crash/panic detected: %s", err))

		es.ServiceOutput = fmt.Sprintf(
			"%s: plugin crash detected. See details via web UI or run plugin manually via CLI.",
			StateCRITICALLabel,
		)

		// Gather stack trace associated with panic.
		stackTrace := debug.Stack()

		// Wrap stack trace details in an attempt to prevent these details
		// from being interpreted as formatting characters when passed through
		// web UI, text, email, Teams, etc. We use Markdown fenced code blocks
		// instead of `<pre>` start/end tags because Nagios strips out angle
		// brackets (due to default `illegal_macro_output_chars` settings).
		es.LongServiceOutput = fmt.Sprintf(
			"```%s%s%s%s%s%s```",
			CheckOutputEOL,
			err,
			CheckOutputEOL,
			CheckOutputEOL,
			stackTrace,
			CheckOutputEOL,
		)

		es.ExitStatusCode = StateCRITICALExitCode

	}

	// ##################################################################
	// Note: fmt.Println() has the same issue as `\n`: Nagios seems to
	// interpret them literally instead of emitting an actual newline.
	// We work around that by using fmt.Printf() and fmt.Print() for
	// output that is intended for display within the Nagios web UI.
	// ##################################################################

	// One-line output used as the summary or short explanation for the
	// specific Nagios state that we are returning. We apply no formatting
	// changes to this content, simply emit it as-is. This helps avoid
	// potential issues with literal characters being interpreted as
	// formatting verbs.
	fmt.Print(es.ServiceOutput)

	if es.LongServiceOutput != "" || es.LastError != nil || len(es.Errors) > 0 {

		fmt.Printf(
			"%s%s**%s**%s",
			CheckOutputEOL,
			CheckOutputEOL,
			es.getErrorsLabelText(),
			CheckOutputEOL,
		)

		// If an error occurred or if there are additional details to share ...

		if es.LastError != nil {
			fmt.Printf("* %v%s", es.LastError, CheckOutputEOL)
		}

		if len(es.Errors) > 0 {
			for _, err := range es.Errors {
				if err != nil {
					fmt.Printf("* %v%s", err, CheckOutputEOL)
				}
			}
		}

		if es.LastError == nil && len(es.Errors) == 0 {
			fmt.Printf("%s* None%s", CheckOutputEOL, CheckOutputEOL)
		}

		if es.LongServiceOutput != "" {

			fmt.Printf(
				"%s**%s**%s",
				CheckOutputEOL,
				es.getThresholdsLabelText(),
				CheckOutputEOL,
			)

			if es.CriticalThreshold != "" || es.WarningThreshold != "" {

				fmt.Print(CheckOutputEOL)

				if es.CriticalThreshold != "" {
					fmt.Printf(
						"* %s: %v%s",
						StateCRITICALLabel,
						es.CriticalThreshold,
						CheckOutputEOL,
					)
				}

				if es.WarningThreshold != "" {
					fmt.Printf(
						"* %s: %v%s",
						StateWARNINGLabel,
						es.WarningThreshold,
						CheckOutputEOL,
					)
				}
			} else {
				fmt.Printf("%s* Not specified%s", CheckOutputEOL, CheckOutputEOL)
			}

			fmt.Printf(
				"%s**%s**%s",
				CheckOutputEOL,
				es.getDetailedInfoLabelText(),
				CheckOutputEOL,
			)

			// Note: fmt.Println() has the same issue as `\n`: Nagios seems to
			// interpret them literally instead of emitting an actual newline.
			// We work around that by using fmt.Printf() for output that is
			// intended for display within the Nagios web UI.
			fmt.Printf(
				"%s%v%s",
				CheckOutputEOL,
				es.LongServiceOutput,
				CheckOutputEOL,
			)
		}

	}

	// If set, call user-provided branding function before emitting
	// performance data and exiting application.
	if es.BrandingCallback != nil {
		fmt.Printf("%s%s%s", CheckOutputEOL, es.BrandingCallback(), CheckOutputEOL)
	}

	// Generate formatted performance data if provided. Only emit if a
	// one-line summary is set by client code.
	if len(es.perfData) != 0 && es.ServiceOutput != "" {

		// Performance data metrics are appended to plugin output. These
		// metrics are provided as a single line, leading with a pipe
		// character, a space and one or more metrics each separated from
		// another by a single space.
		fmt.Print(" |")

		for _, pd := range es.perfData {
			fmt.Printf(
				// The expected format of a performance data metric:
				//
				// 'label'=value[UOM];[warn];[crit];[min];[max]
				//
				// References:
				//
				// https://nagios-plugins.org/doc/guidelines.html
				// https://assets.nagios.com/downloads/nagioscore/docs/nagioscore/3/en/perfdata.html
				// https://assets.nagios.com/downloads/nagioscore/docs/nagioscore/3/en/pluginapi.html
				// https://www.monitoring-plugins.org/doc/guidelines.html
				// https://icinga.com/docs/icinga-2/latest/doc/05-service-monitoring/#performance-data-metrics
				" '%s'=%s%s;%s;%s;%s;%s",
				pd.Label,
				pd.Value,
				pd.UnitOfMeasurement,
				pd.Warn,
				pd.Crit,
				pd.Min,
				pd.Max,
			)
		}

		// Add final trailing newline to satisfy Nagios plugin output format.
		fmt.Print(CheckOutputEOL)
	}

	os.Exit(es.ExitStatusCode)
}

// getThresholdsLabelText retrieves the custom thresholds label text if set,
// otherwise returns the default value.
func (es ExitState) getThresholdsLabelText() string {
	switch {
	case es.thresholdsLabel != "":
		return es.thresholdsLabel
	default:
		return defaultThresholdsLabel
	}
}

// getErrorsLabelText retrieves the custom errors label text if set, otherwise
// returns the default value.
func (es ExitState) getErrorsLabelText() string {
	switch {
	case es.errorsLabel != "":
		return es.errorsLabel
	default:
		return defaultErrorsLabel
	}
}

// getErrorsLabelText retrieves the custom detailed info label text if set,
// otherwise returns the default value.
func (es ExitState) getDetailedInfoLabelText() string {
	switch {
	case es.detailedInfoLabel != "":
		return es.detailedInfoLabel
	default:
		return defaultDetailedInfoLabel
	}
}

// AddPerfData appends provided performance data. Validation is skipped if
// requested, otherwise an error is returned if validation fails. Validation
// failure results in no performance data being appended.
//
// Client code may wish to disable validation if performing this step
// directly.
func (es *ExitState) AddPerfData(skipValidate bool, pd ...PerformanceData) error {

	if len(pd) == 0 {
		return fmt.Errorf("no performance data provided")
	}

	if !skipValidate {
		for i := range pd {
			if err := pd[i].Validate(); err != nil {
				return err
			}
		}
	}

	es.perfData = append(es.perfData, pd...)

	return nil

}

// AddError appends provided errors to the collection.
func (es *ExitState) AddError(err ...error) {
	es.Errors = append(es.Errors, err...)
}

// SetThresholdsLabel overrides the default thresholds label text.
func (es *ExitState) SetThresholdsLabel(newLabel string) {
	es.thresholdsLabel = newLabel
}

// SetErrorsLabel overrides the default errors label text.
func (es *ExitState) SetErrorsLabel(newLabel string) {
	es.errorsLabel = newLabel
}

// SetDetailedInfoLabel overrides the default detailed info label text.
func (es *ExitState) SetDetailedInfoLabel(newLabel string) {
	es.detailedInfoLabel = newLabel
}
