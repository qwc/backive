package core

// Backup contains all necessary information for executing a configured backup.
type Backup struct {
	Name         string `mapstructure:",omitempty"`
	TargetDevice string `mapstructure:"targetDevice"`
	TargetDir    string `mapstructure:"targetDir"`
	SourceDir    string `mapstructure:"sourceDir"`
	ScriptPath   string `mapstructure:"scriptPath"`
	Frequency    int    `mapstructure:"frequency"`
	ExeUser      string `mapstructure:"user,omitempty"`
}
