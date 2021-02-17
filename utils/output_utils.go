package utils

import (
	"commander/datamodels"
	"fmt"
	"os"
	"strings"
)

/* -----------------------------------------------------------------------------
 * The main function for writing the slurm script and accompanying bash scripts.
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
	writeSlurmCommandPreamble(slurmFile, cmd.Preamble)
	writeCommandOptions(slurmFile, cmd.CommandParams.CommandOptions)
	writeCommandArgs(slurmFile, cmd.CommandParams.CommandArgs)
	return nil
}

/* -----------------------------------------------------------------------------
 * The main function for writing an SGE script.
 * -------------------------------------------------------------------------- */
func WriteSGEJobScript(job datamodels.Job) error {
	var err error

	fmt.Println("Writing sge script preamble...")

	// Open the parent sge script
	// filename := fmt.Sprintf("%s.qsub", job.SGEPreamble.JobName)
	filename := "testfile.qsub"
	sgeFile, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer func() {
		if err = sgeFile.Close(); err != nil {
			panic(err)
		}
	}()

	// Write the slurm preamble for the parent slurm script
	writeSGEJobPreamble(sgeFile, job.SGEPreamble)

	/* -----------------
	 * All slurm preamble is written at this point. All that remains is to write
	 * command specific jargon.
	 * -------------- */

	// If there are multiple commands, it is safe to assume we are generating
	// a slurm script for a pipeline. Write the command details in a pipeline
	// format.
	// if len(job.Commands) > 1 {
	// 	fmt.Println("Writing pipeline slurm script...")
	// 	err = writePipelineSlurmScript(sgeFile, job)
	// 	return err
	// }

	// There is a single command, we will either write this as as single .slurm
	// file or as a batch slurm file depending on the command definition.
	// Since there is only a single command, grab the 0th command object.
	cmd := job.Commands[0]

	// If this is a batch command, write the required batch scripts.
	// TODO: Make this batch bash script writing into a function. It's being used
	// more than once.
	if cmd.Batch {
		fmt.Println("Writing command script...")
		err = writeBatchCommand(sgeFile, cmd, job)
		return err
	}

	// TODO: Revisit this.
	// The command is not a batch command, write the command to the slurm file
	// we opened earlier.
	writeSingularityPreamble(sgeFile, cmd)
	writeCommandOptions(sgeFile, cmd.CommandParams.CommandOptions)
	writeCommandArgs(sgeFile, cmd.CommandParams.CommandArgs)
	return nil
}

/* -----------------------------------------------------------------------------
 * Various helper functions.
 * -------------------------------------------------------------------------- */

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

	// Write any misc preamble
	writeMiscPreamble(slurmFile, preamble.MiscPreamble)
}

/* ---
 * Write the sge job preamble to a .qsub file.
 * --- */
func writeSGEJobPreamble(sgeFile *os.File, preamble datamodels.SGEPreamble) {
	fmt.Fprintln(sgeFile, fmt.Sprintf("%s", datamodels.SGE_PREAMBLE["header"]))
	/* --- Handle boolean flags in the parameter file. --- */
	// Use current working directory
	if preamble.CWD {
		fmt.Fprintln(sgeFile, fmt.Sprintf(datamodels.SGE_PREAMBLE["cwd"]))
	}
	// Join output
	if preamble.JoinOutput {
		fmt.Fprintln(sgeFile, fmt.Sprintf(datamodels.SGE_PREAMBLE["join_output"]))
	}

	fmt.Fprintln(sgeFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SGE_PREAMBLE["shell"], preamble.Shell)))
	fmt.Fprintln(sgeFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SGE_PREAMBLE["email"], preamble.EmailAddress)))
	fmt.Fprintln(sgeFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SGE_PREAMBLE["parallel_environment"], preamble.ParallelEnv)))
	fmt.Fprintln(sgeFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SGE_PREAMBLE["memory"], preamble.Memory)))
	fmt.Fprintln(sgeFile)

	// Write any misc preamble
	writeMiscPreamble(sgeFile, preamble.MiscPreamble)
}

/* ---
 * Write misc preamble
 * --- */
