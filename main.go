// Copyright 2020 Adam Chalkley
//
// https://github.com/atc0005/go-nagios
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

// Package nagios is a small collection of common types and package-level
// variables intended for use with various plugins to reduce code duplication.
package nagios

// State is a map of specific Nagios plugin/service check states.
// This map replicates the values from utils.sh which is normally found at one
// of these two locations:
//
// /usr/lib/nagios/plugins/utils.sh
// /usr/local/nagios/libexec/utils.sh
var State = map[string]int{
	"OK":        0,
	"WARNING":   1,
	"CRITICAL":  2,
	"UNKNOWN":   3,
	"DEPENDENT": 4,
}
