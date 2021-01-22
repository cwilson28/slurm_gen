package utils

import (
	"commander/datamodels"
	"fmt"
	"os"
	"strings"
)

/* -----------------------------------------------------------------------------
 * Functions for writing output files for the slurm job.
 * -------------------------------------------------------------------------- */
func WriteSlurmJobScript(job datamodels.Job) error {
	var err error

	fmt.Println("Writing slurm script preamble...")

	// Open the parent slurm file
	filename := fmt.Sprintf("%s.slurm", job.SlurmPreamble.JobName)
	slurmFile, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer func() {
		if err = slurmFile.Close(); err != nil {
			panic(err)
		}
	}()

	// Write the slurm preamble for the parent slurm script
	writeSlurmJobPreamble(slurmFile, job.SlurmPreamble)

	// Write the max CPUs for this job. For pipeline jobs, this will be the
	// max CPUs requested by any single step in the pipeline. For single command
	// jobs, this will be the CPUs required for that command.
	writeJobCPU(slurmFile, job)

	// Write intermediary job shit.
	writeIntermediateJobShit(slurmFile)

	/* -----------------
	 * All slurm preamble is written at this point. All that remains is to write
	 * command specific jargon.
	 * -------------- */

	// If there are multiple commands, it is safe to assume we are generating
	// a slurm script for a pipeline. Write the command details in a pipeline
	// format.
	if len(job.Commands) > 1 {
		fmt.Println("Writing pipeline slurm script...")
		err = writePipelineSlurmScript(slurmFile, job)
		return err
	}

	// There is a single command, we will either write this as as single .slurm
	// file or as a batch slurm file depending on the command definition.
	// Since there is only a single command, grab the 0th command object.
	cmd := job.Commands[0]

	// If this is a batch command, write the required batch scripts.
	// TODO: Make this batch bash script writing into a function. It's being used
	// more than once.
	if cmd.Batch {
		fmt.Println("Writing command script...")
		err = writeBatchCommand(slurmFile, cmd, job)
		return err
	}

	// TODO: Revisit this.
	// The command is not a batch command, write the command to the slurm file
	// we opened earlier.
	WriteCommandPreamble(slurmFile, cmd.Preamble)
	WriteCommandOptions(slurmFile, cmd.CommandParams.CommandOptions)
	WriteCommandArgs(slurmFile, cmd.CommandParams.CommandArgs)
	return nil
}

/* -----------------------------------------------------------------------------
 * Helpers for writing different tupes of scripts and script components.
 * -------------------------------------------------------------------------- */

/* ---
 * Write intermidate job shit to the slurm file.slurmFile :-)
 *  --- */
func writeIntermediateJobShit(slurmFile *os.File) {
	for _, line := range datamodels.MORE_JOBSHIT {
		fmt.Fprintln(slurmFile, line)
	}
}

/* ---
 * Write the remaining contents of a pipeline slurm script. This function will
 * also generate the individual bash scripts for the commands being executed.
 * --- */
func writePipelineSlurmScript(slurmFile *os.File, job datamodels.Job) error {

	// Write the bash scripts for each command.
	for _, cmd := range job.Commands {
		if cmd.Batch {
			// User has indicated the command will be run in a batch format.
			fmt.Println("Writing command bash scripts...")
			// Get the samples over which this command will be run
			samples := ParseSamplesFile(job.SamplesFile)
			for _, sample := range samples {

				bashScriptName, err := writeCommandScriptForSample(cmd, sample)

				// Write the script line for the tool.
				fmt.Fprintln(
					slurmFile,
					fmt.Sprintf(
						"srun --input=none -K1 -N%d -c%d --tasks-per-node=%d -w %s --mem-per-cpu=%d ./%s&",
						cmd.Preamble.Tasks,
						cmd.Preamble.CPUs,
						cmd.Preamble.Tasks,
						job.SlurmPreamble.Partition,
						cmd.Preamble.Memory,
						bashScriptName,
					),
				)
				// Make the bash script executable.
				if err = os.Chmod(bashScriptName, 0755); err != nil {
					return err
				}
			}
		} else {
			// Write the bash script for the command.
			bashScript, err := WriteCommandScript(cmd)
			if err != nil {
				return err
			}

			// Write the line for the command in the slurm file.
			fmt.Fprintln(
				slurmFile,
				fmt.Sprintf(
					"srun --input=none -K1 -N%d -c%d --tasks-per-node=%d -w %s --mem-per-cpu=%d ./%s",
					cmd.Preamble.Tasks,
					cmd.Preamble.CPUs,
					cmd.Preamble.Tasks,
					job.SlurmPreamble.Partition,
					cmd.Preamble.Memory,
					bashScript,
				),
			)
			// Make the bash script executable.
			if err = os.Chmod(bashScript, 0755); err != nil {
				return err
			}
		}
		WriteWait(slurmFile)
	}
	return nil
}

