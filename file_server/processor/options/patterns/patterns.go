package patterns

// PatternInfo holds the data for a predefined filename matching pattern.
type PatternInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Regex       string `json:"regex"`
}

// GetPatternInfo returns a slice of all predefined patterns.
func GetPatternInfo() []PatternInfo {
	return []PatternInfo{
		{
			ID:          "default_untitled",
			Name:        "Default Screenshots/Untitled",
			Description: "Matches common names for screenshots and untitled files.",
			Regex:       `(?i)^(?:untitled|screenshot(?:[\s_-]\d{4}-\d{2}-\d{2})?|untitiled)(?:[\s_-]*(?:\(\d+\)|\d+))?$`,
		},
		{
			ID:          "any_file",
			Name:        "Any File",
			Description: "Matches any file name.",
			Regex:       `.*`,
		},
	}
}
