package main

import "os"

type Module interface {
	Name() string
	Description() string

	CanBeDisabled() bool
	UpdateSettings()
	NeedsExternalData() bool
	UpdateExternalData()
	WriteExternalData(file *os.File)
	CreateActions(tags []Tag) []action
	ReadExternalData(data []byte) error
}
