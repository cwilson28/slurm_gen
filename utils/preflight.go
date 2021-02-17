package utils

import (
	"commander/datamodels"
	"errors"
	"fmt"
	"os"
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
	msgBuffer = append(msgBuffer, "Checking existence of analysis directories...")
	err = testAnalysisDirectory(experiment)
	if err != nil {
		return err
	}
	msgBuffer = append(msgBuffer, "Done.\n")
	msgBuffer = printMsgBuffer(msgBuffer)

	// Test tool directories
	for _, cmd := range job.Commands {
		err = testToolDirectory(experiment, cmd.CommandParams.Command)
		if os.IsNotExist(err) {
			msgBuffer = append(msgBuffer, fmt.Sprintf("Directory %s/%s does not exist.\n", experiment.DumpAnalysisPath(), cmd.CommandParams.Command))
			msgBuffer = append(msgBuffer, "Creating directory... ")
			err = createToolDirectory(experiment, cmd.CommandParams.Command)
			if err != nil {
				return err
			}
			msgBuffer = append(msgBuffer, "Done.\n")
			msgBuffer = printMsgBuffer(msgBuffer)
		}

	}
	return nil
}

func testSampleDirectory(experiment datamodels.Experiment) error {
	_, err := os.Stat(experiment.DumpSamplePath())
	if os.IsNotExist(err) {
		errString := fmt.Sprintf(
			"Directory %s does not exist. \nPlease check that you have specified the sample path correctly.",
			experiment.DumpSamplePath(),
		)
		return errors.New(errString)
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
		if os.IsNotExist(err) {
			notfound = append(notfound, readfile)
		}

		// If reverse read was specified, test for the existence of that file too.
		if s.ReverseReadFile != "" {
			readfile = fmt.Sprintf("%s/%s", experiment.DumpSamplePath(), s.ReverseReadFile)
			_, err := os.Stat(readfile)
			if os.IsNotExist(err) {
				notfound = append(notfound, readfile)
			}
		}
	}

	if len(notfound) > 0 {
		fmt.Println("The following sample files could not be found...\n")
		for _, f := range notfound {
			fmt.Println(f)
		}
		fmt.Println()
		errString := "Missing sample files. Please check that you have specified the sample path correctly."
		err = errors.New(errString)
	}
	return err
}

func testAnalysisDirectory(experiment datamodels.Experiment) error {
	_, err := os.Stat(experiment.DumpAnalysisPath())
	if os.IsNotExist(err) {
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

func testToolDirectory(experiment datamodels.Experiment, tool string) error {
	path := fmt.Sprintf("%s/%s", experiment.DumpAnalysisPath(), tool)
	_, err := os.Stat(path)
	return err
}

func createToolDirectory(experiment datamodels.Experiment, tool string) error {
	path := fmt.Sprintf("%s/%s", experiment.DumpAnalysisPath(), tool)
	err := os.MkdirAll(path, 0755)
	return err
}

func printMsgBuffer(mbuffer []string) []string {
	for _, m := range mbuffer {
		fmt.Printf(m)
	}
	fmt.Println()
	return make([]string, 0)
}
