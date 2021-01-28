package datamodels

var SLURM_PREAMBLE = map[string]string{
	"header":        "#!/bin/bash",
	"job_name":      "#SBATCH --job-name=%s",
	"partition":     "#SBATCH --partition=%s",
	"notifications": "#SBATCH --mail-type=%s",
	"email":         "#SBATCH --mail-user=%s",
	"tasks":         "#SBATCH --ntasks=%d",
	"cpus":          "#SBATCH --cpus-per-task=%d",
	"memory":        "#SBATCH --mem=%d",
	"time":          "#SBATCH --time=%s",
	"job_log":       "#SBATCH --output=%s_%%j.log",
}

var SGE_PREAMBLE = map[string]string{
	"header":               "#!/bin/bash",
	"cwd":                  "#$ -cwd",
	"join_output":          "#$ -j y",
	"shell":                "#$ -S %s",
	"email":                "#$ -M %s -m be",
	"parallel_environment": "#$ -pe %s",
	"memory":               "#$ -l %s",
}

var COMMAND_PREAMBLE = map[string]string{
	"job_name": "#SBATCH --job-name=%s",
	"tasks":    "#SBATCH --ntasks=%d",
	"cpus":     "#SBATCH --cpus-per-task=%d",
	"memory":   "#SBATCH --mem=%d",
	"time":     "#SBATCH --time=%s",
}

var MORE_JOBSHIT = []string{
	"",
	`echo "====================================================="`,
	"pwd; hostname; date",
	`echo "====================================================="`,
	"",
	"# Load the compbio module",
	"module load compbio",
	"# Load the singularity module",
	"module load singularity",
	"export LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH",
	"",
}

var JOB_SHIT = map[string]string{
	"singularity_cmd":  "singularity run \\",
	"singularity_bind": "--bind %s:/compbio \\",
	"singularity_env":  "%s/%s \\",
	"command":          "%s \\",
}

var HELP_MSG = `
	Usage: commander [--options] <param_file>

	Summary: commander is a command line tool for generating reproducible
	bioinformatics tools scripts that can be run in different computational
	environments. 

	Currently, commander can generate scripts for Slurm and SGE clusters.

	Options:
	--slurm: Tells commander that the scripts should be written for submission to a Slurm cluster.
	--sge:   Tells commander that the scripts should be written for submission to a SGE cluster

	Arguments:
	A single parameter file that defines the workflow to be executed.
	
	Example usage:

	# With parameters specified in JSON format.
	commander --slurm worflow-params.json

	# With parameters specified in plaintext format.
	commander --sge workflow-params.txt

	Output:
	If the --slurm option is provided, commander will produce a main .slurm file that can
	be submitted to a Slurm cluster using sbatch.

	If the --sge options is provided, commander will produce a main .qsub file that can be
	submitted to a SGE cluster using qsub.
`
