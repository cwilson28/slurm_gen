package utils

import (
	"bufio"
	"commander/datamodels"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/Jeffail/gabs"
)

var Platform string

/* -----------------------------------------------------------------------------
 * Generate job object from JSON.
 * -------------------------------------------------------------------------- */

func ParseJSONParams(filename string) (datamodels.Job, error) {
	var err error
	var job datamodels.Job

	fmt.Printf("Parsing JSON parameter file... \n\n")

	rawJSON, err := ioutil.ReadFile(filename)
	if err != nil {
		return job, err
	}

	// Parse the raw json.
	jsonParsed, err := gabs.ParseJSON(rawJSON)
	if err != nil {
		return job, err
	}

	// Extract and set the job details from the json file.
	jobDetails, err := jobDetailsFromJSON(jsonParsed)
	if err != nil {
		if err != nil {
			return job, err
		}
	}
	job.Details = jobDetails

	// Extract and set experiment details from the json file.
	// If no experiment block is provided, this will initialize a default experiment.
	job.ExperimentDetails = experimentDetailsFromJSON(jsonParsed)

	// Extract and set any platform specific preamble.
	if Platform == "slurm" {
		slurmPreamble, err := slurmPreambleFromJSON(jsonParsed.Path("slurm_preamble"))
		if err != nil {
			return job, err
		}
		job.SlurmPreamble = slurmPreamble
	} else if Platform == "sge" {
		sgePreamble, err := sgePreambleFromJSON(jsonParsed.Path("sge_preamble"))
		if err != nil {
			return job, err
		}
		job.SGEPreamble = sgePreamble
	}

	// Extract and set any miscellaneous preamble.
	miscPreamble, err := miscPreambleFromJSON(jsonParsed)
	if err != nil {
		if err != nil {
			return job, err
		}
	}
	job.MiscPreamble = miscPreamble

	// Extract the commands from the params file.
	var cmdErr error
	var commands = make([]datamodels.Command, 0)
	for _, c := range jsonParsed.Path("commands").Children() {
		// Initialize empty command obj.
		command := datamodels.Command{}

		// Set batch argument.
		command.Batch, cmdErr = isBatchCommand(c)
		if cmdErr != nil {
			return job, cmdErr
		}

		// Set input_from argument.
		command.InputFromStep = inputFromStep(c)

		// Extract and set command preamble.
		preamble, cmdErr := commandPreambleFromJSON(c)
		if cmdErr != nil {
			return job, cmdErr
		}
		command.Preamble = preamble

		// Extract and set the command params.
		params, cmdErr := commandParamsFromJSON(c)
		if cmdErr != nil {
			return job, cmdErr
		}
		command.CommandParams = params
		commands = append(commands, command)
	}
	job.Commands = commands

	// Extract any cleanup actions for the job.
	cleanup, err := cleanupFromJSON(jsonParsed)
	if err != nil {
		return job, err
	}
	job.CleanUp = cleanup
	return job, nil
}

