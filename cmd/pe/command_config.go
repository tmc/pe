package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/config"
)

const commandTimeoutAnnotation = "pe.command.timeout"

func applyRootConfigDefaults(root *cobra.Command) error {
	cfg, err := loadConfigForCommand()
	if err != nil {
		return err
	}
	applyCommandConfigDefaults(root, cfg)
	return nil
}

func applyCommandConfigDefaults(cmd *cobra.Command, cfg *config.Config) {
	if cmd == nil || cfg == nil {
		return
	}
	if commandCfg, ok := cfg.Command(cmd.Name()); ok {
		applyCommandConfig(cmd, commandCfg)
	}
	for _, child := range cmd.Commands() {
		applyCommandConfigDefaults(child, cfg)
	}
}

func applyCommandConfig(cmd *cobra.Command, cfg config.CommandConfig) {
	if cfg.Provider != "" {
		setFlagDefault(cmd, "provider", cfg.Provider)
	}
	if cfg.Timeout > 0 {
		if cmd.Annotations == nil {
			cmd.Annotations = make(map[string]string)
		}
		cmd.Annotations[commandTimeoutAnnotation] = cfg.Timeout.String()
		setFlagDefault(cmd, "timeout", cfg.Timeout.String())
	}
	for key, value := range cfg.Options {
		setFlagDefault(cmd, key, configValueString(value))
	}
}

func setFlagDefault(cmd *cobra.Command, name, value string) {
	flag := cmd.Flags().Lookup(name)
	if flag == nil {
		return
	}
	flag.DefValue = value
	_ = flag.Value.Set(value)
}

func configValueString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case time.Duration:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}
