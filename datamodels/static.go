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

var MORE_JOBSHIT = []string{
	"",
	`echo "====================================================="`,
	"pwd; hostname; date",
	`echo "====================================================="`,
	"",
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
	Usage: slurm_gen [--options]

	Summary: slurm_gen is a command line tool for generating scripts to be submitted
	to the slurm scheduler with sbatch.

	Available Options:
	
	--params: full path to the parameters file defining your tool pipeline
	--pipeline: will invoke the pipeline option and will generate a full pipeline script

	Example usage:

	slurm_gen --params /home/username/fastqc_params.txt
	Output: fastqc.slurm
`
