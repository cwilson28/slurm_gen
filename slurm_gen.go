package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"slurm_gen/datamodels"
	"strconv"
	"strings"
)

// Show the help message for slurm_gen
func ShowHelp() {
	fmt.Println(datamodels.HELP_MSG)
}

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

	// Read the file line by line.
	for scanner.Scan() {
		line := scanner.Text()

		sample := datamodels.Sample{}
		// We are assuming the files are separated with a space.
		fileNames := strings.Split(line, " ")

		// Get the file prefix from the forward read
		sample.Prefix = parseSamplePrefix(fileNames[0])

		// Append the forward reads.
		sample.ForwardReadFile = fileNames[0]

		// Append the reverse reads if provided.
		if len(fileNames) == 2 {
			sample.ReverseReadFile = fileNames[1]
		}

		samples = append(samples, sample)
	}
	return samples
}

func parseSamplePrefix(sampleFileName string) string {
	chunks := strings.Split(sampleFileName, "/")
	filename := chunks[len(chunks)-1]
	prefix := strings.Split(strings.TrimRight(filename, ".fastq.gz"), "_R1")[0]
	return prefix
}

func parseBatchOpt(batchCommands string) (bool, []string) {
	var commands = make([]string, 0)
	if batchCommands == "" {
		return false, commands
	}

	commands = strings.Split(batchCommands, ",")
	return true, commands
}

func isBatchCmd(cmd string, cmds []string) bool {
	for _, c := range cmds {
		if c == cmd {
			return true
		}
	}
	return false
}

func isJobPreamble(tag string) bool {
	if tag == "BATCH" ||
		tag == "SAMPLES_FILE" {
		return true
	}
	return false
}

func isSlurmPreamble(tag string) bool {
	if tag == "PARTITION" ||
		tag == "NOTIFICATION_BEGIN" ||
		tag == "NOTIFICATION_END" ||
		tag == "NOTIFICATION_FAIL" ||
		tag == "NOTIFICATION_EMAIL" {
		return true
	}
	return false
}

func isCommandPreamble(tag string) bool {
	if tag == "JOB_NAME" ||
		tag == "TASKS" ||
		tag == "CPUS" ||
		tag == "MEMORY" ||
		tag == "TIME" {
		return true
	}
	return false
}

func JobGen(scanner *bufio.Scanner) (datamodels.Job, error) {
	var job = datamodels.Job{}
	var commands = make([]datamodels.Command, 0)
	var command = datamodels.Command{}
	var slurmPreamble = datamodels.SlurmPreamble{}
	var commandPreamble = datamodels.CommandPreamble{}
	var CommandParams = datamodels.CommandParams{}
	var lastTag string

	// Read the file line by line.
	for scanner.Scan() {
		line := scanner.Text()

		// Ignore any blank or commented lines.
		if ignoreLine(line) {
			continue
		}

		// Get the line tag and assigned value.
		tag, val := parseLine(line)

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
			command.CommandName = val
		}

		if isJobPreamble(tag) {
			setJobPreamble(tag, val, &job)
		}

		// Set slurm preamble
		if isSlurmPreamble(tag) {
			setSlurmPreamble(tag, val, &slurmPreamble)
		}

		// Set command preamble
		if isCommandPreamble(tag) {
			setCommandPreamble(tag, val, &commandPreamble)
		}

		// Get and set the parameters.
		setCommandParams(tag, val, &CommandParams)
		lastTag = tag
	}

	// If there was an error with the scan, panic!
	err := scanner.Err()
	if err != nil {
		panic(err)
	}

	// Handle the last script def that was parsed before scanner ended.
	job.SlurmPreamble = slurmPreamble
	command.Preamble = commandPreamble
	command.CommandParams = CommandParams
	commands = append(commands, command)

	// Assign the commands to the job.
	job.Commands = commands
	return job, nil
}

func ignoreLine(l string) bool {
	if l == "" || strings.Contains(l, "#") {
		return true
	}
	return false
}

func parseLine(l string) (string, string) {
	lineChunks := strings.Split(l, "=")
	return lineChunks[0], cleanLine(lineChunks[1])
}

