package utils

import (
	"fmt"
	"os"
	"slurm_gen/datamodels"
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

/* ---
 * Write the command options to file.
 * --- */
func WriteCommandOptions(outfile *os.File, options []string) {
	// Write all command options.
	for _, opt := range options {
		fmt.Fprintln(outfile, fmt.Sprintf("%s \\", opt))
	}
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