func ParsePlainTextParams(filename string) (datamodels.Job, error) {
	var job = datamodels.Job{}
	var commands = make([]datamodels.Command, 0)
	var command = datamodels.Command{}
	var slurmPreamble = datamodels.SlurmPreamble{}
	var commandPreamble = datamodels.CommandPreamble{}
	var CommandParams = datamodels.CommandParams{}
	var lastTag string

	fmt.Printf("Parsing plain text parameter file... ")

	// Open the file for buffer based read.
	fileBuf, err := os.Open(filename)
	if err != nil {
		return job, err
	}

	// Defer file handle closing.
	defer func() {
		if err = fileBuf.Close(); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}()

	// Create a file scanner for reading the lines of the file.
	scanner := bufio.NewScanner(fileBuf)

	// Read the file line by line.
	for scanner.Scan() {
		line := scanner.Text()

		// Ignore any blank or commented lines.
		if IgnoreLine(line) {
			continue
		}

		// Get the line tag and assigned value.
		tag, val := ParseLine(line)

		if (tag == "JOB_NAME") && (lastTag == "ARGUMENT" || lastTag == "OPTION") {
			// Finalize the command and add command to array once we hit the next command block.
			command.Preamble = commandPreamble
			command.CommandParams = CommandParams
			commands = append(commands, command)

			// Reset the objects.
			commandPreamble = datamodels.CommandPreamble{}
			command = datamodels.Command{}
			CommandParams = datamodels.CommandParams{}
			lastTag = ""
		}

		if tag == "JOB_NAME" {
			// Set the command name.
			command.CommandParams.Command = val
		}

		// Set batch preamble
		if IsBatchPreamble(tag) {
			setBatchPreamble(tag, val, &command)
		}

		// Set slurm preamble
		if IsSlurmPreamble(tag) {
			setSlurmPreamble(tag, val, &slurmPreamble)
		}

		// Set command preamble
		if IsSlurmCommandPreamble(tag) {
			setCommandPreamble(tag, val, &commandPreamble)
		}

		// Get and set the parameters.
		setCommandParams(tag, val, &CommandParams)
		lastTag = tag
	}

	// If there was an error with the scan, panic!
	err = scanner.Err()
	if err != nil {
		return job, err
	}

	// Handle the last script def that was parsed before scanner ended.
	job.SlurmPreamble = slurmPreamble
	command.Preamble = commandPreamble
	command.CommandParams = CommandParams
	commands = append(commands, command)

	// Assign the commands to the job.
	job.Commands = commands
	fmt.Printf("Done.\n")
	return job, nil
}

/* -----------------------------------------------------------------------------
 * JSON helpers
 * -------------------------------------------------------------------------- */

func isBatchCommand(jsonParsed *gabs.Container) (bool, error) {
	var err error

	if jsonParsed.Exists("batch") {
		return jsonParsed.Path("batch").Data().(bool), nil
	}

	err = errors.New(`JSON error: Missing parameter "batch"`)
	return false, err
}

func commandNameFromJSON(jsonParsed *gabs.Container) (string, error) {
	var err error

	if jsonParsed.Exists("command") {
		return jsonParsed.Path("command").Data().(string), nil
	}

	err = errors.New(`JSON error: Missing parameter "command"`)
	return "", err
}

func inputFromStep(jsonParsed *gabs.Container) string {
	if jsonParsed.Exists("input_from_step") && jsonParsed.Path("input_from_step").Data() != nil {
		return jsonParsed.Path("input_from_step").Data().(string)
	}
	return ""
}

func jobDetailsFromJSON(jsonParsed *gabs.Container) (datamodels.JobDetails, error) {
	var details datamodels.JobDetails
	var err error

	if jsonParsed.Exists("job_details") {
		detailsJSON := jsonParsed.Path("job_details")
		if detailsJSON.Exists("job_name") {
			details.Name = detailsJSON.Path("job_name").Data().(string)
		} else {
			err = errors.New(`JSON error: Missing parameter "job_name"`)
			return details, err
		}
		if detailsJSON.Exists("design_file") {
			details.DesignFile = detailsJSON.Path("design_file").Data().(string)
		}
	}
	return details, nil
}

func slurmPreambleFromJSON(jsonParsed *gabs.Container) (datamodels.SlurmPreamble, error) {
	var err error
	var preamble = datamodels.SlurmPreamble{}

	if jsonParsed.Exists("wall_time") {
		preamble.WallTime = jsonParsed.Path("wall_time").Data().(string)
	} else {
		err = errors.New(`JSON error: Missing parameter "wall_time"`)
		return preamble, err
	}
	if jsonParsed.Exists("partition") {
		preamble.Partition = jsonParsed.Path("partition").Data().(string)
	} else {
		err = errors.New(`JSON error: Missing parameter "partition"`)
		return preamble, err
	}
	if jsonParsed.Exists("email_begin") {
		preamble.EmailBegin = jsonParsed.Path("email_begin").Data().(bool)
	} else {
		err = errors.New(`JSON error: Missing parameter "email_begin"`)
		return preamble, err
	}
	if jsonParsed.Exists("email_end") {
		preamble.EmailEnd = jsonParsed.Path("email_end").Data().(bool)
	} else {
		err = errors.New(`JSON error: Missing parameter "email_end"`)
		return preamble, err
	}
	if jsonParsed.Exists("email_fail") {
		preamble.EmailFail = jsonParsed.Path("email_fail").Data().(bool)
	} else {
		err = errors.New(`JSON error: Missing parameter "email_fail"`)
		return preamble, err
	}
	if jsonParsed.Exists("email_address") {
		preamble.EmailAddress = jsonParsed.Path("email_address").Data().(string)
	} else {
		err = errors.New(`JSON error: Missing parameter "email_address"`)
		return preamble, err
	}
	return preamble, nil
}