func writeMiscPreamble(sgeFile *os.File, preamble []string) {
	for _, line := range preamble {
		fmt.Fprintln(sgeFile, fmt.Sprintf(line))
	}
	fmt.Fprintln(sgeFile)
}

/* ---
 * Write intermidate job shit to the slurm file.slurmFile :-)
 * This will probably be omitted in the future.
 *  --- */
func writeIntermediateJobShit(slurmFile *os.File) {
	for _, line := range datamodels.INTERMEDIATE_SLURM_SHIT {
		fmt.Fprintln(slurmFile, line)
	}
}

/* ---
 * Write the CPU count to the slurm file.
 * --- */
func writeJobCPU(slurmFile *os.File, job datamodels.Job) {
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["cpus"], job.MaxCPUUsage())))
}

/* ---
 * Write the remaining contents of a pipeline slurm script. This function will
 * also generate the individual bash scripts for the commands being executed.
 * --- */
func writePipelineSlurmScript(slurmFile *os.File, job datamodels.Job) error {
	fmt.Println("Writing pipeline scripts...")
	// Write the bash scripts for each command.
	for _, cmd := range job.Commands {
		if cmd.Batch {
			// User has indicated the command will be run in a batch format.
			err := writeBatchCommand(slurmFile, cmd, job)
			if err != nil {
				return err
			}
		} else {
			fmt.Println("Writing command script...")
			// Write the bash script for the command.
			bashScript, err := writeCommandScript(cmd)
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
	}
	return nil
}

/* ---
 * Finish writing slurm file given a single batch command.
 * --- */
func writeBatchCommand(slurmFile *os.File, cmd datamodels.Command, job datamodels.Job) error {
	fmt.Println("Command is a batch command.")
	fmt.Println("Writing batch bash scripts...")
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
	writeWait(slurmFile)
	return nil
}

/* ---
 * Write the command to a bash file.
 * --- */
func writeCommandScript(command datamodels.Command) (string, error) {
	// Write a bash script for each sample.
	scriptName := fmt.Sprintf("%s.sh", command.CommandParams.Command)
	outfile, err := os.Create(scriptName)
	if err != nil {
		return scriptName, err
	}

	// Defer the file closing until the function returns.
	defer outfile.Close()

	// Write the script header.
	writeBashScriptHeader(outfile)

	// Write singularity preamble.
	writeSingularityPreamble(outfile, command)

	// Write the command we are calling.
	if command.CommandParams.Subcommand != "" {
		fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["command"], fmt.Sprintf("%s %s", command.CommandParams.Command, command.CommandParams.Subcommand))))
	} else {
		fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["command"], command.CommandParams.Command)))
	}

	// Write the command options and command arguments
	writeCommandOptions(outfile, command.CommandParams.CommandOptions)
	writeCommandArgs(outfile, command.CommandParams.CommandArgs)

	return scriptName, nil
}

/* ---
 * Write bash script header to a file.
 * --- */
func writeBashScriptHeader(outfile *os.File) {
	// Write the script header.
	fmt.Fprintln(outfile, fmt.Sprintf("#!/bin/bash\n"))
	fmt.Fprintln(outfile, "ulimit -n 10000")
}

/* ---
 * Write singularity preamble for the command we are calling.
 * --- */
func writeSingularityPreamble(outfile *os.File, cmd datamodels.Command) {
	// Write singularity shit.
	fmt.Fprintln(outfile, fmt.Sprintf("%s", datamodels.JOB_SHIT["singularity_cmd"]))
	fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["singularity_bind"], cmd.CommandParams.Volume)))
	// fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["singularity_env"], cmd.CommandParams.SingularityImage)))
	fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["singularity_env"], cmd.CommandParams.SingularityPath, cmd.CommandParams.SingularityImage)))
}

/* ---
 * Format and write STAR specific options.
 * --- */
func writeStarCommandOptions(outfile *os.File, command datamodels.Command, sample datamodels.Sample) {
	// Star has specific formatting for certain options. Write them here.
	for _, opt := range command.CommandParams.CommandOptions {
		chunks := strings.Split(opt, " ")
		if chunks[0] == "--outFileNamePrefix" {
			opt = fmt.Sprintf("%s %s", chunks[0], fmt.Sprintf("%s/%s_", sample.OutputPath, sample.Prefix))
		} else if chunks[0] == "--readFilesIn" {
			opt = fmt.Sprintf("%s %s", chunks[0], sample.DumpReadFiles())
		}
		writeCommandOption(outfile, opt)
	}
}

