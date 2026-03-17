package promptfoo

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ApplyVars applies test variables to a prompt string.
// It supports both {{name}} and {{.name}} forms.
func ApplyVars(prompt string, vars map[string]interface{}) string {
	result := prompt
	for key, value := range vars {
		strValue := formatPromptValue(value)
		result = strings.ReplaceAll(result, "{{"+key+"}}", strValue)
		result = strings.ReplaceAll(result, "{{."+key+"}}", strValue)
	}
	return result
}

func formatPromptValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%g", v)
	case int:
		return fmt.Sprintf("%d", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		jsonValue, err := json.Marshal(v)
		if err == nil {
			return string(jsonValue)
		}
		return fmt.Sprintf("%v", v)
	}
}
