package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/tmc/pe/internal/pemod"
)

func readRuntimePolicyFile() (*pemod.File, bool, error) {
	data, err := os.ReadFile("pe.mod")
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("reading pe.mod policy: %w", err)
	}
	file, err := pemod.Parse(strings.NewReader(string(data)))
	if err != nil {
		return nil, false, fmt.Errorf("parsing pe.mod policy: %w", err)
	}
	return file, true, nil
}

func enforceRuntimeToolPolicy(tool string) error {
	file, ok, err := readRuntimePolicyFile()
	if err != nil || !ok || file.Capability == nil {
		return err
	}
	for _, deny := range file.Capability.Tools.Deny {
		if deny == tool {
			return fmt.Errorf("tool %s is denied by pe.mod", tool)
		}
	}
	return nil
}
