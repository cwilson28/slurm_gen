package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"slurm_gen/datamodels"
	"slurm_gen/utils"
	"strconv"
	"strings"
)

// Show the help message for slurm_gen
func ShowHelp() {
	fmt.Println(datamodels.HELP_MSG)
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
		if utils.IgnoreLine(line) {
			continue
		}

		// Get the line tag and assigned value.
		tag, val := utils.ParseLine(line)

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

		// Set batch preamble
		if utils.IsBatchPreamble(tag) {
			setBatchPreamble(tag, val, &command)
		}

		// Set slurm preamble
		if utils.IsSlurmPreamble(tag) {
			setSlurmPreamble(tag, val, &slurmPreamble)
		}

		// Set command preamble
		if utils.IsCommandPreamble(tag) {
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

func setSlurmParams(tag, val string, SlurmParams *datamodels.SlurmParams, CommandParams *datamodels.CommandParams) {
	if tag == "JOB_NAME" {
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

func setBatchPreamble(tag, val string, cmd *datamodels.Command) {
	if tag == "BATCH" {
		cmd.Batch = true
	} else if tag == "SAMPLES_FILE" {
		cmd.SamplesFile = val
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

func main() {
	var submit bool

	// Declare command line flags.
	flag.Bool("help", false, "Show help message")
	flag.String("params", "", "Full path to job parameter file")
	flag.Bool("pipeline", false, "Generate an accompanying pipeline script")
	flag.Bool("submit", false, "Submit job on the user's behalf")
	flag.Parse()

	/* -------------------------------------------------------------------------
	 * Check for the help param and display help message if provided.
	 * ---------------------------------------------------------------------- */
	help := flag.Lookup("help")
	if help.Value.String() == "true" {
		ShowHelp()
		os.Exit(1)
	}

	/* -------------------------------------------------------------------------
	 * Check for the help param and display help message if provided.
	 * ---------------------------------------------------------------------- */
	submitFlag := flag.Lookup("submit")
	if submitFlag.Value.String() == "true" {
		submit = true
	}

	/* -------------------------------------------------------------------------
	 * Check for the param file flag. If it's not provided, exit with message.
	 * ---------------------------------------------------------------------- */
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
		utils.WriteSlurmPreamble(parentSlurmFile, job.Commands[0].CommandName, job.SlurmPreamble)

		// Write intermediary job shit.
		for _, line := range datamodels.MORE_JOBSHIT {
			fmt.Fprintln(parentSlurmFile, line)
		}

		// Write the command files
		for _, c := range job.Commands {
			bashScript, err := utils.WriteBashScript(parentSlurmFile, c)
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
		os.Exit(0)
	}
	// Generate a slurm script for a single tool.
	fmt.Println("Generating slurm script...")

	cmd := job.Commands[0]

	// This is a batch command.
	// Write multiple slurm scripts.
	if cmd.Batch {
		fmt.Println("Writing batch files...")

		batchFiles := make([]string, 0)

		// Get the samples over which this command will be run
		samples := utils.ParseSamplesFile(cmd.SamplesFile)

		for _, s := range samples {
			// We are going to write a slurm script for each
			filename := fmt.Sprintf("%s_%s.slurm", cmd.CommandName, s.Prefix)
			outfile, err := os.Create(filename)
			if err != nil {
				panic(err)
			}

			defer func() {
				if err = outfile.Close(); err != nil {
					log.Fatal(err)
				}
			}()

			utils.WriteSlurmPreamble(outfile, strings.TrimRight(filename, ".slurm"), job.SlurmPreamble)
			utils.WriteCommandPreamble(outfile, cmd.Preamble)
			// Write intermediary job shit.
			// TODO: Wrap this in a "Write" command in the output_utils file.
			for _, line := range datamodels.MORE_JOBSHIT {
				fmt.Fprintln(outfile, line)
			}

			// TODO: Wrap this in a "Write" command in the output_utils file.
			fmt.Fprintln(outfile, fmt.Sprintf("%s", datamodels.JOB_SHIT["singularity_cmd"]))
			fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["singularity_bind"], cmd.CommandParams.Volume)))
			fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["singularity_env"], cmd.CommandParams.SingularityPath, cmd.CommandParams.SingularityImage)))
			fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["command"], cmd.CommandParams.Command)))

			// TODO: This needs to be wrapped in star options command or something. We will likely
			// have other tools that require specific option formattin.
			if cmd.CommandName == "star" {
				for _, opt := range cmd.CommandParams.CommandOptions {
					chunks := strings.Split(opt, " ")
					if chunks[0] == "--outFileNamePrefix" {
						opt = fmt.Sprintf("%s %s", chunks[0], fmt.Sprintf("%s/%s_", s.Path, s.Prefix))
					}
					if chunks[0] == "--readFilesIn" {
						opt = fmt.Sprintf("%s %s", chunks[0], s.DumpReadFiles())
					}
					utils.WriteCommandOption(outfile, opt)
				}
			} else {
				utils.WriteCommandOptions(outfile, cmd.CommandParams.CommandOptions)
			}

			// Write any command args that are provided.
			utils.WriteCommandArgs(outfile, cmd.CommandParams.CommandArgs)
			fmt.Printf("%s written successfully\n", filename)
			batchFiles = append(batchFiles, filename)
		}
		if submit {
			// Submit the jobs
			for _, f := range batchFiles {
				_, err := exec.Command("sbatch", f).Output()
				if err != nil {
					fmt.Printf("%s", err)
				}
			}
		}
		os.Exit(0)
	}

	// Name the slurm file after the command.
	filename := fmt.Sprintf("%s.slurm", cmd.CommandName)
	outfile, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err = outfile.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	// Write the slurm script
	utils.WriteSlurmScript(outfile, job)
	fmt.Printf("Slurm script %s written.", filename)
	os.Exit(0)

	// // Check for submit flag
	// submitFlag := flag.Lookup("submit")
	// if submitFlag == nil {
	// 	os.Exit(0)
	// }

	// Submit the job.
}
