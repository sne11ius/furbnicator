package main

type BitbucketModule struct {
}

func NewBitbucketModule() *BitbucketModule {
	return new(BitbucketModule)
}

func (b BitbucketModule) Name() string {
	return "Bitbucket"
}

func (b BitbucketModule) Description() string {
	return "Provides access to bitbucket projects"
}

func (b BitbucketModule) CanBeDisabled() bool {
	return true
}

func (b BitbucketModule) UpdateSettings() {
	panic("not yet")
}
