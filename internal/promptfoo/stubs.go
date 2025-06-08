package promptfoo

// Temporary stubs to allow compilation while we migrate to plugin architecture

// Results represents the results structure
type Results struct {
	Table Table `json:"table"`
}

// Table represents a results table
type Table struct {
	Head Head      `json:"head"`
	Body []BodyRow `json:"body"`
}

// Head represents table headers
type Head struct {
	Prompts []PromptInfo `json:"prompts"`
	Vars    []string     `json:"vars"`
}

// PromptInfo represents prompt information
type PromptInfo struct {
	Raw      string `json:"raw"`
	Label    string `json:"label"`
	Provider string `json:"provider"`
}

// Types PEPrompt, PETestCase, and PETestMetadata are defined in export.go
