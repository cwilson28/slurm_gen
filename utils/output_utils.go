package utils

import (
	"fmt"
	"os"
	"slurm_gen/datamodels"
	"strings"
)

/* -----------------------------------------------------------------------------
 * Functions for writing output files for the slurm job.
 * -------------------------------------------------------------------------- */

/* ---
 * Write a parent slurm file.
 * --- */
func WriteSlurmScript(slurmFile *os.File, job datamodels.Job) {
	// Write the slurm preamble for the parent slurm script
	WriteSlurmPreamble(slurmFile, job.Commands[0].CommandName, job.SlurmPreamble)
	// Write the command preamble. We can use Commands[0] since there *should* only be one command.
	WriteCommandPreamble(slurmFile, job.Commands[0].Preamble)

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

	WriteCommandOptions(slurmFile, command.CommandParams.CommandOptions)
	WriteCommandArgs(slurmFile, command.CommandParams.CommandArgs)
}

/* ---
 * Write the command to a bash file.
 * --- */
func WriteBashScript(outfile *os.File, command datamodels.Command) (string, error) {
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

	WriteCommandOptions(outfile, command.CommandParams.CommandOptions)
	WriteCommandArgs(outfile, command.CommandParams.CommandArgs)
	outfile.Close()
	fmt.Printf("%s.sh written successfully\n", command.CommandName)
	return filename, nil
}

func WriteBatchBashScript(outfile *os.File, command datamodels.Command, sample datamodels.Sample) (string, error) {
	filename := fmt.Sprintf("%s_%s.sh", command.CommandName, sample.Prefix)
	outfile, err := os.Create(filename)
	if err != nil {
		return filename, err
	}

	// *** The following can be written independent of the actual command. *** //
	// Write the script header.
	fmt.Fprintln(outfile, fmt.Sprintf("#!/bin/bash\n"))
	fmt.Fprintln(outfile, "ulimit -n 10000")

	// Write job shit.
	fmt.Fprintln(outfile, fmt.Sprintf("%s", datamodels.JOB_SHIT["singularity_cmd"]))
	fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["singularity_bind"], command.CommandParams.Volume)))
	fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["singularity_env"], command.CommandParams.SingularityPath, command.CommandParams.SingularityImage)))
	fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["command"], command.CommandParams.Command)))

	// TODO: This needs to be wrapped in star options command or something. We will likely
	// have other tools that require specific option formattin.
	if command.CommandName == "star" {
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
	} else if command.CommandName == "trim_galore" {
		for _, opt := range command.CommandParams.CommandOptions {
			chunks := strings.Split(opt, " ")
			if chunks[0] == "--output_dir" {
				opt = fmt.Sprintf("%s %s", chunks[0], fmt.Sprintf("%s", sample.OutputPath))
			}
			WriteCommandOption(outfile, opt)
		}
	} else {
		WriteCommandOptions(outfile, command.CommandParams.CommandOptions)
	}

	if command.CommandName == "trim_galore" {
		// Write the forward read file arg to the script.
		WriteCommandArg(outfile, sample.DumpForwardReadFileWithPath())
		// Write the reverse read file arg to the script.
		WriteCommandArg(outfile, sample.DumpReverseReadFileWithPath())
	} else if command.CommandName == "rsem-calculate-expression" {
		// First, we will write the readfiles argument.
		// We want trimmed reads here. So drop the file extention from the readfile name.
		noExt := true
		forwardReads := fmt.Sprintf("%s/%s_.trimmed.fq.gz", sample.OutputPath, sample.DumpForwardReadFile(noExt))
		reverseReads := fmt.Sprintf("%s/%s_.trimmed.fq.gz", sample.OutputPath, sample.DumpReverseReadFile(noExt))
		readFilesArg := fmt.Sprintf("%s %s", forwardReads, reverseReads)
		WriteCommandArg(outfile, readFilesArg)

		// Next we will write the reference argument. This will be supplied in the params.txt file
		WriteCommandArgs(outfile, command.CommandParams.CommandArgs)

		// Write the samplename arg
		sampleNameArg := fmt.Sprintf("%s/%s", sample.OutputPath, sample.Prefix)
		WriteCommandArg(outfile, sampleNameArg)
	} else {
		WriteCommandArgs(outfile, command.CommandParams.CommandArgs)
	}

	outfile.Close()
	fmt.Printf("%s written successfully\n", filename)
	return filename, nil

}

/* ---
 * Write the slurm preamble to a .slurm file.
 * --- */
func WriteSlurmPreamble(slurmFile *os.File, jobname string, preamble datamodels.SlurmPreamble) {
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", datamodels.SLURM_PREAMBLE["header"]))
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["job_name"], jobname)))
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["partition"], preamble.Partition)))
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["notifications"], preamble.NotificationType())))
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["email"], preamble.NotificationEmail)))
	fmt.Fprintln(slurmFile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.SLURM_PREAMBLE["job_log"], jobname)))
	fmt.Fprintln(slurmFile)
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
