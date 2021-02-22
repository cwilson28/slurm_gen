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
	var platform string
	var sge, slurm, submit, preflight bool

	// Declare command line flags.
	flag.Bool("help", false, "Show help message")
	flag.Bool("submit", false, "Submit job on the user's behalf")
	flag.Bool("preflight", false, "Run all preflight tests")
	flag.Bool("slurm", false, "Generate scripts for a Slurm cluster")
	flag.Bool("sge", false, "Generate scripts for a SGE cluster")
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
	 * Check for the submit flag
	 * ---------------------------------------------------------------------- */
	submitFlag := flag.Lookup("submit")
	if submitFlag.Value.String() == "true" {
		submit = true
	}

	/* -------------------------------------------------------------------------
	 * Check for the slurm flag
	 * ---------------------------------------------------------------------- */
	slurmFlag := flag.Lookup("slurm")
	if slurmFlag.Value.String() == "true" {
		slurm = true
		platform = "slurm"
	}

	/* -------------------------------------------------------------------------
	 * Check for the sge flag
	 * ---------------------------------------------------------------------- */
	sgeFlag := flag.Lookup("sge")
	if sgeFlag.Value.String() == "true" {
		sge = true
		platform = "sge"
	}

	if sge == false && slurm == false {
		log.Print("Error: You must specify a compute environment.")
		ShowHelp()
		os.Exit(1)
	}

	/* -------------------------------------------------------------------------
	 * Check for the preflight flag
	 * ---------------------------------------------------------------------- */
	preflightFlag := flag.Lookup("preflight")
	if preflightFlag.Value.String() == "true" {
		preflight = true
	}

	/* -------------------------------------------------------------------------
	 * Get the param file from the command line. It should be the only elem in
	 * flag.Args()
	 * ---------------------------------------------------------------------- */

	if len(flag.Args()) < 2 {
		log.Fatal(fmt.Sprintf("Error: Wrong number of args. \nExpecting: commander <path_to_param_file.json> <path_to_design_file.txt>\nReceived: commander %s", flag.Args()[0]))
		os.Exit(1)
	}

	// Grab the parameter file from the command line.
	paramFile := flag.Args()[0]

	// Grab the design file from the command line.
	designFile := flag.Args()[1]

	/* -------------------------------------------------------------------------
	 * We are supporting both plain text and json param files.
	 * ---------------------------------------------------------------------- */

	// Set the platform variable in the utils package.
	utils.Platform = platform

	// Create the primary job object.
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

	// Parse the design file.
	experiment := utils.ParseDesignFile(designFile)

	// Perform preflight experiment path checks.
	if preflight {
		err := utils.PreflightTests(experiment, job)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	// Write the job files.
	if slurm {
		fmt.Println("Writing slurm job script...")
		err = utils.WriteSlurmJobScript(job, experiment)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		fmt.Println("Done.")
	} else if sge {
		fmt.Println("Writing sge job script...")
		err = utils.WriteSGEJobScript(job, experiment)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		fmt.Println("Done.")
	}

	if submit {
		fmt.Println("submit")
	}
}
