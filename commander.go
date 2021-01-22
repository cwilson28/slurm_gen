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
		fmt.Println("Parsing JSON parameter file...")
		job, err = utils.ParseJSONParams(paramFile)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	} else {
		fmt.Println("Parsing plain text parameter file...")
		job, err = utils.ParsePlainTextParams(paramFile)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}

	// Write the job files.
	fmt.Println("Writing job scripts...")
	err = utils.WriteSlurmJobScript(job)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	fmt.Println("Done.")
}
