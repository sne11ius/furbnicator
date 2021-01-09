package main

import (
	"github.com/spf13/viper"
	"log"
	"os"
)

type ActivationModule struct {
	moduleActivations map[string]bool
}

func NewActivationModule() *ActivationModule {
	a := new(ActivationModule)
	a.moduleActivations = make(map[string]bool)
	return a
}

func (a *ActivationModule) Name() string {
	return "Activation"
}

func (a *ActivationModule) Description() string {
	return "(De-)activates optional modules"
}

func (a *ActivationModule) CanBeDisabled() bool {
	return false
}

func (a *ActivationModule) IsModuleActive(module Module) bool {
	if !module.CanBeDisabled() {
		return true
	}
	return a.moduleActivations[module.Name()]
}

func (a *ActivationModule) UpdateSettings() {
	for i := range modules {
		module := modules[i]
		if module.CanBeDisabled() {
			configKey := a.Name() + "." + module.Name()
			if !viper.IsSet(configKey) {
				log.Fatalf("Missing configuration key `%s` (true|false)", configKey)
			}
			a.moduleActivations[module.Name()] = viper.GetBool(configKey)
		}
	}
}

func (a *ActivationModule) NeedsExternalData() bool {
	return false
}

func (a *ActivationModule) UpdateExternalData() {
	// this intentionally empty
}

func (a *ActivationModule) WriteExternalData(file *os.File) {
	// this intentionally empty
}