/* ---
 * Format and write trim_galore specific options.
 * --- */
func writeTrimGaloreCommandOptions(outfile *os.File, command datamodels.Command, sample datamodels.Sample) {
	// TrimGalore has specific formatting for certain options. Write them here.
	for _, opt := range command.CommandParams.CommandOptions {
		chunks := strings.Split(opt, " ")
		if chunks[0] == "--output_dir" {
			opt = fmt.Sprintf("%s %s", chunks[0], fmt.Sprintf("%s", sample.OutputPath))
		}
		writeCommandOption(outfile, opt)
	}
}

/* ---
 * Format and write kallisto specific options.
 * --- */
func writeKallistoQuantOptions(outfile *os.File, command datamodels.Command, sample datamodels.Sample) {
	for _, opt := range command.CommandParams.CommandOptions {
		chunks := strings.Split(opt, " ")
		if chunks[0] == "--output-dir" {
			// Create the sample name directory.
			basePath := fmt.Sprintf("%s/%s", sample.OutputPath, sample.Prefix)
			// os.Mkdir(basePath, 0775)
			// Format the output option for kallisto quant
			opt = fmt.Sprintf("%s %s/kallisto_quant", chunks[0], basePath)
		}
		writeCommandOption(outfile, opt)
	}
}

/* ---
 * Format and write trim_galore specific arguments.
 * --- */
func writeTrimGaloreArguments(outfile *os.File, sample datamodels.Sample) {
	// Write the forward read file arg to the script.
	writeCommandArg(outfile, sample.DumpForwardReadFileWithPath())

	if sample.IsPairedEnd() {
		// Write the reverse read file arg to the script.
		writeCommandArg(outfile, sample.DumpReverseReadFileWithPath())
	}
}

/* ---
 * Format and write rsem specific arguments.
 * --- */
func writeRSEMArguments(outfile *os.File, command datamodels.Command, sample datamodels.Sample) {
	// TODO: Revisit this. It can be improved.
	// First, we will write the readfiles argument.
	// We want trimmed reads here. So drop the file extention from the readfile name.
	noExt := true
	if sample.IsPairedEnd() {
		forwardReads := fmt.Sprintf("%s/%s_val_1.fq.gz", sample.OutputPath, sample.DumpForwardReadFile(noExt))
		writeCommandArg(outfile, fmt.Sprintf("%s", forwardReads))
		reverseReads := fmt.Sprintf("%s/%s_val_2.fq.gz", sample.OutputPath, sample.DumpReverseReadFile(noExt))
		writeCommandArg(outfile, fmt.Sprintf("%s", reverseReads))
	} else {
		forwardReads := fmt.Sprintf("%s/%s_trimmed.fq.gz", sample.OutputPath, sample.DumpForwardReadFile(noExt))
		writeCommandArg(outfile, fmt.Sprintf("%s", forwardReads))
	}

	// Next we will write the reference argument. This will be supplied in the params.txt file
	writeCommandArgs(outfile, command.CommandParams.CommandArgs)

	// Write the samplename arg
	sampleNameArg := fmt.Sprintf("%s/%s", sample.OutputPath, sample.Prefix)
	writeCommandArg(outfile, sampleNameArg)
}

/* ---
 * Format and write fastqc specific arguments.
 * --- */
func writeFastQCArguments(outfile *os.File, sample datamodels.Sample) {
	sequenceFilesArg := sample.DumpReadFiles()
	writeCommandArg(outfile, sequenceFilesArg)
}

/* ---
 * Format and write kallisto quant specific arguments.
 * --- */
func writeKallistoQuantArguments(outfile *os.File, sample datamodels.Sample) {
	noExt := true
	forwardReads := fmt.Sprintf("%s/%s_val_1.fq.gz", sample.OutputPath, sample.DumpForwardReadFile(noExt))
	reverseReads := fmt.Sprintf("%s/%s_val_2.fq.gz", sample.OutputPath, sample.DumpReverseReadFile(noExt))
	writeCommandArg(outfile, fmt.Sprintf("%s", forwardReads))
	writeCommandArg(outfile, fmt.Sprintf("%s", reverseReads))
}

