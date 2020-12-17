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

func ScriptGen(scanner *bufio.Scanner) ([]datamodels.ScriptDef, error) {
	var scriptDefs = make([]datamodels.ScriptDef, 0)
	var slurmParams = datamodels.SlurmParams{}
	var toolParams = datamodels.ToolParams{}
	var lastTag string

	// Read the file line by line.
	for scanner.Scan() {
		line := scanner.Text()
		if ignoreLine(line) && lastTag == "ARGUMENT" {
			def := datamodels.ScriptDef{
				SlurmParams: slurmParams,
				ToolParams:  toolParams,
			}
			scriptDefs = append(scriptDefs, def)
			slurmParams = datamodels.SlurmParams{}
			toolParams = datamodels.ToolParams{}
			lastTag = ""
			continue
		} else if ignoreLine(line) {
			continue
		}

		tag, val := parseLine(line)
		setSlurmParams(tag, val, &slurmParams, &toolParams)
		lastTag = tag
	}

	// If there was an error with the scan, panic!
	err := scanner.Err()
	if err != nil {
		return scriptDefs, err
	}

	// Handle the last script def that was parsed before scanner ended.
	def := datamodels.ScriptDef{
		SlurmParams: slurmParams,
		ToolParams:  toolParams,
	}

	scriptDefs = append(scriptDefs, def)
	return scriptDefs, nil
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

func setSlurmParams(tag, val string, SlurmParams *datamodels.SlurmParams, ToolParams *datamodels.ToolParams) {
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
		ToolParams.SingularityPath = val
	} else if tag == "SINGULARITY_IMAGE" {
		ToolParams.SingularityImage = val
	} else if tag == "WORK_DIR" {
		ToolParams.WorkDir = val
	} else if tag == "VOLUME" {
		ToolParams.Volume = val
	} else if tag == "COMMAND" {
		ToolParams.Command = val
	} else if tag == "OPTION" {
		ToolParams.CommandOptions = append(ToolParams.CommandOptions, val)
	} else if tag == "ARGUMENT" {
		ToolParams.CommandArgs = append(ToolParams.CommandArgs, val)
	}
}

func cleanLine(l string) string {
	chunks := strings.Split(l, ";;")
	// We want everything up to the comment delimiter
	return strings.TrimRight(chunks[0], " ")
}

func writeSlurmScript(def datamodels.ScriptDef) (string, error) {
	filename := fmt.Sprintf("%s.slurm", def.SlurmParams.JobName)
	outfile, err := os.Create(filename)
	if err != nil {
		return "", err
	}

	// Write the slurm script header.
	fmt.Fprintln(outfile, fmt.Sprintf("%s", datamodels.SLURM_PREAMBLE["header"]))

	// Write the rest of the preamble.
	fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["job_name"], def.SlurmParams.JobName)))
	fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["partition"], def.SlurmParams.Partition)))
	fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["notifications"], def.SlurmParams.NotificationType())))
	fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["email"], def.SlurmParams.NotificationEmail)))
	fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["tasks"], def.SlurmParams.Tasks)))
	fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["cpus"], def.SlurmParams.CPUs)))
	fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["memory"], def.SlurmParams.Memory)))
	fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["job_log"], def.SlurmParams.JobName)))

	// Write intermediary job shit.
	for _, line := range datamodels.MORE_JOBSHIT {
		fmt.Fprintln(outfile, line)
	}

	// Write job shit.
	fmt.Fprintln(outfile, fmt.Sprintf("%s", datamodels.JOB_SHIT["singularity_cmd"]))
	fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["singularity_bind"], def.ToolParams.Volume)))
	fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["singularity_env"], def.ToolParams.SingularityPath, def.ToolParams.SingularityImage)))
	fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["command"], def.ToolParams.Command)))

	// Write all command options.
	for _, opt := range def.ToolParams.CommandOptions {
		fmt.Fprintln(outfile, fmt.Sprintf("%s \\", opt))
	}

	// Write all command args.
	for _, arg := range def.ToolParams.CommandArgs {
		fmt.Fprintln(outfile, fmt.Sprintf("%s \\", arg))
	}

	err = outfile.Close()
	if err != nil {
		return "", err
	}

	fmt.Printf("%s.slurm written successfully\n", def.SlurmParams.JobName)
	return filename, nil
}

