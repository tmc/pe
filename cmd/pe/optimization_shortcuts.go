package main

import "github.com/spf13/cobra"

func textgradCmd() *cobra.Command {
	return optimizeMethodCmd("textgrad", "Optimize prompts with textual gradients")
}

func pe2Cmd() *cobra.Command {
	return optimizeMethodCmd("pe2", "Optimize prompts with PE2 meta-prompting")
}

func optimizeMethodCmd(method, short string) *cobra.Command {
	cmd := optimizeCmd()
	cmd.Use = method + " [prompt-file]"
	cmd.Short = short
	cmd.Long = ""
	cmd.Example = ""
	_ = cmd.Flags().Set("method", method)
	_ = cmd.Flags().MarkHidden("method")
	return cmd
}