func sgePreambleFromJSON(jsonParsed *gabs.Container) (datamodels.SGEPreamble, error) {
	var err error
	var preamble = datamodels.SGEPreamble{}

	if jsonParsed.Exists("current_directory") {
		preamble.CWD = jsonParsed.Path("current_directory").Data().(bool)
	} else {
		preamble.CWD = false
	}
	if jsonParsed.Exists("join_output") {
		preamble.JoinOutput = jsonParsed.Path("join_output").Data().(bool)
	} else {
		preamble.JoinOutput = false
	}
	if jsonParsed.Exists("email_address") {
		preamble.EmailAddress = jsonParsed.Path("email_address").Data().(string)
	} else {
		err = errors.New(`JSON error: Missing parameter "email_address"`)
		return preamble, err
	}
	if jsonParsed.Exists("shell") {
		preamble.Shell = jsonParsed.Path("shell").Data().(string)
	} else {
		err = errors.New(`JSON error: Missing parameter "parallel_environment"`)
		return preamble, err
	}
	if jsonParsed.Exists("parallel_environment") {
		preamble.ParallelEnv = jsonParsed.Path("parallel_environment").Data().(string)
	} else {
		err = errors.New(`JSON error: Missing parameter "parallel_environment"`)
		return preamble, err
	}
	if jsonParsed.Exists("memory") {
		preamble.Memory = jsonParsed.Path("memory").Data().(string)
	} else {
		err = errors.New(`JSON error: Missing parameter "memory"`)
		return preamble, err
	}

	return preamble, nil
}

func experimentDetailsFromJSON(jsonParsed *gabs.Container) datamodels.Experiment {
	var experimentDetails = datamodels.DefaultExperiment()

	if jsonParsed.Exists("experiment_details") {
		experimentJSON := jsonParsed.Path("experiment_details")
		if experimentJSON.Exists("pi") {
			experimentDetails.PI = experimentJSON.Path("pi").Data().(string)
		}
		if experimentJSON.Exists("experiment_name") {
			experimentDetails.Name = experimentJSON.Path("experiment_name").Data().(string)
		}
		if experimentJSON.Exists("analysis_id") {
			experimentDetails.AnalysisID = experimentJSON.Path("analysis_id").Data().(string)
		}
		if experimentJSON.Exists("sample_path") && experimentJSON.Path("sample_path").Data() != nil {
			experimentDetails.SamplePath = experimentJSON.Path("sample_path").Data().(string)
		}
		if experimentJSON.Exists("analysis_path") && experimentJSON.Path("analysis_path").Data() != nil {
			experimentDetails.AnalysisPath = experimentJSON.Path("analysis_path").Data().(string)
		}
		if experimentJSON.Exists("workdir") && experimentJSON.Path("workdir").Data() != nil {
			experimentDetails.WorkDir = experimentJSON.Path("workdir").Data().(string)
		}
		if experimentJSON.Exists("samples_file") && experimentJSON.Path("samples_file").Data() != nil {
			experimentDetails.SamplesFile = experimentJSON.Path("samples_file").Data().(string)
		}
	}
	return experimentDetails

}

