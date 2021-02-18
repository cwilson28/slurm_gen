package utils

import (
	"errors"
	"fmt"
	"os"

	"commander/datamodels"
)

func PreflightTests(experiment datamodels.Experiment, job datamodels.Job) error {
	var err error
	var msgBuffer = make([]string, 0)
	// Test sample file directory
	fmt.Println("Performing pipeline preflight checks...\n")
	msgBuffer = append(msgBuffer, "Checking existence of sample directory... ")
	err = testSampleDirectory(experiment)
	if err != nil {
		return err
	}
	msgBuffer = append(msgBuffer, "Done.\n")
	msgBuffer = printMsgBuffer(msgBuffer)

	// Test sample files
	msgBuffer = append(msgBuffer, "Checking existence of sample files...")
	err = testSampleFiles(experiment)
	if err != nil {
		return err
	}
	msgBuffer = append(msgBuffer, "Done.\n")
	msgBuffer = printMsgBuffer(msgBuffer)

	// Test analysis directory
	msgBuffer = append(msgBuffer, "Checking existence of analysis directory... ")
	err = testAnalysisDirectory(experiment)
	if err != nil {
		return err
	}
	msgBuffer = append(msgBuffer, "Done.\n")
	msgBuffer = printMsgBuffer(msgBuffer)

	// Test tool directories
	fmt.Println("Checking the existence of pipeline output directories...\n")
	for _, cmd := range job.Commands {
		err = testOutputDirectory(experiment, cmd.CommandParams.Command)
		if err != nil {
			return err
		}
	}
	return nil
}

func testSampleDirectory(experiment datamodels.Experiment) error {
	_, err := os.Stat(experiment.DumpSamplePath())
	if err != nil && os.IsNotExist(err) {
		// Format custom message for user.
		errString := fmt.Sprintf(
			"Directory %s does not exist. \nPlease check that you have specified the sample path correctly.",
			experiment.DumpSamplePath(),
		)
		return errors.New(errString)
	} else if err != nil {
		return err
	}
	return nil
}

func testSampleFiles(experiment datamodels.Experiment) error {
	var err error
	var readfile string
	var notfound = make([]string, 0)

	// Test for the existence of each sample file. This test uses absolute paths.
	for _, s := range experiment.Samples {
		// First test for the existence of the forward read file.
		readfile = fmt.Sprintf("%s/%s", experiment.DumpSamplePath(), s.ForwardReadFile)
		_, err := os.Stat(readfile)
		if err != nil && os.IsNotExist(err) {
			notfound = append(notfound, readfile)
		} else if err != nil {
			return err
		}

		// If reverse read was specified, test for the existence of that file too.
		if s.ReverseReadFile != "" {
			readfile = fmt.Sprintf("%s/%s", experiment.DumpSamplePath(), s.ReverseReadFile)
			_, err := os.Stat(readfile)
			if err != nil && os.IsNotExist(err) {
				notfound = append(notfound, readfile)
			} else if err != nil {
				return err
			}
		}
	}

	if len(notfound) > 0 {
		// Echo the files that were not found
		fmt.Println("The following sample files could not be found...\n")
		for _, f := range notfound {
			fmt.Println(f)
		}
		fmt.Println()
		// Format a custom error for user.
		errString := "Missing sample files. Please check that you have specified the sample path correctly."
		err = errors.New(errString)
	}
	return err
}

func testAnalysisDirectory(experiment datamodels.Experiment) error {
	_, err := os.Stat(experiment.DumpAnalysisPath())
	if err != nil && os.IsNotExist(err) {
		// Notify the user the directory does not exist and that we will create
		// the directory.
		fmt.Printf("Directory %s does not exist.\n", experiment.DumpAnalysisPath())
		fmt.Printf("Creating directory... ")
		err = createAnalysisDirectory(experiment)
		if err != nil {
			return err
		}
		fmt.Println("Done\n")
	}
	// If the directory exists, check permissions.
	return nil
}

func createAnalysisDirectory(experiment datamodels.Experiment) error {
	err := os.MkdirAll(experiment.DumpAnalysisPath(), 0755)
	return err
}

func testOutputDirectory(experiment datamodels.Experiment, tool string) error {
	msgBuffer := newMsgBuffer()

	path := fmt.Sprintf("%s/%s", experiment.DumpAnalysisPath(), tool)
	dirInfo, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		// The directory does not exist. Try to create it on user's behalf.
		msgBuffer = append(msgBuffer, fmt.Sprintf("Output directory %s does not exist.\n", path))
		msgBuffer = append(msgBuffer, "Creating directory... ")
		err = createAnalysisDirectory(experiment)
		if err != nil {
			return err
		}
		msgBuffer = append(msgBuffer, "Done.\n")
		msgBuffer = printMsgBuffer(msgBuffer)

	} else if err != nil {
		// Something went really wrong.
		return err
	}

	// Check that path is to a directory
	if dirInfo.IsDir() {
		msgBuffer = append(msgBuffer, fmt.Sprintf("Path to output directory %s exists.\n", path))
	} else {
		msgBuffer = append(msgBuffer, fmt.Sprintf("Path to output directory %s exists but is not a directory.\n", path))
		errString := fmt.Sprintf("Directory error. Please verify the path to output directory %s.", path)
		err = errors.New(errString)
	}
	msgBuffer = printMsgBuffer(msgBuffer)
	if err != nil {
		return err
	}

	// Path is a directory, test write permissions
	msgBuffer = append(msgBuffer, "Testing output directory write permissions... ")
	err = createTestFile(path)
	if err != nil {
		msgBuffer = append(msgBuffer, "\n")
		msgBuffer = append(msgBuffer, fmt.Sprintf("Output directory is not writeable. Permissions are %s\n", dirInfo.Mode().Perm()))
		// Trigger an error.
		errString := fmt.Sprintf("Permission error. Please check that you have correct privleges on %s.", path)
		err = errors.New(errString)
	} else {
		msgBuffer = append(msgBuffer, "Done.\n")
	}
	msgBuffer = printMsgBuffer(msgBuffer)
	return err
}

func createToolDirectory(experiment datamodels.Experiment, tool string) error {
	path := fmt.Sprintf("%s/%s", experiment.DumpAnalysisPath(), tool)
	err := os.MkdirAll(path, 0755)
	return err
}

func createTestFile(path string) error {
	testfile := fmt.Sprintf("%s/test.txt", path)
	// Use os.Create to create a file for writing.
	_, err := os.Create(testfile)
	if err != nil {
		return err
	}

	err = os.Remove(testfile)
	return err
}

func newMsgBuffer() []string {
	return make([]string, 0)
}

func printMsgBuffer(mbuffer []string) []string {
	for _, m := range mbuffer {
		fmt.Printf(m)
	}
	fmt.Println()
	return make([]string, 0)
}
