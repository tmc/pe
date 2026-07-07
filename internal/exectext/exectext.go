// Package exectext parses and renders PE executable text files.
package exectext

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// File is an executable text artifact.
type File struct {
	Shebang string
	Meta    Metadata
	Body    string
}

// Metadata is optional front matter for executable text.
type Metadata struct {
	Kind      string                   `yaml:"kind" json:"kind"`
	Run       string                   `yaml:"run" json:"run"`
	Name      string                   `yaml:"name" json:"name"`
	Metadata  map[string]interface{}   `yaml:"metadata" json:"metadata"`
	Inputs    map[string]Input         `yaml:"inputs" json:"inputs"`
	Outputs   map[string]Input         `yaml:"outputs" json:"outputs"`
	Budget    map[string]interface{}   `yaml:"budget" json:"budget"`
	Imports   map[string]string        `yaml:"imports" json:"imports"`
	Safety    map[string]AllowDenyList `yaml:"safety" json:"safety"`
	Placement map[string]interface{}   `yaml:"placement" json:"placement"`
}

// Input describes a template input or output.
type Input struct {
	Type        string      `yaml:"type" json:"type"`
	Description string      `yaml:"description" json:"description"`
	Default     interface{} `yaml:"default" json:"default"`
	Required    *bool       `yaml:"required" json:"required"`
}

// AllowDenyList is a simple policy dimension.
type AllowDenyList struct {
	Allow []string `yaml:"allow" json:"allow"`
	Deny  []string `yaml:"deny" json:"deny"`
}

// Parse reads an executable text file.
func Parse(r io.Reader) (*File, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading executable text: %w", err)
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	f := &File{}

	if strings.HasPrefix(text, "#!") {
		line, rest, ok := strings.Cut(text, "\n")
		f.Shebang = strings.TrimSpace(line)
		if ok {
			text = rest
		} else {
			text = ""
		}
	}

	trimmed := strings.TrimPrefix(text, "\ufeff")
	if strings.HasPrefix(trimmed, "---\n") {
		meta, rest, ok := strings.Cut(trimmed[len("---\n"):], "\n---\n")
		if !ok {
			return nil, fmt.Errorf("front matter missing closing ---")
		}
		if err := yaml.Unmarshal([]byte(meta), &f.Meta); err != nil {
			return nil, fmt.Errorf("parsing front matter: %w", err)
		}
		text = rest
	}

	f.Body = strings.TrimLeft(text, "\n")
	return f, nil
}

// Render validates declared inputs and renders the text body.
func (f *File) Render(vars map[string]string) (string, error) {
	return f.RenderWithImports(vars, nil)
}

// RenderWithImports renders the text body with named imported text components.
func (f *File) RenderWithImports(vars map[string]string, imports map[string]string) (string, error) {
	vals := make(map[string]interface{})
	for k, v := range vars {
		vals[k] = v
	}
	for name, in := range f.Meta.Inputs {
		if _, ok := vals[name]; ok {
			continue
		}
		if in.Default != nil {
			vals[name] = in.Default
			continue
		}
		if in.Required == nil || *in.Required {
			return "", fmt.Errorf("missing input %s", name)
		}
	}

	funcs := template.FuncMap{
		"import": func(name string) (string, error) {
			body, ok := imports[name]
			if !ok {
				return "", fmt.Errorf("unknown import %s", name)
			}
			child, err := Parse(strings.NewReader(body))
			if err != nil {
				return "", err
			}
			if err := child.Validate(); err != nil {
				return "", err
			}
			if err := CheckComposition(f, child); err != nil {
				return "", fmt.Errorf("import %s: %w", name, err)
			}
			return child.RenderWithImports(vars, nil)
		},
	}

	tmpl, err := template.New(nameOrDefault(f.Meta.Name)).Funcs(funcs).Option("missingkey=error").Parse(f.Body)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vals); err != nil {
		return "", fmt.Errorf("rendering template: %w", err)
	}
	return buf.String(), nil
}

// CheckComposition reports whether child loosens parent's declared safety
// policy. Composition is conservative: a child allow entry for a value the
// parent denies is an error, and when the parent declares an allow list for
// a dimension, child allow entries must stay within it. A child may always
// tighten policy with additional deny entries.
func CheckComposition(parent, child *File) error {
	for dim, childList := range child.Meta.Safety {
		parentList, ok := parent.Meta.Safety[dim]
		if !ok {
			continue
		}
		denied := make(map[string]bool, len(parentList.Deny))
		for _, v := range parentList.Deny {
			denied[v] = true
		}
		for _, v := range childList.Allow {
			if denied[v] {
				return fmt.Errorf("safety %s allows %s denied by parent", dim, v)
			}
		}
		if len(parentList.Allow) > 0 {
			allowed := make(map[string]bool, len(parentList.Allow))
			for _, v := range parentList.Allow {
				allowed[v] = true
			}
			for _, v := range childList.Allow {
				if !allowed[v] {
					return fmt.Errorf("safety %s allows %s outside parent allow list", dim, v)
				}
			}
		}
	}
	return nil
}

// Validate checks the static contract without running tools or providers.
func (f *File) Validate() error {
	if f.Meta.Kind != "" && f.Meta.Kind != "pe.text.v1" && f.Meta.Kind != "pe.workflow.v1" {
		return fmt.Errorf("unsupported executable text kind %s", f.Meta.Kind)
	}
	if f.Meta.Run != "" && f.Meta.Run != "pe run-text" {
		return fmt.Errorf("unsupported runner %s", f.Meta.Run)
	}
	if f.Shebang != "" && !strings.Contains(f.Shebang, "pe run-text") {
		return fmt.Errorf("unsupported shebang %s", f.Shebang)
	}
	return nil
}

func nameOrDefault(name string) string {
	if name == "" {
		return "text"
	}
	return name
}
