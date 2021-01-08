package main

type Module interface {
	Name() string
	Description() string

	CanBeDisabled() bool
	NeedsExternalData() bool
	UpdateSettings()
}
