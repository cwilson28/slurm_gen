package datamodels

import (
	"fmt"
	"strings"
)

type SlurmPreamble struct {
	JobName      string
	Partition    string
	EmailBegin   bool
	EmailEnd     bool
	EmailFail    bool
	EmailAddress string
	WallTime     string
	MiscPreamble []string
}
type SGEPreamble struct {
	CWD          bool
	JoinOutput   bool
	Shell        string
	EmailAddress string
	ParallelEnv  string
	Memory       string
	MiscPreamble []string
}

type MiscPreamble struct {
	Lines []string
}

type CommandPreamble struct {
	Tasks  int64
	CPUs   int64
	Memory int64
}

type CommandParams struct {
	SingularityPath  string
	SingularityImage string
	WorkDir          string
	Volume           string
	Command          string
	Subcommand       string
	CommandOptions   []string
	CommandArgs      []string
}

type Job struct {
	Details           JobDetails
	ExperimentDetails Experiment
	SlurmPreamble     SlurmPreamble
	SGEPreamble       SGEPreamble
	MiscPreamble      MiscPreamble
	Commands          []Command
	CleanupActions    []CleanupAction
}

type JobDetails struct {
	Name       string
	DesignFile string
}

type Command struct {
	Batch            bool
	SamplesFile      string
	InputFromStep    string
	InputPathPrefix  string
	OutputPathPrefix string
	Preamble         CommandPreamble
	CommandParams    CommandParams
}

type BatchParams struct {
	SamplePrefix string
	ForwardReads []string
	ReverseReads []string
}

type Experiment struct {
	PI           string
	Name         string
	AnalysisID   string
	SamplePath   string
	AnalysisPath string
	WorkDir      string
	SamplesFile  string
	Samples      []Sample
}

type Sample struct {
	SamplePath      string
	OutputPath      string
	Prefix          string
	ForwardReadFile string
	ReverseReadFile string
}

type CleanupAction struct {
	ToolName    string
	Action      string
	Source      string
	Destination string
}

func (j *Job) IsPipeline() bool {
	if len(j.Commands) > 1 {
		return true
	}
	return false
}

func (j *Job) InitializeCMDIOPaths() {
	for i := range j.Commands {

		if j.Commands[i].InputFromStep == "" {
			// Command does not expext input as output from a previous command.
			// Use raw sample path as input path prefix.
			j.Commands[i].InputPathPrefix = j.ExperimentDetails.PrintRawSamplePath()
		} else {
			// Otherwise, the input is the output of a previous step in the
			// pipeline. The input path should reflect this
			j.Commands[i].InputPathPrefix = fmt.Sprintf("%s/%s", j.ExperimentDetails.PrintAnalysisPath(), j.Commands[i].InputFromStep)
		}

		// The output path prefix should account for the tool name
		j.Commands[i].OutputPathPrefix = fmt.Sprintf("%s/%s", j.ExperimentDetails.PrintAnalysisPath(), j.Commands[i].CommandName())
	}
}

func (j *Job) FormatCleanupActions() []string {
	var cleanupActions = make([]string, 0)

	for _, cua := range j.CleanupActions {
		sourcePath := fmt.Sprintf("%s/%s/%s", j.ExperimentDetails.PrintAnalysisPath(), cua.ToolName, cua.Source)
		if cua.Action == "mv" || cua.Action == "cp" {
			destPath := fmt.Sprintf("%s/%s", j.ExperimentDetails.PrintAnalysisPath(), cua.Destination)
			actionString := fmt.Sprintf("%s %s %s", cua.Action, sourcePath, destPath)
			cleanupActions = append(cleanupActions, actionString)
		} else {
			actionString := fmt.Sprintf("%s %s", cua.Action, sourcePath)
			cleanupActions = append(cleanupActions, actionString)
		}
	}
	return cleanupActions
}

