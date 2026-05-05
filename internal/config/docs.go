package config

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"
)

// WriteDocumentation writes a Markdown reference for the configuration schema.
func WriteDocumentation(w io.Writer) error {
	if _, err := fmt.Fprintln(w, "# PE Configuration"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	return writeConfigFields(w, reflect.TypeOf(Config{}), "")
}

func writeConfigFields(w io.Writer, typ reflect.Type, prefix string) error {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			continue
		}
		name := configFieldName(field)
		if name == "" || name == "-" {
			continue
		}
		key := name
		if prefix != "" {
			key = prefix + "." + name
		}
		ft := field.Type
		if ft == reflect.TypeOf(time.Duration(0)) {
			if err := writeConfigField(w, key, field, "duration"); err != nil {
				return err
			}
			continue
		}
		if ft.Kind() == reflect.Struct {
			if err := writeConfigFields(w, ft, key); err != nil {
				return err
			}
			continue
		}
		if err := writeConfigField(w, key, field, ft.String()); err != nil {
			return err
		}
	}
	return nil
}

func writeConfigField(w io.Writer, key string, field reflect.StructField, typ string) error {
	if _, err := fmt.Fprintf(w, "## `%s`\n\n", key); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "- Type: `%s`\n", typ); err != nil {
		return err
	}
	if def := field.Tag.Get("default"); def != "" {
		if _, err := fmt.Fprintf(w, "- Default: `%s`\n", def); err != nil {
			return err
		}
	}
	if env := field.Tag.Get("env"); env != "" {
		if _, err := fmt.Fprintf(w, "- Environment: `%s`\n", env); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w)
	return err
}

func configFieldName(field reflect.StructField) string {
	if tag := field.Tag.Get("yaml"); tag != "" {
		return strings.Split(tag, ",")[0]
	}
	if tag := field.Tag.Get("json"); tag != "" {
		return strings.Split(tag, ",")[0]
	}
	return strings.ToLower(field.Name)
}
