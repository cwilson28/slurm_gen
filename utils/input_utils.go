package utils

import (
	"bufio"
	"log"
	"os"
	"slurm_gen/datamodels"
	"strings"
)

/* -----------------------------------------------------------------------------
 * Functions for parsing samples files.
 * -------------------------------------------------------------------------- */
func ParseSamplesFile(filename string) []datamodels.Sample {
	samples := make([]datamodels.Sample, 0)

	// Open the file for buffer based read.
	fileBuf, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	// Defer file handle closing.
	defer func() {
		if err = fileBuf.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	// Create a file scanner for reading the lines of the file.
	scanner := bufio.NewScanner(fileBuf)
	sample := datamodels.Sample{}

	// Read the file line by line.
	for scanner.Scan() {
		line := scanner.Text()
		chunks := strings.Split(line, "=")
		if chunks[0] == "SAMPLE_PATH" {
			sample.Path = chunks[1]
		} else if chunks[0] == "SAMPLE" {
			// We are assuming the files are separated with a space.
			fileNames := strings.Split(chunks[1], " ")

			// Get the file prefix from the forward read
			sample.Prefix = ParseSamplePrefix(fileNames[0])

			// Append the forward reads.
			sample.ForwardReadFile = fileNames[0]

			// Append the reverse reads if provided.
			if len(fileNames) == 2 {
				sample.ReverseReadFile = fileNames[1]
			}
			samples = append(samples, sample)
			sample = datamodels.Sample{}
		}
	}
	return samples
}

/* ---
 * Parse the sample prefix from a read file.
 * --- */
func ParseSamplePrefix(sampleFileName string) string {
	chunks := strings.Split(sampleFileName, "/")
	filename := chunks[len(chunks)-1]
	prefix := strings.Split(strings.TrimRight(filename, ".fastq.gz"), "_R1")[0]
	return prefix
}

/* -----------------------------------------------------------------------------
 * Functions for parsing input parameter file.
 * -------------------------------------------------------------------------- */

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
 * Check if a line is slurm preamble.
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
 * Check if a line is command preamble.
 * --- */
func IsCommandPreamble(tag string) bool {
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
