package utils

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"commander/datamodels"
)

/* -----------------------------------------------------------------------------
 * Main preflight test routine.
 * -------------------------------------------------------------------------- */
func PreflightTests(job datamodels.Job) error {
	var err error
	experiment := job.ExperimentDetails

	// Test sample file directory
	fmt.Println("Performing pipeline preflight checks...\n")
	err = testSampleDirectory(experiment)
	if err != nil {
		return err
	}

	// Test existence of sample files.
	if len(experiment.Samples) > 0 {
		fmt.Println("Checking existence of sample files... \n")
		err = testSampleFiles(experiment)
		if err != nil {
			return err
		}
	}

	// Test analysis directory
	fmt.Printf("Checking existence of analysis directory %s... \n\n", experiment.PrintAnalysisPath())
	err = testAnalysisDirectory(experiment)
	if err != nil {
		return err
	}

	// Test working directory
	if experiment.WorkDir != "" {
		fmt.Printf("Checking existence of working directory %s... \n\n", experiment.PrintAnalysisPath())
		err = testAnalysisDirectory(experiment)
		if err != nil {
			return err
		}
	}

	// Test tool directories
	fmt.Println("Checking the existence of pipeline tool directories...\n")
	for _, cmd := range job.Commands {
		// Check for the existence of the tool output directory
		err := testToolOutputDirectory(experiment, cmd.CommandParams.Command)
		if err != nil {
			return err
		}
	}

	if len(job.CleanupActions) > 0 {
		fmt.Println("Checking the existence of cleanup action destination paths...\n")
		for _, a := range job.CleanupActions {
			baseDir := job.ExperimentDetails.PrintAnalysisPath()
			sourcePath := fmt.Sprintf("%s/%s", baseDir, a.ToolName)
			destPath := fmt.Sprintf("%s/%s", baseDir, a.Destination)
			err = testCleanupPaths(a.Action, sourcePath, destPath)
			if err != nil {
				return err
			}
		}

	}

	fmt.Println("Checking the existence of configuration archive directories...\n")
	// Test configuration archive directory.
	err = testArchiveDirectory(experiment)
	if err != nil {
		return err
	}

	// Test logging directory.
	err = testLoggingDirectory(experiment)
	if err != nil {
		return err
	}
	return nil
}

/* -----------------------------------------------------------------------------
 * Preflight configuration preservation.
 * -------------------------------------------------------------------------- */
func ArchiveParamFile(paramFile string, experiment datamodels.Experiment) (int64, error) {
	var paramFileName string

	fmt.Printf("Archiving parameter file... \n")
	// Get paramfile name.
	if strings.Contains(paramFile, "/") {
		chunks := strings.Split(paramFile, "/")
		paramFileName = chunks[len(chunks)-1]
	} else {
		paramFileName = paramFile
	}

	// Create the param file destination!
	paramDstPath := fmt.Sprintf("%s/config/%s", experiment.PrintAnalysisPath(), paramFileName)
	paramFileSrc, err := os.Open(paramFile)
	if err != nil {
		return 0, err
	}
	defer paramFileSrc.Close()

	paramFileDst, err := os.Create(paramDstPath)
	if err != nil {
		return 0, err
	}
	defer paramFileDst.Close()

	nBytes, err := io.Copy(paramFileDst, paramFileSrc)
	if err != nil {
		return 0, err
	}
	fmt.Printf("Done.\n\n")
	return nBytes, nil
}