/* ---
 * Finish writing slurm file given a single batch command.
 * --- */
func writeBatchCommand(slurmFile *os.File, cmd datamodels.Command, job datamodels.Job) error {

	// Parse the provided samples file
	samples := ParseSamplesFile(job.SamplesFile)
	for _, sample := range samples {

		// Write the command details to a bash script.
		bashScriptName, err := writeCommandScriptForSample(cmd, sample)

		// Make the bash script executable.
		if err = os.Chmod(bashScriptName, 0755); err != nil {
			return err
		}

		// Write the script line for the tool in the slurm file.
		fmt.Fprintln(
			slurmFile,
			fmt.Sprintf(
				"srun --input=none -K1 -N%d -c%d --tasks-per-node=%d -w %s --mem-per-cpu=%d ./%s&",
				cmd.Preamble.Tasks,
				cmd.Preamble.CPUs,
				cmd.Preamble.Tasks,
				job.SlurmPreamble.Partition,
				cmd.Preamble.Memory,
				bashScriptName,
			),
		)
	}
	// Write a wait block to the slurm file. Don't want the parent script to
	// exit before the children.
	WriteWait(slurmFile)
	return nil
}

/* ---
 * Write a parent slurm file.
 * --- */
// func WriteSlurmScript(slurmFile *os.File, job datamodels.Job) {
// 	// Write the slurm preamble for the parent slurm script
// 	writeSlurmJobPreamble(slurmFile, job.SlurmPreamble)
// 	// Write the command preamble. We can use Commands[0] since there *should* only be one command.
// 	WriteCommandPreamble(slurmFile, job.Commands[0].Preamble)

// 	// Write intermediary job shit.
// 	for _, line := range datamodels.MORE_JOBSHIT {
// 		fmt.Fprintln(slurmFile, line)
// 	}

// 	// Write command shit.
// 	// TODO: Wrap this in a function.
// 	command := job.Commands[0]
// 	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", datamodels.JOB_SHIT["singularity_cmd"]))
// 	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["singularity_bind"], command.CommandParams.Volume)))
// 	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["singularity_env"], command.CommandParams.SingularityPath, command.CommandParams.SingularityImage)))
// 	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["command"], command.CommandParams.Command)))

// 	WriteCommandOptions(slurmFile, command.CommandParams.CommandOptions)
// 	WriteCommandArgs(slurmFile, command.CommandParams.CommandArgs)
// }

/* ---
 * Write the command to a bash file.
 * --- */