/* ---
 * Write a command script for a given command given a particular sample.
 *  --- */
func writeCommandScriptForSample(command datamodels.Command, sample datamodels.Sample) (string, error) {
	outfileName := fmt.Sprintf("%s_%s.sh", command.CommandParams.Command, sample.Prefix)
	outfile, err := os.Create(outfileName)
	if err != nil {
		return outfileName, err
	}

	// Defer file closing until after the function returns.
	defer outfile.Close()

	// Write the header lines to the bash script.
	writeBashScriptHeader(outfile)

	// Write the singuarity command preamble.
	writeSingularityPreamble(outfile, command)

	// Write the command we are calling. If there is a subcommand (e.g., kallisto "quant") include it!
	if command.CommandParams.Subcommand != "" {
		fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["command"], fmt.Sprintf("%s %s", command.CommandParams.Command, command.CommandParams.Subcommand))))
	} else {
		fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["command"], command.CommandParams.Command)))
	}

	// Write command options.
	if command.CommandParams.Command == "STAR" {
		// Format and write star specific options
		writeStarCommandOptions(outfile, command, sample)
	} else if command.CommandParams.Command == "trim_galore" {
		// Format and write trim_galore specific options
		writeTrimGaloreCommandOptions(outfile, command, sample)
	} else if command.CommandParams.Command == "kallisto" && command.CommandParams.Subcommand == "quant" {
		// Format and write kallisto quant specific options
		writeKallistoQuantOptions(outfile, command, sample)
	} else {
		// Otherwise, write the option as it was provided.
		writeCommandOptions(outfile, command.CommandParams.CommandOptions)
	}

	// Write command args.
	if command.CommandParams.Command == "trim_galore" {
		// Format and write trim_galore arguments.
		writeTrimGaloreArguments(outfile, sample)
	} else if command.CommandParams.Command == "rsem-calculate-expression" {
		// Format and write rsem arguments.
		writeRSEMArguments(outfile, command, sample)
	} else if command.CommandParams.Command == "fastqc" {
		// Format and write fastqc arguments.
		writeFastQCArguments(outfile, sample)
	} else if command.CommandParams.Command == "kallisto" && command.CommandParams.Subcommand == "quant" {
		// Format and write kallisto quant arguments.
		writeKallistoQuantArguments(outfile, sample)
	} else {
		// Otherwise, write the arguments as they were provided.
		writeCommandArgs(outfile, command.CommandParams.CommandArgs)
	}
	return outfileName, nil
}

/* ---
 * Write the command preamble to a .slurm file.
 * --- */
func writeSlurmCommandPreamble(slurmFile *os.File, preamble datamodels.CommandPreamble) {
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["tasks"], preamble.Tasks)))
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["cpus"], preamble.CPUs)))
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["memory"], preamble.Memory)))
	fmt.Fprintln(slurmFile)
}

/* --
 * Write all command options to file.
 * --- */
func writeCommandOptions(outfile *os.File, options []string) {
	// Write all command options.
	for _, opt := range options {
		fmt.Fprintln(outfile, fmt.Sprintf("%s \\", opt))
	}
}

/* --
 * Write a single command option to file.
 * --- */
func writeCommandOption(outfile *os.File, option string) {
	// Write a single command option to the file
	fmt.Fprintln(outfile, fmt.Sprintf("%s \\", option))
}

/* ---
 * Write all command args to file.
 * --- */
func writeCommandArgs(outfile *os.File, args []string) {
	// Write all command options.
	for _, opt := range args {
		fmt.Fprintln(outfile, fmt.Sprintf("%s \\", opt))
	}
}

/* ---
 * Write a single command arg to a file.
 * --- */
func writeCommandArg(outfile *os.File, arg string) {
	// Write a single command arg.
	fmt.Fprintln(outfile, fmt.Sprintf("%s \\", arg))
}

/* ---
 * Write a single wait block to the slurm file.
 * --- */
func writeWait(outfile *os.File) {
	// Write a single command arg.
	fmt.Fprintln(outfile, "wait")
}