func ArchiveDesignFile(experiment datamodels.Experiment) (int64, error) {
	var designFileName string
	designFile := experiment.SamplesFile

	fmt.Printf("Archiving design file... \n")
	// Get the designfile name.
	if strings.Contains(designFile, "/") {
		chunks := strings.Split(designFile, "/")
		designFileName = chunks[len(chunks)-1]
	} else {
		designFileName = designFile
	}

	// Create the param file destination!
	designDstPath := fmt.Sprintf("%s/config/%s", experiment.PrintAnalysisPath(), designFileName)
	designFileSrc, err := os.Open(designFile)
	if err != nil {
		return 0, err
	}
	defer designFileSrc.Close()

	designFileDst, err := os.Create(designDstPath)
	if err != nil {
		return 0, err
	}
	defer designFileDst.Close()

	nBytes, err := io.Copy(designFileDst, designFileSrc)
	if err != nil {
		return 0, err
	}
	fmt.Printf("Done.\n\n")
	return nBytes, nil
}

/* -----------------------------------------------------------------------------
 * Local test helpers.
 * -------------------------------------------------------------------------- */

/* ---
 * Test for the existence of the sample file directory
 * --- */
func testSampleDirectory(experiment datamodels.Experiment) error {
	_, err := os.Stat(experiment.PrintRawSamplePath())
	if err != nil && os.IsNotExist(err) {
		// Format custom message for user.
		errString := fmt.Sprintf(
			"Sample directory %s does not exist. \nPlease check that you have specified the sample path correctly.",
			experiment.PrintRawSamplePath(),
		)
		return errors.New(errString)
	} else if err != nil {
		return err
	}
	fmt.Printf("Sample directory %s exists.\n\n", experiment.PrintRawSamplePath())
	return nil
}

/* ---
 * Test for the existence of the individual sample files.
 * --- */
