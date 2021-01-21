package main

import (
	"os"
	"strconv"
	"time"
)

type TimestampModule struct {
}

func NewTimestampModule() *TimestampModule {
	return new(TimestampModule)
}

func (t TimestampModule) Name() string {
	return "Timestamp"
}

func (t TimestampModule) Description() string {
	return "Displays timestamps"
}

func (t TimestampModule) CanBeDisabled() bool {
	return true
}

func (t TimestampModule) UpdateSettings() {
	// this intentionally empty
}

func (t TimestampModule) NeedsExternalData() bool {
	return false
}

func (t TimestampModule) UpdateExternalData() {
	// this intentionally empty
}

func (t TimestampModule) WriteExternalData(_ *os.File) {
	// this intentionally empty
}

type TimestampAction struct {
	tstype string
	value  string
}

func (t TimestampAction) GetLabel() string {
	return "[timestamp[] " + t.tstype + " " + t.value
}

func (t TimestampAction) Run() string {
	return "Current " + t.tstype + " timestamp: " + t.value
}

func (t TimestampModule) CreateActions(tags []Tag) []action {
	var actions []action
	now := time.Now()
	tsUnix := now.Unix()
	tsJava := now.UnixNano() / 1e6
	strs := []string{"timestamp", "ts", "unix"}
	if DoMatch(strs, tags) {
		actions = append(actions, TimestampAction{
			tstype: "UNIX",
			value:  strconv.FormatInt(tsUnix, 10),
		})
	}
	strs = []string{"timestamp", "ts", "java"}
	if DoMatch(strs, tags) {
		actions = append(actions, TimestampAction{
			tstype: "JAVA",
			value:  strconv.FormatInt(tsJava, 10),
		})
	}
	return actions
}

func (t TimestampModule) ReadExternalData(_ []byte) {
	// this intentionally empty
}
