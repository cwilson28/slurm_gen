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
	SlurmPreamble     SlurmPreamble
	SGEPreamble       SGEPreamble
	Commands          []Command
	ExperimentDetails Experiment
	SamplesFile       string
}

type Command struct {
	Batch         bool
	SamplesFile   string
	CommandName   string
	Preamble      CommandPreamble
	CommandParams CommandParams
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
	Samples      []Sample
}

type Sample struct {
	SamplePath      string
	OutputPath      string
	Prefix          string
	ForwardReadFile string
	ReverseReadFile string
}

func (j *Job) IsPipeline() bool {
	if len(j.Commands) > 1 {
		return true
	}
	return false
}

/* --- Class functions --- */
func NewExperiment() Experiment {
	return Experiment{
		SamplePath:   "compbio/data",
		AnalysisPath: "compbio/analysis",
	}
}

func (e *Experiment) NewAnalysisID() {
	e.AnalysisID = "1234567890"
}

func (e *Experiment) DumpSamplePath() string {
	return fmt.Sprintf("%s/%s/%s", e.SamplePath, e.PI, e.Name)
}

func (e *Experiment) DumpAnalysisPath() string {
	return fmt.Sprintf("%s/%s/%s/%s", e.AnalysisPath, e.PI, e.Name, e.AnalysisID)
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