func testSampleFiles(experiment datamodels.Experiment) error {
	var err error
	var readfile string
	var notfound = make([]string, 0)

	// Test for the existence of each sample file. This test uses absolute paths.
	for _, s := range experiment.Samples {
		// First test for the existence of the forward read file.
		readfile = fmt.Sprintf("%s/%s", experiment.PrintRawSamplePath(), s.ForwardReadFile)
		_, err := os.Stat(readfile)
		if err != nil && os.IsNotExist(err) {
			notfound = append(notfound, readfile)
		} else if err != nil {
			return err
		}

		// If reverse read was specified, test for the existence of that file too.
		if s.ReverseReadFile != "" {
			readfile = fmt.Sprintf("%s/%s", experiment.PrintRawSamplePath(), s.ReverseReadFile)
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

/* ---
 * Check for the existence of the analysis directory.
 * This directory follows the convention /compbio/analysis/<PI>/<experiment>
 * --- */
func testAnalysisDirectory(experiment datamodels.Experiment) error {

	_, err := os.Stat(experiment.PrintAnalysisPath())
	if err != nil && os.IsNotExist(err) {
		// Notify the user the directory does not exist and that we will create
		// the directory.
		fmt.Printf("Directory %s does not exist.\n", experiment.PrintAnalysisPath())
		fmt.Printf("Creating directory...\n\n")
		err = os.MkdirAll(experiment.PrintAnalysisPath(), 0755)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		fmt.Println("Directory exists.\n")
	}
	return nil
}

/* ---
 * Check for the existence of the working directory.
 * This directory does not follow a specific convention.
 * --- */
func testWorkingDirectory(experiment datamodels.Experiment) error {
	_, err := os.Stat(experiment.PrintWorkingDirectory())
	if err != nil && os.IsNotExist(err) {
		// Notify the user the directory does not exist and that we will create
		// the directory.
		fmt.Printf("Directory %s does not exist.\n", experiment.PrintWorkingDirectory())
		fmt.Printf("Creating directory...\n\n")
		err = os.MkdirAll(experiment.PrintWorkingDirectory(), 0755)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		fmt.Println("Directory exists.\n")
	}
	return nil
}

/* ---
 * Check for the existence of the tool output directory.
 * This directory follows the convention /compbio/analysis/<PI>/<experiment>/<tool_name>
 * --- */
func testToolOutputDirectory(experiment datamodels.Experiment, tool string) error {

	path := fmt.Sprintf("%s/%s", experiment.PrintAnalysisPath(), tool)
	dirInfo, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		// The directory does not exist. Try to create it on user's behalf.
		fmt.Printf("Directory %s does not exist.\n", path)
		fmt.Printf("Creating directory...\n\n")
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		// Something went really wrong.
		return err
	}

	// Check that path is to a directory
	if !dirInfo.IsDir() {
		fmt.Printf("Path to output directory %s exists but is not a directory.\n", path)
		errString := fmt.Sprintf("Directory error. Please verify the path to output directory %s.", path)
		err = errors.New(errString)
		return err
	}

	// Path is a directory, test write permissions
	fmt.Printf("Testing write permissions on directory %s\n\n", path)
	err = createTestFile(path)
	if err != nil {
		fmt.Printf("Directory is not writeable. Permissions are %s\n\n", dirInfo.Mode().Perm())
		// Trigger an error.
		errString := fmt.Sprintf("Permission error. Please check that you have correct privleges on %s.", path)
		err = errors.New(errString)
	}
	return err
}

/* ---
 * Check for the existence of a cleanup action destination directory.
 * --- */
func testCleanupPaths(action, sourcePath, destPath string) error {

	cleanupAction := fmt.Sprintf("%s %s %s", action, sourcePath, destPath)
	// Check for the existence of the config directory.
	_, err := os.Stat(sourcePath)
	if err != nil && os.IsNotExist(err) {
		// Notify the user the directory does not exist and that we will create
		// the directory.
		fmt.Printf("Source directory for cleanup action '%s' does not exist.\n", cleanupAction)
		fmt.Printf("Creating directory... \n\n")
		err = os.MkdirAll(sourcePath, 0755)
		if err != nil {
			return err
		}
	}

	_, err = os.Stat(destPath)
	if err != nil && os.IsNotExist(err) {
		// Notify the user the directory does not exist and that we will create
		// the directory.
		fmt.Printf("Destination directory for cleanup action '%s' does not exist.\n", cleanupAction)
		fmt.Printf("Creating directory... \n\n")
		err = os.MkdirAll(destPath, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

/* ---
 * Check for the existence of the archive directory.
 * This directory follows the convention /compbio/analysis/<PI>/<experiment>/<analysis_id>/config
 * --- */
func testArchiveDirectory(experiment datamodels.Experiment) error {
	archivePath := fmt.Sprintf("%s/config", experiment.PrintAnalysisPath())

	// Check for the existence of the config directory.
	_, err := os.Stat(archivePath)
	if err != nil && os.IsNotExist(err) {
		// Notify the user the directory does not exist and that we will create
		// the directory.
		fmt.Printf("Archive directory %s does not exist.\n", archivePath)
		fmt.Printf("Creating directory... \n")
		err = os.MkdirAll(archivePath, 0755)
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	}
	return nil
}

/* ---
 * Check for the existence of the logs directory.
 * This directory follows the convention /compbio/analysis/<PI>/<experiment>/<analysis_id>/logs
 * --- */
func testLoggingDirectory(experiment datamodels.Experiment) error {
	msgBuffer := newMsgBuffer()
	logPath := fmt.Sprintf("%s/logs", experiment.PrintAnalysisPath())

	// Check for the existence of the config directory.
	_, err := os.Stat(logPath)
	if err != nil && os.IsNotExist(err) {
		// Notify the user the directory does not exist and that we will create
		// the directory.
		msgBuffer = append(msgBuffer, fmt.Sprintf("Logging directory %s does not exist.\n", logPath))
		msgBuffer = append(msgBuffer, fmt.Sprintf("Creating directory... "))
		err = os.MkdirAll(logPath, 0755)
		if err != nil {
			return err
		}
		msgBuffer = append(msgBuffer, "Done\n")
		msgBuffer = printMsgBuffer(msgBuffer)
		return nil
	} else if err != nil {
		return err
	}
	return nil
}

/* ---
 * Create a test file in an output path.
 * This is used to test permissions on an output directory.
 * --- */
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
