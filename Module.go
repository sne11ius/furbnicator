package main

type Module interface {
	Name() string
	Description() string

	CanBeDisabled() bool
	UpdateSettings()
	NeedsExternalData() bool
	UpdateExternalData()
}
