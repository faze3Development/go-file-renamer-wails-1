package config

// Config holds all the configuration for the file watcher and its components.
type Config struct {
	WatchPaths     []string          `json:"WatchPaths"`
	Recursive      bool              `json:"Recursive"`
	DryRun         bool              `json:"DryRun"`
	NamePattern    string            `json:"NamePattern"`
	RandomLength   int               `json:"RandomLength"`
	Settle         int               `json:"Settle"`
	SettleTimeout  int               `json:"SettleTimeout"`
	Retries        int               `json:"Retries"`
	NoInitialScan  bool              `json:"NoInitialScan"`
	NamerID        string            `json:"NamerID"`
	ActionID       string            `json:"ActionID"`
	TemplateString string            `json:"TemplateString"`
	DateTimeFormat string            `json:"DateTimeFormat"`
	ActionConfig   map[string]string `json:"ActionConfig"`
}
