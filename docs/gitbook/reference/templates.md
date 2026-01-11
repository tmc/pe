# Template Syntax

PE uses the powerful [Go text/template](https://pkg.go.dev/text/template) engine.

## Basic Syntax

Use double curly braces `{{ }}` with a dot `.` to access variables.

```text
Hello, {{.name}}!
```

**Correct**: `{{.variable}}`
**Incorrect**: `{{variable}}` (missing dot)

## Common Actions

### Conditionals
```text
{{if .is_formal}}
Dear {{.name}},
{{else}}
Hi {{.name}},
{{end}}
```

### Loops (Ranges)
```text
{{range .items}}
- {{.}}
{{end}}
```

### Functions
PE provides standard functions plus extras:
```text
{{.content | printf "%.100s"}}  # Truncate
{{.header | upper}}             # Uppercase
```

## JSON Input
When passing JSON variables (e.g. via `--var-json`), nested objects are accessible via dot notation:

```text
User: {{.user.name}}
Role: {{.user.role}}
```
