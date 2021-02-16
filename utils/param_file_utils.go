package utils

import "strings"

/* -----------------------------------------------------------------------------
 * Functions for parsing input parameter file.
 * -------------------------------------------------------------------------- */

func IsJSONParam(filename string) bool {
	chunks := strings.Split(filename, ".")
	extnIndex := len(chunks) - 1
	if chunks[extnIndex] == "json" {
		return true
	}

	return false
}

/* ---
 * If batch is listed in the params file, get the list of commands that will be
 * run in batch mode.
 * --- */
func ParseBatchOpt(batchCommands string) (bool, []string) {
	var commands = make([]string, 0)
	if batchCommands == "" {
		return false, commands
	}

	commands = strings.Split(batchCommands, ",")
	return true, commands
}

/* ---
 * Parse a line from the parameters file.
 * --- */
func ParseLine(l string) (string, string) {
	lineChunks := strings.Split(l, "=")
	if len(lineChunks) == 1 {
		return lineChunks[0], ""
	}
	return lineChunks[0], CleanLine(lineChunks[1])
}

/* ---
 * Clean extraneous characters from an parameter file line.
 * --- */
func CleanLine(l string) string {
	chunks := strings.Split(l, ";;")
	// We want everything up to the comment delimiter
	return strings.TrimRight(chunks[0], " ")
}

/* -----------------------------------------------------------------------------
 * Functions for interpreting lines of the parameters file.
 * -------------------------------------------------------------------------- */

/* ---
 * Determine if a parameter file line should be ignored.
 * --- */
func IgnoreLine(l string) bool {
	if l == "" || strings.Contains(l, "#") {
		return true
	}
	return false
}

/* ---
 * Check if a command is listed as a batch command.
 * --- */
func IsBatchCmd(cmd string, cmds []string) bool {
	for _, c := range cmds {
		if c == cmd {
			return true
		}
	}
	return false
}

/* ---
 * Check if a line is preamble for the job being submitted.
 * --- */
func IsBatchPreamble(tag string) bool {
	if tag == "BATCH" ||
		tag == "SAMPLES_FILE" {
		return true
	}
	return false
}

/* ---
 * Check if a line contains slurm preamble.
 * --- */
func IsSlurmPreamble(tag string) bool {
	if tag == "PARTITION" ||
		tag == "NOTIFICATION_BEGIN" ||
		tag == "NOTIFICATION_END" ||
		tag == "NOTIFICATION_FAIL" ||
		tag == "NOTIFICATION_EMAIL" {
		return true
	}
	return false
}

/* ---
 * Check if a line contains slurm command preamble.
 * --- */
func IsSlurmCommandPreamble(tag string) bool {
	if tag == "BATCH" ||
		tag == "SAMPLES_FILE" ||
		tag == "JOB_NAME" ||
		tag == "TASKS" ||
		tag == "CPUS" ||
		tag == "MEMORY" ||
		tag == "TIME" {
		return true
	}
	return false
}

/* ---
 * Check if a line contains sge preamble.
 * --- */
func IsSGEPreamble(tag string) bool {
	if tag == "PARTITION" ||
		tag == "NOTIFICATION_BEGIN" ||
		tag == "NOTIFICATION_END" ||
		tag == "NOTIFICATION_FAIL" ||
		tag == "NOTIFICATION_EMAIL" {
		return true
	}
	return false
}

/* ---
 * Check if a line contains sge command preamble.
 * --- */
func IsSGECommandPreamble(tag string) bool {
	if tag == "BATCH" ||
		tag == "SAMPLES_FILE" ||
		tag == "JOB_NAME" ||
		tag == "TASKS" ||
		tag == "CPUS" ||
		tag == "MEMORY" ||
		tag == "TIME" {
		return true
	}
	return false
}
