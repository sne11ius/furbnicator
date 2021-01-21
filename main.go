package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/bep/debounce"
	"github.com/gdamore/tcell/v2"
	"github.com/mitchellh/go-homedir"
	"github.com/rivo/tview"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type action interface {
	GetLabel() string
	Run() string
}

var activationModule = NewActivationModule()
var modules = []Module{
	activationModule,
	NewJenkinsModule(),
	NewBitbucketModule(),
	NewTimestampModule(),
}

var activeModules []Module

func main() {
	updateModuleSettings()
	for _, module := range modules {
		if activationModule.IsModuleActive(module) {
			activeModules = append(activeModules, module)
		}
	}

	doUpdate := flag.Bool("u", false, "Update data and exit. Consider running with this option as cron task")
	feelingLucky := flag.Bool("l", false, "Feeling lucky? Add this flag to skip the run prompt if the other args filter the actions down to just a single one.")
	flag.Parse()

	if *doUpdate == true {
		updateDataForActiveModules()
		return
	} else {
		readCacheDataForActiveModules()
	}

	params := remove(os.Args[1:], "-l")
	tags := TagsFromStrings(params)
	var actions []action
	for _, module := range activeModules {
		actions = append(actions, module.CreateActions(tags)...)
	}
	doRun := func(a action) {
		message := a.Run()
		fmt.Println(message)
		os.Exit(0)
	}
	if len(actions) == 1 {
		action := actions[0]
		if *feelingLucky {
			doRun(action)
		} else {
			fmt.Printf("Run %s? (Y/n) ", action.GetLabel())
			if readBool() {
				doRun(action)
			}
		}
	}

	app := tview.NewApplication()
	initialText := strings.Join(params, " ")
	list := tview.NewList().
		ShowSecondaryText(false).
		SetWrapAround(false).
		SetSelectedTextColor(tcell.ColorBlack).
		SetSelectedFunc(func(index int, _ string, _ string, _ rune) {
			app.Stop()
			action := actions[index]
			message := action.Run()
			fmt.Println("ﴪ >>> " + message)
			os.Exit(0)
		})
	for _, action := range actions {
		list.AddItem(action.GetLabel(), "", 0, nil)
	}

	var searchArgs []string
	doSearch := func() {
		tags := TagsFromStrings(searchArgs)
		actions = []action{}
		for _, module := range activeModules {
			actions = append(actions, module.CreateActions(tags)...)
		}
		list.Clear()
		for _, action := range actions {
			list.AddItem(action.GetLabel(), "", 0, nil)
		}
		app.Draw()
	}
	debounced := debounce.New(200 * time.Millisecond)
	inputField := tview.NewInputField().
		SetLabel("furbnicator > ﴪ >>> ").
		SetFieldWidth(0).
		SetText(initialText).
		SetChangedFunc(func(text string) {
			searchArgs = strings.Split(text, " ")
			debounced(doSearch)
		}).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEsc {
				app.Stop()
			}
		})
	inputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyDown && list.GetItemCount() != 0 {
			app.SetFocus(list)
			return nil
		} else {
			return event
		}
	})
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyUp && list.GetCurrentItem() == 0 {
			app.SetFocus(inputField)
			return nil
		} else {
			return event
		}
	})
	var inputHasFocus = true
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTAB {
			if inputHasFocus {
				app.SetFocus(list)
			} else {
				app.SetFocus(inputField)
			}
			inputHasFocus = !inputHasFocus
			return nil
		}
		return event
	})
	flex := tview.NewFlex().
		SetFullScreen(true).
		SetDirection(tview.FlexRow).
		AddItem(inputField, 1, 0, true).
		AddItem(list, 0, 1, false)
	if err := app.SetRoot(flex, true).SetFocus(inputField).Run(); err != nil {
		panic(err)
	}
}

func readBool() bool {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := scanner.Text()
	switch {
	case input == "":
		fallthrough
	case input == "Y":
		fallthrough
	case input == "y":
		fallthrough
	case input == "yes":
		fallthrough
	case input == "true":
		return true
	}
	return false
}

func updateModuleSettings() {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal("Error: cannot home directory")
	}
	configDir := filepath.Join(home, ".config", "furbnicator")
	viper.SetConfigName("furbnicator")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		fullConfigName := filepath.Join(configDir, "furbnicator.yaml")
		localConfigName := "./furbnicator.yaml"
		log.Fatalf("No/invalid config found at %s or %s\n: %v", localConfigName, fullConfigName, err)
	}
	for i := range modules {
		module := modules[i]
		if activationModule.IsModuleActive(module) {
			module.UpdateSettings()
		}
	}
}

func updateDataForActiveModules() {
	for _, module := range activeModules {
		if module.NeedsExternalData() {
			fmt.Printf("Updating %s\n", module.Name())
			module.UpdateExternalData()
			filenameForModule := LocateConfigFile(module)
			file, err := os.Create(filenameForModule)
			if err != nil {
				log.Fatalf("Cannot open configuration file %s: %s", filenameForModule, err)
			}
			module.WriteExternalData(file)
			err = file.Close()
			if err != nil {
				log.Fatalf("Cannot close file %s: %s", filenameForModule, err)
			}
		}
	}
}

func readCacheDataForActiveModules() {
	for _, module := range activeModules {
		if module.NeedsExternalData() {
			filenameForModule := LocateConfigFile(module)
			data, err := ioutil.ReadFile(filenameForModule)
			if err != nil {
				log.Fatalf("Cannot read configuration from %s. Consider running with `-u` parameter first.", filenameForModule)
			}
			module.ReadExternalData(data)
		}
	}
}

func LocateConfigFile(module Module) string {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal("Error: cannot home directory")
	}
	return filepath.Join(home, ".config", "furbnicator", module.Name()+".json")
}
