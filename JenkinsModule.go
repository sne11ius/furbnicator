package main

type JenkinsModule struct {
}

func NewJenkinsModule() *JenkinsModule {
	return new(JenkinsModule)
}

func (j JenkinsModule) Name() string {
	return "Jenkins"
}

func (j JenkinsModule) Description() string {
	return "Provides access to jenkins jobs"
}

func (j JenkinsModule) CanBeDisabled() bool {
	return true
}

func (j JenkinsModule) UpdateSettings() {
	panic("not yet")
}
