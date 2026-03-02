package bulk_file_processing

import (
	"regexp"
	"strings"
)

var (
	backendUnsafeSequentialChars = regexp.MustCompile(`[^a-zA-Z0-9-_ ]+`)
	backendSequentialWhitespace  = regexp.MustCompile(`\s+`)
)

const backendSequentialMaxLength = 120

func sanitizeSequentialBaseName(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	sanitized := backendUnsafeSequentialChars.ReplaceAllString(trimmed, " ")
	sanitized = backendSequentialWhitespace.ReplaceAllString(sanitized, " ")
	sanitized = strings.TrimSpace(sanitized)

	if len(sanitized) > backendSequentialMaxLength {
		sanitized = sanitized[:backendSequentialMaxLength]
	}

	return strings.ReplaceAll(sanitized, " ", "-")
}

func normalizeSequentialOptions(opts SequentialNamingOptions) SequentialNamingOptions {
	opts.BaseName = sanitizeSequentialBaseName(opts.BaseName)

	if opts.StartIndex < 0 {
		opts.StartIndex = 0
	}

	if opts.PadLength < 0 {
		opts.PadLength = 0
	}
	if opts.PadLength > 6 {
		opts.PadLength = 6
	}

	if !opts.Enabled {
		opts.KeepExtension = true
	} else if !opts.KeepExtension {
		opts.KeepExtension = false
	} else {
		opts.KeepExtension = true
	}

	return opts
}

func normalizeRenameOptions(opts RenameOperationOptions) RenameOperationOptions {
	opts.Sequential = normalizeSequentialOptions(opts.Sequential)
	opts.CustomDate = strings.TrimSpace(opts.CustomDate)
	return opts
}

func intFromInterface(value interface{}, defaultValue int) int {
	switch v := value.(type) {
	case int:
		return v
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	default:
		return defaultValue
	}
}

func renameOptionsFromInterface(value interface{}) RenameOperationOptions {
	switch v := value.(type) {
	case RenameOperationOptions:
		return normalizeRenameOptions(v)
	case *RenameOperationOptions:
		if v == nil {
			return normalizeRenameOptions(RenameOperationOptions{})
		}
		return normalizeRenameOptions(*v)
	case map[string]interface{}:
		opts := RenameOperationOptions{}
		if preserve, ok := v["preserveOriginalName"].(bool); ok {
			opts.PreserveOriginalName = preserve
		}
		if addTimestamp, ok := v["addTimestamp"].(bool); ok {
			opts.AddTimestamp = addTimestamp
		}
		if addRandom, ok := v["addRandomId"].(bool); ok {
			opts.AddRandomID = addRandom
		}
		if addCustomDate, ok := v["addCustomDate"].(bool); ok {
			opts.AddCustomDate = addCustomDate
		}
		if customDate, ok := v["customDate"].(string); ok {
			opts.CustomDate = customDate
		}
		if useRegex, ok := v["useRegexReplace"].(bool); ok {
			opts.UseRegexReplace = useRegex
		}
		if regexFind, ok := v["regexFind"].(string); ok {
			opts.RegexFind = regexFind
		}
		if regexReplace, ok := v["regexReplace"].(string); ok {
			opts.RegexReplace = regexReplace
		}
		if sequential, ok := v["sequentialNaming"].(map[string]interface{}); ok {
			seq := SequentialNamingOptions{KeepExtension: true}
			if enabled, ok := sequential["enabled"].(bool); ok {
				seq.Enabled = enabled
			}
			if baseName, ok := sequential["baseName"].(string); ok {
				seq.BaseName = baseName
			}
			if start, exists := sequential["startIndex"]; exists {
				seq.StartIndex = intFromInterface(start, seq.StartIndex)
			}
			if pad, exists := sequential["padLength"]; exists {
				seq.PadLength = intFromInterface(pad, seq.PadLength)
			}
			if keep, ok := sequential["keepExtension"].(bool); ok {
				seq.KeepExtension = keep
			}
			opts.Sequential = seq
		}
		return normalizeRenameOptions(opts)
	default:
		return normalizeRenameOptions(RenameOperationOptions{})
	}
}
