package datamodels

type SlurmParams struct {
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

type ToolParams struct {
	SingularityPath  string
	SingularityImage string
	WorkDir          string
	Volume           string
	Command          string
	CommandOptions   []string
	CommandArgs      []string
}

type ScriptDef struct {
	SlurmParams SlurmParams
	ToolParams  ToolParams
}

func (bp *SlurmParams) NotificationType() string {
	var notifications = ""

	// Add BEGIN tag
	if bp.NotificationBegin {
		notifications = "BEGIN"
	}

	// Add END tag
	if bp.NotificationEnd {
		if notifications == "" {
			notifications = "END"
		} else {
			notifications += ",END"
		}
	}

	// Add FAIL tag
	if bp.NotificationFail {
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
