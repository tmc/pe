package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/config"
	"sigs.k8s.io/yaml"
)

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Inspect PE configuration",
	}
	cmd.AddCommand(configGetCmd())
	cmd.AddCommand(configListCmd())
	cmd.AddCommand(configValidateCmd())
	return cmd
}

func configGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get key",
		Short: "Print one configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfigForCommand()
			if err != nil {
				return err
			}
			value, ok, err := configValue(cfg, args[0])
			if err != nil {
				return err
			}
			if !ok {
				return fmt.Errorf("config key %q not found", args[0])
			}
			fmt.Fprintln(cmd.OutOrStdout(), value)
			return nil
		},
	}
}

func configListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Print merged configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfigForCommand()
			if err != nil {
				return err
			}
			out, err := yaml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("marshal config: %w", err)
			}
			fmt.Fprint(cmd.OutOrStdout(), string(out))
			return nil
		},
	}
}

func configValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate merged configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := config.NewManager()
			if err != nil {
				return err
			}
			if err := manager.Validate(); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "ok")
			return nil
		},
	}
}

func loadConfigForCommand() (*config.Config, error) {
	manager, err := config.NewManager()
	if err != nil {
		return nil, err
	}
	return manager.Get(), nil
}

func configValue(cfg *config.Config, key string) (string, bool, error) {
	data, err := json.Marshal(cfg)
	if err != nil {
		return "", false, fmt.Errorf("marshal config: %w", err)
	}
	var root map[string]interface{}
	if err := json.Unmarshal(data, &root); err != nil {
		return "", false, fmt.Errorf("unmarshal config: %w", err)
	}
	var value interface{} = root
	for _, part := range strings.Split(key, ".") {
		obj, ok := value.(map[string]interface{})
		if !ok {
			return "", false, nil
		}
		value, ok = obj[part]
		if !ok {
			return "", false, nil
		}
	}
	switch v := value.(type) {
	case string:
		return v, true, nil
	case bool:
		return fmt.Sprintf("%t", v), true, nil
	case float64:
		return fmt.Sprintf("%g", v), true, nil
	default:
		out, err := json.Marshal(v)
		if err != nil {
			return "", false, fmt.Errorf("marshal value: %w", err)
		}
		return string(out), true, nil
	}
}
