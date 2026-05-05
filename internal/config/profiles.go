package config

import "fmt"

// Profile is a named set of configuration overrides.
type Profile struct {
	Name     string                 `json:"name" yaml:"name"`
	Inherits string                 `json:"inherits,omitempty" yaml:"inherits,omitempty"`
	Values   map[string]interface{} `json:"values,omitempty" yaml:"values,omitempty"`
}

// ProfileSet holds named configuration profiles.
type ProfileSet struct {
	Profiles map[string]Profile `json:"profiles" yaml:"profiles"`
}

// Resolve returns the inherited values for name.
func (s ProfileSet) Resolve(name string) (map[string]interface{}, error) {
	return s.resolve(name, nil)
}

func (s ProfileSet) resolve(name string, stack map[string]bool) (map[string]interface{}, error) {
	profile, ok := s.Profiles[name]
	if !ok {
		return nil, fmt.Errorf("profile %q not found", name)
	}
	if stack == nil {
		stack = make(map[string]bool)
	}
	if stack[name] {
		return nil, fmt.Errorf("profile %q has inheritance cycle", name)
	}
	stack[name] = true

	values := make(map[string]interface{})
	if profile.Inherits != "" {
		parent, err := s.resolve(profile.Inherits, stack)
		if err != nil {
			return nil, err
		}
		for k, v := range parent {
			values[k] = v
		}
	}
	for k, v := range profile.Values {
		values[k] = v
	}
	delete(stack, name)
	return values, nil
}

// ApplyOverrides applies dot-notation overrides to the current configuration.
func (cm *ConfigManager) ApplyOverrides(overrides map[string]interface{}) error {
	for key, value := range overrides {
		if err := cm.Set(key, value); err != nil {
			return err
		}
	}
	return nil
}

// ApplyProfile resolves and applies a profile to the current configuration.
func (cm *ConfigManager) ApplyProfile(profiles ProfileSet, name string) error {
	overrides, err := profiles.Resolve(name)
	if err != nil {
		return err
	}
	return cm.ApplyOverrides(overrides)
}