func setSlurmParams(tag, val string, SlurmParams *datamodels.SlurmParams, CommandParams *datamodels.CommandParams) {
	if tag == "BATCH" {
		SlurmParams.Batch, _ = strconv.ParseBool(val)
	} else if tag == "SAMPLES_FILE" {
		SlurmParams.SamplesFile = val
	} else if tag == "SAMPLE_FILE_PREFIX" {
		SlurmParams.SampleFilePrefix = val
	} else if tag == "JOB_NAME" {
		SlurmParams.JobName = val
	} else if tag == "PARTITION" {
		SlurmParams.Partition = val
	} else if tag == "NOTIFICATION_BEGIN" {
		SlurmParams.NotificationBegin, _ = strconv.ParseBool(val)
	} else if tag == "NOTIFICATION_END" {
		SlurmParams.NotificationEnd, _ = strconv.ParseBool(val)
	} else if tag == "NOTIFICATION_FAIL" {
		SlurmParams.NotificationFail, _ = strconv.ParseBool(val)
	} else if tag == "NOTIFICATION_EMAIL" {
		SlurmParams.NotificationEmail = val
	} else if tag == "TASKS" {
		SlurmParams.Tasks, _ = strconv.ParseInt(val, 10, 64)
	} else if tag == "CPUS" {
		SlurmParams.CPUs, _ = strconv.ParseInt(val, 10, 64)
	} else if tag == "MEMORY" {
		SlurmParams.Memory, _ = strconv.ParseInt(val, 10, 64)
	} else if tag == "SINGULARITY_PATH" {
		CommandParams.SingularityPath = val
	} else if tag == "SINGULARITY_IMAGE" {
		CommandParams.SingularityImage = val
	} else if tag == "WORK_DIR" {
		CommandParams.WorkDir = val
	} else if tag == "VOLUME" {
		CommandParams.Volume = val
	} else if tag == "COMMAND" {
		CommandParams.Command = val
	} else if tag == "OPTION" {
		CommandParams.CommandOptions = append(CommandParams.CommandOptions, val)
	} else if tag == "ARGUMENT" {
		CommandParams.CommandArgs = append(CommandParams.CommandArgs, val)
	}
}

func setCommandParams(tag, val string, params *datamodels.CommandParams) {
	if tag == "SINGULARITY_PATH" {
		params.SingularityPath = val
	} else if tag == "SINGULARITY_IMAGE" {
		params.SingularityImage = val
	} else if tag == "WORK_DIR" {
		params.WorkDir = val
	} else if tag == "VOLUME" {
		params.Volume = val
	} else if tag == "COMMAND" {
		params.Command = val
	} else if tag == "OPTION" {
		params.CommandOptions = append(params.CommandOptions, val)
	} else if tag == "ARGUMENT" {
		params.CommandArgs = append(params.CommandArgs, val)
	}
}

func setJobPreamble(tag, val string, job *datamodels.Job) {
	if tag == "BATCH" {
		isBatch, batchCommands := parseBatchOpt(val)
		job.Batch = isBatch
		job.BatchCommands = batchCommands
	} else if tag == "SAMPLES_FILE" {
		job.SamplesFile = val
	}
}

func setSlurmPreamble(tag, val string, preamble *datamodels.SlurmPreamble) {
	if tag == "PARTITION" {
		preamble.Partition = val
	} else if tag == "NOTIFICATION_BEGIN" {
		preamble.NotificationBegin, _ = strconv.ParseBool(val)
	} else if tag == "NOTIFICATION_END" {
		preamble.NotificationEnd, _ = strconv.ParseBool(val)
	} else if tag == "NOTIFICATION_FAIL" {
		preamble.NotificationFail, _ = strconv.ParseBool(val)
	} else if tag == "NOTIFICATION_EMAIL" {
		preamble.NotificationEmail = val
	}
}

func setCommandPreamble(tag, val string, preamble *datamodels.CommandPreamble) {
	if tag == "JOB_NAME" {
		preamble.JobName = val
	} else if tag == "TASKS" {
		preamble.Tasks, _ = strconv.ParseInt(val, 10, 64)
	} else if tag == "CPUS" {
		preamble.CPUs, _ = strconv.ParseInt(val, 10, 64)
	} else if tag == "MEMORY" {
		preamble.Memory, _ = strconv.ParseInt(val, 10, 64)
	}
}

func cleanLine(l string) string {
	chunks := strings.Split(l, ";;")
	// We want everything up to the comment delimiter
	return strings.TrimRight(chunks[0], " ")
}

func writeCommandOptions(outfile *os.File, options []string) {
	// Write all command options.
	for _, opt := range options {
		fmt.Fprintln(outfile, fmt.Sprintf("%s \\", opt))
	}
}

func writeCommandArgs(outfile *os.File, args []string) {
	// Write all command options.
	for _, opt := range args {
		fmt.Fprintln(outfile, fmt.Sprintf("%s \\", opt))
	}
}

func writeBashScript(outfile *os.File, command datamodels.Command) (string, error) {
	filename := fmt.Sprintf("%s.sh", command.CommandName)
	outfile, err := os.Create(filename)
	if err != nil {
		return filename, err
	}

	// *** The following can be written independent of the actual command. *** //
	// Write the script header.
	fmt.Fprintln(outfile, fmt.Sprintf("#!/bin/bash\n"))
	// Write job shit.
	fmt.Fprintln(outfile, fmt.Sprintf("%s", datamodels.JOB_SHIT["singularity_cmd"]))
	fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["singularity_bind"], command.CommandParams.Volume)))
	fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["singularity_env"], command.CommandParams.SingularityPath, command.CommandParams.SingularityImage)))
	fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["command"], command.CommandParams.Command)))

	writeCommandOptions(outfile, command.CommandParams.CommandOptions)
	writeCommandArgs(outfile, command.CommandParams.CommandArgs)
	outfile.Close()
	fmt.Printf("%s.sh written successfully\n", command.CommandName)
	return filename, nil
}

