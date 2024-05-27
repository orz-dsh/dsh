package dsh_core

type AppFactory struct {
	workspace *Workspace
	Profile   *AppProfile
}

func newAppFactory(workspace *Workspace, profile *AppProfile) *AppFactory {
	return &AppFactory{
		workspace: workspace,
		Profile:   profile,
	}
}

func (f *AppFactory) MakeApp(rawLink string) (*App, error) {
	link, err := ParseProjectLink(rawLink)
	if err != nil {
		return nil, err
	}
	resolvedLink, err := f.Profile.resolveProjectLink(link)
	if err != nil {
		return nil, err
	}
	app, err := loadApp(f.workspace, resolvedLink, f.Profile)
	if err != nil {
		return nil, err
	}
	return app, nil
}
