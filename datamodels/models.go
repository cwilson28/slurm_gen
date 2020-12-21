package datamodels

type SlurmParams struct {
	Batch             bool
	SamplesFile       string
	SampleFilePrefix  string
	JobName           string
	Partition         string
	NotificationBegin bool
	NotificationEnd   bool
	NotificationFail  bool
	NotificationEmail string
	Tasks             int64
	CPUs              int64
	Memory            int64
	Time              string
}

type SlurmPreamble struct {
	Partition         string
	NotificationBegin bool
	NotificationEnd   bool
	NotificationFail  bool
	NotificationEmail string
}

type CommandPreamble struct {
	JobName string
	Tasks   int64
	CPUs    int64
	Memory  int64
	Time    string
}

type CommandParams struct {
	SingularityPath  string
	SingularityImage string
	WorkDir          string
	Volume           string
	Command          string
	CommandOptions   []string
	CommandArgs      []string
}

type Job struct {
	Batch         bool
	BatchCommands []string
	SamplesFile   string
	SlurmPreamble SlurmPreamble
	Commands      []Command
}

type Command struct {
	CommandName   string
	Preamble      CommandPreamble
	CommandParams CommandParams
}

type BatchParams struct {
	SamplePrefix string
	ForwardReads []string
	ReverseReads []string
}

type Sample struct {
	Prefix          string
	ForwardReadFile string
	ReverseReadFile string
}

func (p *SlurmPreamble) NotificationType() string {
	var notifications = ""

	// Add BEGIN tag
	if p.NotificationBegin {
		notifications = "BEGIN"
	}

	// Add END tag
	if p.NotificationEnd {
		if notifications == "" {
			notifications = "END"
		} else {
			notifications += ",END"
		}
	}

	// Add FAIL tag
	if p.NotificationFail {
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