func writeSlurmPreamble(slurmFile *os.File, jobname string, preamble datamodels.SlurmPreamble) {
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", datamodels.SLURM_PREAMBLE["header"]))
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["job_name"], jobname)))
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["partition"], preamble.Partition)))
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["notifications"], preamble.NotificationType())))
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["email"], preamble.NotificationEmail)))
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["job_log"], jobname)))
	fmt.Fprintln(slurmFile)
}

func writeCommandPreamble(slurmFile *os.File, preamble datamodels.CommandPreamble) {
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["tasks"], preamble.Tasks)))
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["cpus"], preamble.CPUs)))
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["memory"], preamble.Memory)))
	fmt.Fprintln(slurmFile)
}

func writeSlurmScript(slurmFile *os.File, job datamodels.Job) {
	// Write the slurm preamble for the parent slurm script
	writeSlurmPreamble(slurmFile, job.Commands[0].CommandName, job.SlurmPreamble)
	// Write the command preamble. We can use Commands[0] since there *should* only be one command.
	writeCommandPreamble(slurmFile, job.Commands[0].Preamble)

	// Write intermediary job shit.
	for _, line := range datamodels.MORE_JOBSHIT {
		fmt.Fprintln(slurmFile, line)
	}

	// Write command shit.
	// TODO: Wrap this in a function.
	command := job.Commands[0]
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", datamodels.JOB_SHIT["singularity_cmd"]))
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["singularity_bind"], command.CommandParams.Volume)))
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["singularity_env"], command.CommandParams.SingularityPath, command.CommandParams.SingularityImage)))
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["command"], command.CommandParams.Command)))

	writeCommandOptions(slurmFile, command.CommandParams.CommandOptions)
	writeCommandArgs(slurmFile, command.CommandParams.CommandArgs)
}

func main() {

	flag.Bool("help", false, "Show help message")
	flag.String("params", "", "Full path to job parameter file")
	flag.Bool("pipeline", false, "Generate an accompanying pipeline script")
	flag.Bool("submit", false, "Submit job on the user's behalf")
	flag.Parse()

	// Check for the help param
	help := flag.Lookup("help")
	if help.Value.String() == "true" {
		ShowHelp()
		os.Exit(1)
	}

	// Check for the param file flag
	paramFile := flag.Lookup("params")
	if paramFile == nil {
		fmt.Println("Error: Missing parameter file. Please include --params <path to your parameter file>")
		os.Exit(1)
	}

	// Open the file for buffer based read.
	fileBuf, err := os.Open(paramFile.Value.String())
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

	// Parse all script definitions supplied in the parameter file
	job, err := JobGen(scanner)
	if err != nil {
		panic(err)
	}

	pipelineFlag := flag.Lookup("pipeline")
	if pipelineFlag.Value.String() == "true" {
		// Generate a slurm pipeline script composed of the scripts we just generated.
		// The script will use srun and each tool command will be written as a bash script.
		fmt.Println("Generating pipeline slurm script...")

		// Open the parent slurm file
		filename := fmt.Sprintf("pipeline.slurm")
		parentSlurmFile, err := os.Create(filename)
		if err != nil {
			panic(err)
		}

		defer func() {
			if err = parentSlurmFile.Close(); err != nil {
				log.Fatal(err)
			}
		}()

		// Write the slurm preamble for the parent slurm script
		writeSlurmPreamble(parentSlurmFile, job.Commands[0].CommandName, job.SlurmPreamble)

		// Write intermediary job shit.
		for _, line := range datamodels.MORE_JOBSHIT {
			fmt.Fprintln(parentSlurmFile, line)
		}

		// Write the command files
		for _, c := range job.Commands {
			if isBatchCmd(c.CommandName, job.BatchCommands) {

			} else {
				bashScript, err := writeBashScript(parentSlurmFile, c)
				if err != nil {
					panic(err)
				}
				// Write the script line for the tool.
				fmt.Fprintln(
					parentSlurmFile,
					fmt.Sprintf(
						"srun --input=none -N%d -c%d --tasks-per-node=%d -w %s --mem-per-cpu=%d ./%s",
						c.Preamble.Tasks,
						c.Preamble.CPUs,
						c.Preamble.Tasks,
						job.SlurmPreamble.Partition,
						c.Preamble.Memory,
						bashScript,
					),
				)
			}
		}
	} else {
		// Generate a slurm script for a single tool.
		fmt.Println("Generating slurm script...")

		// Open the slurm file. We will name it after the command name.
		command := job.Commands[0]
		filename := fmt.Sprintf(fmt.Sprintf("%s.slurm", command.CommandName))
		slurmFile, err := os.Create(filename)
		if err != nil {
			panic(err)
		}

		defer func() {
			if err = slurmFile.Close(); err != nil {
				log.Fatal(err)
			}
		}()

		// Write the slurm script
		writeSlurmScript(slurmFile, job)
	}

	// // Check for submit flag
	// submitFlag := flag.Lookup("submit")
	// if submitFlag == nil {
	// 	os.Exit(0)
	// }

	// Submit the job.
}