func miscPreambleFromJSON(jsonParsed *gabs.Container) (datamodels.MiscPreamble, error) {
	var miscPreamble datamodels.MiscPreamble
	var lines = make([]string, 0)

	children := jsonParsed.Path("misc_preamble").Children()
	for _, child := range children {
		lines = append(lines, child.Data().(string))
	}

	miscPreamble.Lines = lines
	return miscPreamble, nil
}

func commandPreambleFromJSON(jsonParsed *gabs.Container) (datamodels.CommandPreamble, error) {
	var err error
	var preamble datamodels.CommandPreamble

	if jsonParsed.Exists("tasks") {
		preamble.Tasks = int64(jsonParsed.Path("tasks").Data().(float64))
	} else {
		err = errors.New(`JSON error: Missing parameter "tasks"`)
		return preamble, err
	}
	if jsonParsed.Exists("cpus") {
		preamble.CPUs = int64(jsonParsed.Path("cpus").Data().(float64))
	} else {
		err = errors.New(`JSON error: Missing parameter "cpus"`)
		return preamble, err
	}
	if jsonParsed.Exists("memory") {
		preamble.Memory = int64(jsonParsed.Path("memory").Data().(float64))
	} else {
		err = errors.New(`JSON error: Missing parameter "memory"`)
		return preamble, err
	}

	return preamble, nil
}

func commandParamsFromJSON(jsonParsed *gabs.Container) (datamodels.CommandParams, error) {
	var err error
	var params datamodels.CommandParams

	if jsonParsed.Exists("command") {
		params.Command = jsonParsed.Path("command").Data().(string)
	} else {
		err = errors.New(`JSON error: Missing parameter "command"`)
		return params, err
	}
	if jsonParsed.Exists("volumes") {
		volumes := jsonParsed.Path("volumes")
		params.Volumes = volumesFromJSON(volumes)
	} else {
		err = errors.New(`JSON error: Missing parameter "volumes"`)
		return params, err
	}
	if jsonParsed.Exists("singularity_path") {
		params.SingularityPath = jsonParsed.Path("singularity_path").Data().(string)
	} else {
		err = errors.New(`JSON error: Missing parameter "singularity_path"`)
		return params, err
	}
	if jsonParsed.Exists("singularity_image") {
		params.SingularityImage = jsonParsed.Path("singularity_image").Data().(string)
	} else {
		err = errors.New(`JSON error: Missing parameter "singularity_image"`)
		return params, err
	}

	// Deal with the "optional params"
	if jsonParsed.Exists("workdir") {
		params.WorkDir = jsonParsed.Path("workdir").Data().(string)
	}
	if jsonParsed.Exists("subcommand") {
		params.Subcommand = jsonParsed.Path("subcommand").Data().(string)
	}

	// Get all command "options".
	params.CommandOptions = commandOptionsFromJSON(jsonParsed)
	// Get all command "arguments".
	params.CommandArgs = commandArgumentsFromJSON(jsonParsed)

	return params, nil
}

func commandOptionsFromJSON(jsonParsed *gabs.Container) []string {
	var options = make([]string, 0)
	for _, c := range jsonParsed.Path("options").Children() {
		options = append(options, c.Data().(string))
	}
	return options
}

func commandArgumentsFromJSON(jsonParsed *gabs.Container) []string {
	var arguments = make([]string, 0)
	for _, c := range jsonParsed.Path("arguments").Children() {
		arguments = append(arguments, c.Data().(string))
	}
	return arguments
}

func volumesFromJSON(jsonParsed *gabs.Container) []datamodels.VolumeMount {
	var volumes = make([]datamodels.VolumeMount, 0)
	children := jsonParsed.Children()

	for _, c := range children {
		v := datamodels.VolumeMount{
			HostPath:      c.Path("host_path").Data().(string),
			ContainerPath: c.Path("container_path").Data().(string),
		}
		volumes = append(volumes, v)
	}
	return volumes
}

// func cleanupFromJSON(jsonParsed *gabs.Container) ([]datamodels.CleanupAction, error) {
// 	var err error
// 	var cleanupActions = make([]datamodels.CleanupAction, 0)

