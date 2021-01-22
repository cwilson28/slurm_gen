package main

import (
	"commander/datamodels"
	"commander/utils"
	"flag"
	"fmt"
	"log"
	"os"
)

// Show the help message for commander
func ShowHelp() {
	fmt.Println(datamodels.HELP_MSG)
}

func main() {
	var err error
	var job datamodels.Job
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
		fmt.Println(submit)
	}

	/* -------------------------------------------------------------------------
	 * Get the param file from the command line. It should be the only elem in
	 * flag.Args()
	 * ---------------------------------------------------------------------- */

	if len(flag.Args()) != 1 {
		log.Fatal("Error: Wrong number of args.")
		os.Exit(1)
	}

	paramFile := flag.Args()[0]

	/* -------------------------------------------------------------------------
	 * We are supporting both plain text and json param files.
	 * ---------------------------------------------------------------------- */

	if utils.IsJSONParam(paramFile) {
		job, err = utils.ParseJSONParams(paramFile)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	} else {
		job, err = utils.ParsePlainTextParams(paramFile)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}

	// Write the job files.
	err = utils.WriteSlurmJobScript(job)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// // Generate a slurm script for a single tool.
	// fmt.Printf("Generating slurm script for %s...\n", cmd.CommandName)

	// // This is a batch command.
	// // Write multiple slurm scripts.
	// if cmd.Batch {
	// 	fmt.Printf("%s will be run in batch mode.\n", cmd.CommandName)
	// 	fmt.Println("Writing batch files...")

	// 	batchFiles := make([]string, 0)

	// 	// Get the samples over which this command will be run
	// 	samples := utils.ParseSamplesFile(cmd.SamplesFile)

	// 	for _, s := range samples {
	// 		// We are going to write a slurm script for each
	// 		filename := fmt.Sprintf("%s_%s.slurm", cmd.CommandName, s.Prefix)
	// 		outfile, err := os.Create(filename)
	// 		if err != nil {
	// 			panic(err)
	// 		}

	// 		defer func() {
	// 			if err = outfile.Close(); err != nil {
	// 				log.Fatal(err)
	// 			}
	// 		}()

	// 		utils.writeSlurmJobPreamble(outfile, strings.TrimRight(filename, ".slurm"), job.SlurmPreamble)
	// 		utils.WriteCommandPreamble(outfile, cmd.Preamble)
	// 		// Write intermediary job shit.
	// 		// TODO: Wrap this in a "Write" command in the output_utils file.
	// 		for _, line := range datamodels.MORE_JOBSHIT {
	// 			fmt.Fprintln(outfile, line)
	// 		}

	// 		fmt.Fprintln(outfile, "ulimit -n 10000")

	// 		// TODO: Wrap this in a "Write" command in the output_utils file.
	// 		fmt.Fprintln(outfile, fmt.Sprintf("%s", datamodels.JOB_SHIT["singularity_cmd"]))
	// 		fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["singularity_bind"], cmd.CommandParams.Volume)))
	// 		fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["singularity_env"], cmd.CommandParams.SingularityPath, cmd.CommandParams.SingularityImage)))
	// 		fmt.Fprintln(outfile, fmt.Sprintf("%s", fmt.Sprintf(datamodels.JOB_SHIT["command"], cmd.CommandParams.Command)))

	// 		// TODO: This needs to be wrapped in star options command or something. We will likely
	// 		// have other tools that require specific option formattin.
	// 		if cmd.CommandName == "star" {
	// 			for _, opt := range cmd.CommandParams.CommandOptions {
	// 				chunks := strings.Split(opt, " ")
	// 				if chunks[0] == "--outFileNamePrefix" {
	// 					opt = fmt.Sprintf("%s %s", chunks[0], fmt.Sprintf("%s/%s_", s.OutputPath, s.Prefix))
	// 				}
	// 				if chunks[0] == "--readFilesIn" {
	// 					opt = fmt.Sprintf("%s %s", chunks[0], s.DumpReadFiles())
	// 				}
	// 				utils.writeCommandOption(outfile, opt)
	// 			}
	// 		} else {
	// 			utils.writeCommandOptions(outfile, cmd.CommandParams.CommandOptions)
	// 		}

	// 		// Write any command args that are provided.
	// 		utils.writeCommandArgs(outfile, cmd.CommandParams.CommandArgs)
	// 		fmt.Printf("%s written successfully\n", filename)
	// 		batchFiles = append(batchFiles, filename)
	// 	}
	// 	if submit {
	// 		// Submit the jobs
	// 		for _, f := range batchFiles {
	// 			_, err := exec.Command("sbatch", f).Output()
	// 			if err != nil {
	// 				fmt.Printf("%s", err)
	// 			}
	// 		}
	// 	}
	// 	os.Exit(0)
	// }

	// // Name the slurm file after the command.
	// filename := fmt.Sprintf("%s.slurm", cmd.CommandName)
	// outfile, err := os.Create(filename)
	// if err != nil {
	// 	panic(err)
	// }

	// defer func() {
	// 	if err = outfile.Close(); err != nil {
	// 		log.Fatal(err)
	// 	}
	// }()

	// // Write the slurm script
	// utils.WriteSlurmScript(outfile, job)
	// fmt.Printf("Slurm script %s written.", filename)
	// os.Exit(0)

	// // Check for submit flag
	// submitFlag := flag.Lookup("submit")
	// if submitFlag == nil {
	// 	os.Exit(0)
	// }

	// Submit the job.
}
