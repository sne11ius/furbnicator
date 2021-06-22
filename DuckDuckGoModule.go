package main

import (
	"log"
	"os"
)

type DuckDuckGoModule struct {
}

func NewDuckDuckGoModule() *DuckDuckGoModule {
	return new(DuckDuckGoModule)
}
func (d *DuckDuckGoModule) Name() string {
	return "DuckDuckGo"
}

func (d *DuckDuckGoModule) Description() string {
	return "Provides web search via ddg"
}

func (d *DuckDuckGoModule) CanBeDisabled() bool {
	return true
}

func (d *DuckDuckGoModule) UpdateSettings() {
	// this intentionally empty
}

func (d *DuckDuckGoModule) NeedsExternalData() bool {
	return false
}

func (d *DuckDuckGoModule) UpdateExternalData() {
	// this intentionally empty
}

func (d *DuckDuckGoModule) WriteExternalData(_ *os.File) {
	// this intentionally empty
}

func (d *DuckDuckGoModule) ReadExternalData(_ []byte) error {
	// this intentionally empty
	return nil
}

type DuckDuckGoSearchAction struct {
	query      string
	queryLabel string
}

func (d DuckDuckGoSearchAction) GetLabel() string {
	return "[ddg[] SEARCH " + d.queryLabel
}

func (d DuckDuckGoSearchAction) Run() string {
	url := "https://duckduckgo.com/?q=" + d.query
	if err := launchUrl(url); err != nil {
		log.Fatalf("Could not browse %s: %v", url, err)
	}
	return "Opened DDG search for " + d.queryLabel
}

func (d *DuckDuckGoModule) CreateActions(tags []Tag) []action {
	if len(tags) != 0 {
		for _, tag := range tags {
			if tag.matchMode != Contains {
				// Looks like the user is looking for something spectific - not
				// very likely they wanted to ddg
				return []action{}
			}
		}
		query := tags[0].value
		queryLabel := tags[0].value
		for i, tag := range tags {
			if i != 0 {
				query += "+" + tag.value
				queryLabel += " " + tag.value
			}
		}
		return []action{
			DuckDuckGoSearchAction{
				query:      query,
				queryLabel: queryLabel,
			},
		}
	}
	return []action{}
}