func WriteCommandScript(cmd datamodels.Command) (string, error) {
	// Write a bash script for each sample.
	scriptName := fmt.Sprintf("%s.sh", cmd.CommandParams.Command)
	outfile, err := os.Create(scriptName)
	if err != nil {
		return scriptName, err
	}

	// Defer the file closing until the function returns.
	defer outfile.Close()

	// *** The following can be written independent of the actual command. *** //
	// Write the script header.
	fmt.Fprintln(outfile, fmt.Sprintf("#!/bin/bash\n"))

	// Write job shit.
	fmt.Fprintln(outfile, fmt.Sprintf("%s", datamodels.JOB_SHIT["singularity_cmd"]))
	fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["singularity_bind"], cmd.CommandParams.Volume)))
	fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["singularity_env"], cmd.CommandParams.SingularityPath, cmd.CommandParams.SingularityImage)))
	fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["command"], cmd.CommandParams.Command)))

	// Write the command options and command arguments
	WriteCommandOptions(outfile, cmd.CommandParams.CommandOptions)
	WriteCommandArgs(outfile, cmd.CommandParams.CommandArgs)

	return scriptName, nil
}

func writeCommandScriptForSample(command datamodels.Command, sample datamodels.Sample) (string, error) {
	outfileName := fmt.Sprintf("%s_%s.sh", command.CommandParams.Command, sample.Prefix)
	outfile, err := os.Create(outfileName)
	if err != nil {
		return outfileName, err
	}

	// *** The following can be written independent of the actual command. *** //
	// Write the script header.
	fmt.Fprintln(outfile, fmt.Sprintf("#!/bin/bash\n"))
	fmt.Fprintln(outfile, "ulimit -n 10000")

	// Write job shit.
	fmt.Fprintln(outfile, fmt.Sprintf("%s", datamodels.JOB_SHIT["singularity_cmd"]))
	fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["singularity_bind"], command.CommandParams.Volume)))
	fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["singularity_env"], command.CommandParams.SingularityPath, command.CommandParams.SingularityImage)))
	if command.CommandParams.Subcommand != "" {
		fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["command"], fmt.Sprintf("%s %s", command.CommandParams.Command, command.CommandParams.Subcommand))))
	} else {
		fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["command"], command.CommandParams.Command)))
	}

	// TODO: This needs to be wrapped in star options command or something. We will likely
	// have other tools that require specific option formattin.
	if command.CommandParams.Command == "star" {
		for _, opt := range command.CommandParams.CommandOptions {
			chunks := strings.Split(opt, " ")
			if chunks[0] == "--outFileNamePrefix" {
				opt = fmt.Sprintf("%s %s", chunks[0], fmt.Sprintf("%s/%s_", sample.OutputPath, sample.Prefix))
			}
			if chunks[0] == "--readFilesIn" {
				opt = fmt.Sprintf("%s %s", chunks[0], sample.DumpReadFiles())
			}
			WriteCommandOption(outfile, opt)
		}
	} else if command.CommandParams.Command == "trim_galore" {
		for _, opt := range command.CommandParams.CommandOptions {
			chunks := strings.Split(opt, " ")
			if chunks[0] == "--output_dir" {
				opt = fmt.Sprintf("%s %s", chunks[0], fmt.Sprintf("%s", sample.OutputPath))
			}
			WriteCommandOption(outfile, opt)
		}
	} else if command.CommandParams.Command == "kallisto" && command.CommandParams.Subcommand == "quant" {
		for _, opt := range command.CommandParams.CommandOptions {
			chunks := strings.Split(opt, " ")
			if chunks[0] == "--output-dir" {
				// Create the sample name directory.
				basePath := fmt.Sprintf("%s/%s", sample.OutputPath, sample.Prefix)
				// os.Mkdir(basePath, 0775)
				// Format the output option for kallisto quant
				opt = fmt.Sprintf("%s %s/kallisto_quant", chunks[0], basePath)
			}
			WriteCommandOption(outfile, opt)
		}
	} else {
		WriteCommandOptions(outfile, command.CommandParams.CommandOptions)
	}

	if command.CommandParams.Command == "trim_galore" {
		// Write the forward read file arg to the script.
		WriteCommandArg(outfile, sample.DumpForwardReadFileWithPath())
		// Write the reverse read file arg to the script.
		WriteCommandArg(outfile, sample.DumpReverseReadFileWithPath())
	} else if command.CommandParams.Command == "rsem-calculate-expression" {
		// First, we will write the readfiles argument.
		// We want trimmed reads here. So drop the file extention from the readfile name.
		noExt := true
		forwardReads := fmt.Sprintf("%s/%s_val_1.fq.gz", sample.OutputPath, sample.DumpForwardReadFile(noExt))
		reverseReads := fmt.Sprintf("%s/%s_val_2.fq.gz", sample.OutputPath, sample.DumpReverseReadFile(noExt))
		WriteCommandArg(outfile, fmt.Sprintf("%s", forwardReads))
		WriteCommandArg(outfile, fmt.Sprintf("%s", reverseReads))

		// Next we will write the reference argument. This will be supplied in the params.txt file
		WriteCommandArgs(outfile, command.CommandParams.CommandArgs)

		// Write the samplename arg
		sampleNameArg := fmt.Sprintf("%s/%s", sample.OutputPath, sample.Prefix)
		WriteCommandArg(outfile, sampleNameArg)
	} else if command.CommandParams.Command == "fastqc" {
		sequenceFilesArg := sample.DumpReadFiles()
		WriteCommandArg(outfile, sequenceFilesArg)
	} else if command.CommandParams.Command == "kallisto" && command.CommandParams.Subcommand == "quant" {
		noExt := true
		forwardReads := fmt.Sprintf("%s/%s_val_1.fq.gz", sample.OutputPath, sample.DumpForwardReadFile(noExt))
		reverseReads := fmt.Sprintf("%s/%s_val_2.fq.gz", sample.OutputPath, sample.DumpReverseReadFile(noExt))
		fmt.Println(forwardReads, reverseReads)
		WriteCommandArg(outfile, fmt.Sprintf("%s", forwardReads))
		WriteCommandArg(outfile, fmt.Sprintf("%s", reverseReads))
	} else {
		WriteCommandArgs(outfile, command.CommandParams.CommandArgs)
	}

	outfile.Close()
	return outfileName, nil

}