func writeBashScript(def datamodels.ScriptDef) (string, error) {
	filename := fmt.Sprintf("%s.sh", def.SlurmParams.JobName)
	outfile, err := os.Create(filename)
	if err != nil {
		return "", err
	}

	// Write the script header.
	fmt.Fprintln(outfile, fmt.Sprintf("#!/bin/bash\n"))

	// Write job shit.
	fmt.Fprintln(outfile, fmt.Sprintf("%s", datamodels.JOB_SHIT["singularity_cmd"]))
	fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["singularity_bind"], def.ToolParams.Volume)))
	fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["singularity_env"], def.ToolParams.SingularityPath, def.ToolParams.SingularityImage)))
	fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["command"], def.ToolParams.Command)))

	// Write all command options.
	for _, opt := range def.ToolParams.CommandOptions {
		fmt.Fprintln(outfile, fmt.Sprintf("%s \\", opt))
	}

	// Write all command args.
	for _, arg := range def.ToolParams.CommandArgs {
		fmt.Fprintln(outfile, fmt.Sprintf("%s \\", arg))
	}

	err = outfile.Close()
	if err != nil {
		return "", err
	}

	fmt.Printf("%s.sh written successfully\n", def.SlurmParams.JobName)
	return filename, nil
}

func main() {
	flag.String("params", "", "Full path to job parameter file")
	flag.Bool("pipeline", false, "Generate an accompanying pipeline script")
	flag.Bool("submit", false, "Submit job on the user's behalf")
	flag.Parse()

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
	scriptDefs, err := ScriptGen(scanner)
	if err != nil {
		panic(err)
	}

	// Check for pipeline parameter
	pipelineFlag := flag.Lookup("pipeline")
	if pipelineFlag != nil {
		// Generate a slurm pipeline script composed of the scripts we just generated.
		// The script will use srun and each tool command will be written as a bash script.
		fmt.Println("Generating pipeline slurm script...")

		// Open the parent slurm file
		filename := fmt.Sprintf("parent.slurm")
		parentSlurmFile, err := os.Create(filename)
		if err != nil {
			panic(err)
		}
		// Write individual bash scripts for each tool in the pipeline.
		for i, def := range scriptDefs {
			scriptName, err := writeBashScript(def)
			if err != nil {
				panic(err)
			}

			// Write the slurm preamble to the parent script.
			if i == 0 {
				fmt.Fprintln(parentSlurmFile, fmt.Sprintf("%s", datamodels.SLURM_PREAMBLE["header"]))
				fmt.Fprintln(parentSlurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["job_name"], "pipeline")))
				fmt.Fprintln(parentSlurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["partition"], def.SlurmParams.Partition)))
				fmt.Fprintln(parentSlurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["notifications"], def.SlurmParams.NotificationType())))
				fmt.Fprintln(parentSlurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["email"], def.SlurmParams.NotificationEmail)))
				fmt.Fprintln(parentSlurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["job_log"], "pipeline")))
				fmt.Fprintln(parentSlurmFile)
				// Write intermediary job shit.
				for _, line := range datamodels.MORE_JOBSHIT {
					fmt.Fprintln(parentSlurmFile, line)
				}
			}

			// Write the script line for the tool.
			fmt.Fprintln(
				parentSlurmFile,
				fmt.Sprintf(
					"srun --input=none -N%d -c%d --tasks-per-node=%d -w %s --mem-per-cpu=%d ./%s",
					def.SlurmParams.Tasks,
					def.SlurmParams.CPUs,
					def.SlurmParams.Tasks,
					def.SlurmParams.Partition,
					def.SlurmParams.Memory,
					scriptName,
				),
			)
		}
		os.Exit(0)
	}

	var scriptFiles = make([]string, 0)
	for _, def := range scriptDefs {
		filename, err := writeSlurmScript(def)
		if err != nil {
			panic(err)
		}
		scriptFiles = append(scriptFiles, filename)
	}

	// Check for submit flag
	submitFlag := flag.Lookup("submit")
	if submitFlag == nil {
		os.Exit(0)
	}

	// Submit the job.
}
