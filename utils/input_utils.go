package utils

import (
	"bufio"
	"commander/datamodels"
	"log"
	"os"
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
	var samplePath, outputPath string

	// Read the file line by line.
	for scanner.Scan() {
		line := scanner.Text()
		chunks := strings.Split(line, "=")
		if chunks[0] == "SAMPLE_PATH" {
			samplePath = chunks[1]
		} else if chunks[0] == "OUTPUT_PATH" {
			outputPath = chunks[1]
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
			sample.SamplePath = samplePath
			sample.OutputPath = outputPath
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
	prefix := strings.Split(strings.TrimRight(filename, ".fastq.gz"), "_R")[0]
	return prefix
}