/* ---
 * Write the slurm job preamble to a .slurm file.
 * --- */
func writeSlurmJobPreamble(slurmFile *os.File, preamble datamodels.SlurmPreamble) {
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", datamodels.SLURM_PREAMBLE["header"]))
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["job_name"], preamble.JobName)))
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["partition"], preamble.Partition)))
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["notifications"], preamble.NotificationType())))
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["email"], preamble.EmailAddress)))
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["time"], preamble.WallTime)))
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["job_log"], preamble.JobName)))
	fmt.Fprintln(slurmFile)
}

func writeJobCPU(slurmFile *os.File, job datamodels.Job) {
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["cpus"], job.MaxCPUUsage())))
}

/* ---
 * Write the command preamble to a .slurm file.
 * --- */
func WriteCommandPreamble(slurmFile *os.File, preamble datamodels.CommandPreamble) {
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["tasks"], preamble.Tasks)))
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["cpus"], preamble.CPUs)))
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["memory"], preamble.Memory)))
	fmt.Fprintln(slurmFile)
}

/* --
 * Write the command options to file.
 * --- */
func WriteCommandOptions(outfile *os.File, options []string) {
	// Write all command options.
	for _, opt := range options {
		fmt.Fprintln(outfile, fmt.Sprintf("%s \\", opt))
	}
}

/* --
 * Write command option.
 * --- */
func WriteCommandOption(outfile *os.File, option string) {
	// Write a single command option to the file
	fmt.Fprintln(outfile, fmt.Sprintf("%s \\", option))
}

/* ---
 * Write the command args to file.
 * --- */
func WriteCommandArgs(outfile *os.File, args []string) {
	// Write all command options.
	for _, opt := range args {
		fmt.Fprintln(outfile, fmt.Sprintf("%s \\", opt))
	}
}

func WriteCommandArg(outfile *os.File, arg string) {
	// Write a single command arg.
	fmt.Fprintln(outfile, fmt.Sprintf("%s \\", arg))
}

func WriteWait(outfile *os.File) {
	// Write a single command arg.
	fmt.Fprintln(outfile, "wait")
}
