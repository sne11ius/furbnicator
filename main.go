package main

//import (
//	"github.com/gdamore/tcell/v2"
//	"github.com/rivo/tview"
//)

import (
	"flag"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"log"
	"path/filepath"
)

var activationModule = NewActivationModule()
var modules = []Module{
	activationModule,
	NewBitbucketModule(),
	NewJenkinsModule(),
}

func main() {
	doUpdate := flag.Bool("u", false, "Update data and exit. Consider running with this option as cron task")
	flag.Parse()

	updateModuleSettings()

	if *doUpdate == true {
		updateAllModuleData()
		return
	}

	// for _, module := range modules {
	// 	fmt.Printf("%s: %s\n", module.Name(), module.Description())
	// }

	// app := tview.NewApplication()
	// inputField := tview.NewInputField().
	// 	SetLabel("Enter a number: ").
	// 	SetFieldWidth(10).
	// 	SetAcceptanceFunc(tview.InputFieldInteger).
	// 	SetDoneFunc(func(key tcell.Key) {
	// 		app.Stop()
	// 	})
	// if err := app.SetRoot(inputField, true).SetFocus(inputField).Run(); err != nil {
	//	panic(err)
	// }
}

func updateModuleSettings() {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal("Error: cannot home directory")
	}
	// fmt.Printf("Home dir: %s\n", home)
	configDir := filepath.Join(home, ".config", "furbnicator")
	// fmt.Printf("Config dir: %s\n", configDir)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		fullConfigName := filepath.Join(configDir, "config.yaml")
		fmt.Printf("No config found at %s\n", fullConfigName)
	}
	for i := range modules {
		module := modules[i]
		if activationModule.IsModuleActive(module) {
			module.UpdateSettings()
		} else {
			fmt.Printf("Module not active: %s\n", module.Name())
		}
	}
	/*
		err = os.MkdirAll(configDir, os.ModePerm)
		if err != nil {
			log.Fatalf("Cannot create config dir %s\n", configDir)
		}
		configFile := filepath.Join(configDir, "config.txt")
		if _, err := os.Stat(configFile); err != nil {
			file, err := os.Create(configFile)
			if err != nil {
				log.Fatalf("Cannot create config file %s", configFile)
			}
			settings = readSettings(file)
			defer file.Close()
		}
			for i := range modules {
				module := modules[i]
			}
	*/
}

func updateAllModuleData() {
	for _, module := range modules {
		fmt.Printf("Updating %s\n", module.Name())
	}
}