// 	if jsonParsed.Exists("cleanup") {
// 		children := jsonParsed.Path("cleanup").Children()
// 		for _, c := range children {
// 			targets := c.Path("targets").Children()
// 			for _, t := range targets {
// 				cua := datamodels.CleanupAction{}
// 				cua.ToolName = c.Path("tool_name").Data().(string)
// 				cua.Action = c.Path("action").Data().(string)
// 				if cua.Action != "rm" && cua.Action != "mv" && cua.Action != "cp" {
// 					err = fmt.Errorf(`json template error: unsupported clean-up action "%s". Commander currently supports rm, mv and cp`, cua.Action)
// 					return cleanupActions, err
// 				}
// 				// Source is expected for all cleanup actions.
// 				cua.Source = t.Path("source").Data().(string)

// 				// Destination is required for mv and cp actions.
// 				if cua.Action == "mv" || cua.Action == "cp" {
// 					if !t.Exists("destination") {
// 						err = fmt.Errorf(`json template error: clean-up action "%s" is missing a destination argument`, cua.Action)
// 						return cleanupActions, err
// 					}
// 					cua.Destination = t.Path("destination").Data().(string)
// 				} else if t.Exists("destination") {
// 					cua.Destination = t.Path("destination").Data().(string)
// 				}
// 				cleanupActions = append(cleanupActions, cua)
// 			}
// 		}
// 	}
// 	return cleanupActions, nil
// }

func cleanupFromJSON(jsonParsed *gabs.Container) ([]string, error) {
	var cleanupActions = make([]string, 0)

	if jsonParsed.Exists("cleanup") {
		children := jsonParsed.Path("cleanup").Children()
		for _, c := range children {
			cleanupActions = append(cleanupActions, c.Data().(string))
		}
	}
	return cleanupActions, nil
}

/* -----------------------------------------------------------------------------
 * Plaintext helpers
 * -------------------------------------------------------------------------- */

func setSlurmPreamble(tag, val string, slurmPreamble *datamodels.SlurmPreamble) {
	if tag == "JOB_NAME" {
		slurmPreamble.JobName = val
	} else if tag == "PARTITION" {
		slurmPreamble.Partition = val
	} else if tag == "EMAIL_BEGIN" {
		slurmPreamble.EmailBegin, _ = strconv.ParseBool(val)
	} else if tag == "EMAIL_END" {
		slurmPreamble.EmailEnd, _ = strconv.ParseBool(val)
	} else if tag == "EMAIL_FAIL" {
		slurmPreamble.EmailFail, _ = strconv.ParseBool(val)
	} else if tag == "EMAIL_ADDRESS" {
		slurmPreamble.EmailAddress = val
	}
}

func setCommandPreamble(tag, val string, commandPreamble *datamodels.CommandPreamble) {
	if tag == "TASKS" {
		commandPreamble.Tasks, _ = strconv.ParseInt(val, 10, 64)
	} else if tag == "CPUS" {
		commandPreamble.CPUs, _ = strconv.ParseInt(val, 10, 64)
	} else if tag == "MEMORY" {
		commandPreamble.Memory, _ = strconv.ParseInt(val, 10, 64)
	}
}

func setCommandParams(tag, val string, params *datamodels.CommandParams) {
	if tag == "SINGULARITY_PATH" {
		params.SingularityPath = val
	} else if tag == "SINGULARITY_IMAGE" {
		params.SingularityImage = val
	} else if tag == "WORK_DIR" {
		params.WorkDir = val
	} else if tag == "COMMAND" {
		params.Command = val
	} else if tag == "SUBCOMMAND" {
		params.Subcommand = val
	} else if tag == "OPTION" {
		params.CommandOptions = append(params.CommandOptions, val)
	} else if tag == "ARGUMENT" {
		params.CommandArgs = append(params.CommandArgs, val)
	}
}

func setBatchPreamble(tag, val string, cmd *datamodels.Command) {
	if tag == "BATCH" {
		cmd.Batch = true
	} else if tag == "SAMPLES_FILE" {
		cmd.SamplesFile = val
	}
}
