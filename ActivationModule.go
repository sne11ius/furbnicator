package main

import (
	"bufio"
	"fmt"
	"github.com/spf13/viper"
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

func (a ActivationModule) Name() string {
	return "Activation"
}

func (a ActivationModule) Description() string {
	return "(De-)activates optional modules"
}

func (a ActivationModule) CanBeDisabled() bool {
	return false
}

func (a ActivationModule) IsModuleActive(module Module) bool {
	if !module.CanBeDisabled() {
		return true
	}
	return a.moduleActivations[module.Name()]
}

func (a ActivationModule) UpdateSettings() {
	for i := range modules {
		module := modules[i]
		if module.CanBeDisabled() {
			configKey := module.Name()
			if !viper.IsSet(configKey) {
				// fmt.Printf("Activation not configured for %s\n", module.Name())
				fmt.Printf("Activate module %s (Y/n)? ", module.Name())
				scanner := bufio.NewScanner(os.Stdin)
				scanner.Scan()
				configInput := scanner.Text()
				isActive := false
				switch {
				case configInput == "":
					fallthrough
				case configInput == "Y":
					fallthrough
				case configInput == "y":
					fallthrough
				case configInput == "yes":
					fallthrough
				case configInput == "true":
					isActive = true
				}
				viper.Set(configKey, isActive)
				// fmt.Printf("Module %s activated: %t\n", module.Name(), isActive)
			}
			a.moduleActivations[module.Name()] = viper.GetBool(configKey)
		}
	}
}