/* --- Class functions --- */
func DefaultExperiment() Experiment {
	// Default experiment will initialize a compbio directory in the directory
	// where it is being run. Alternate paths should be provided in the "experiment"
	// declaration in the accompanying param file.
	return Experiment{
		PI:           "COMMANDER",
		Name:         "COMMANDER_TEST",
		SamplePath:   "./compbio/data",
		AnalysisPath: "./compbio/analysis",
		AnalysisID:   "1234567890",
	}
}

func (e *Experiment) IsEmpty() bool {
	if e.PI == "" {
		return true
	}
	return false
}

func (e *Experiment) NewAnalysisID() {
	e.AnalysisID = "1234567890"
}

func (e *Experiment) PrintRawSamplePath() string {
	return fmt.Sprintf("%s/%s/%s", e.SamplePath, e.PI, e.Name)
}

func (e *Experiment) PrintAnalysisPath() string {
	return fmt.Sprintf("%s/%s/%s/%s", e.AnalysisPath, e.PI, e.Name, e.AnalysisID)
}

func (e *Experiment) PrintWorkingDirectory() string {
	return fmt.Sprintf("%s", e.WorkDir)
}

func (e *Experiment) InitializePaths() {
	for i, _ := range e.Samples {
		e.Samples[i].SamplePath = e.PrintRawSamplePath()
	}
}

func (j *Job) MaxCPUUsage() int64 {
	var maxCPU = int64(0)

	for _, cmd := range j.Commands {
		if cmd.Preamble.CPUs > maxCPU {
			maxCPU = cmd.Preamble.CPUs
		}
	}
	return maxCPU
}

func (c *Command) CommandName() string {
	return c.CommandParams.Command
}
func (c *Command) SubCommandName() string {
	return c.CommandParams.Subcommand
}

func (p *SlurmPreamble) NotificationType() string {
	var notifications = ""

	// Add BEGIN tag
	if p.EmailBegin {
		notifications = "BEGIN"
	}

	// Add END tag
	if p.EmailEnd {
		if notifications == "" {
			notifications = "END"
		} else {
			notifications += ",END"
		}
	}

	// Add FAIL tag
	if p.EmailFail {
		if notifications == "" {
			notifications = "FAIL"
		} else {
			notifications += ",FAIL"
		}
	}

	// Set NONE tag if no options set.
	if notifications == "" {
		notifications = "NONE"
	}
	return notifications
}

func (s *Sample) IsPairedEnd() bool {
	if s.ReverseReadFile == "" {
		return false
	}
	return true
}

func (s *Sample) AddToOutputPath(dir string) {
	s.OutputPath += fmt.Sprintf("/%s", dir)
}

func (s *Sample) DumpReadFiles() string {
	readFileString := fmt.Sprintf("%s/%s", s.SamplePath, s.ForwardReadFile)
	if s.ReverseReadFile != "" {
		readFileString += fmt.Sprintf(" %s/%s", s.SamplePath, s.ReverseReadFile)
	}

	return readFileString
}

func (s *Sample) DumpForwardReadFile(noext bool) string {
	if noext {
		// Drop the file extension from the name
		chunks := strings.Split(s.ForwardReadFile, ".fastq")
		return chunks[0]
	}
	return s.ForwardReadFile
}

func (s *Sample) DumpForwardReadFileWithPath() string {
	readFileString := fmt.Sprintf("%s/%s", s.SamplePath, s.ForwardReadFile)
	return readFileString
}

func (s *Sample) DumpReverseReadFile(noext bool) string {
	if noext {
		// Drop the file extension from the name
		chunks := strings.Split(s.ReverseReadFile, ".fastq")
		return chunks[0]
	}
	return s.ReverseReadFile
}

func (s *Sample) DumpReverseReadFileWithPath() string {
	readFileString := fmt.Sprintf("%s/%s", s.SamplePath, s.ReverseReadFile)
	return readFileString
}
