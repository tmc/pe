package promptfoo

// Results is the table-shaped result format used by promptfoo-compatible output.
type Results struct {
	Table Table `json:"table"`
}

// Table holds promptfoo-compatible result headers and rows.
type Table struct {
	Head Head      `json:"head"`
	Body []BodyRow `json:"body"`
}

// Head holds prompt and variable column metadata.
type Head struct {
	Prompts []PromptInfo `json:"prompts"`
	Vars    []string     `json:"vars"`
}

// PromptInfo describes a prompt column in promptfoo-compatible output.
type PromptInfo struct {
	Raw      string `json:"raw"`
	Label    string `json:"label"`
	Provider string `json:"provider"`
}

// Types PEPrompt, PETestCase, and PETestMetadata are defined in export.go.
