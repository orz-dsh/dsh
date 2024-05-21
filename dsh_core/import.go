package dsh_core

type importRegistryDefinition struct {
	Name  string
	Local *importLocalDefinition
	Git   *importGitDefinition
}

type importRedirectDefinition struct {
	Prefix string
	Local  *importLocalDefinition
	Git    *importGitDefinition
}

type importLocalDefinition struct {
	Dir string
}

type importGitDefinition struct {
	Url string
	Ref string
}

func newImportRegistryLocalDefinition(name string, dir string) *importRegistryDefinition {
	return &importRegistryDefinition{
		Name: name,
		Local: &importLocalDefinition{
			Dir: dir,
		},
	}
}

func newImportRegistryGitDefinition(name string, url string, ref string) *importRegistryDefinition {
	return &importRegistryDefinition{
		Name: name,
		Git: &importGitDefinition{
			Url: url,
			Ref: ref,
		},
	}
}

func newImportRedirectLocalDefinition(prefix string, dir string) *importRedirectDefinition {
	return &importRedirectDefinition{
		Prefix: prefix,
		Local: &importLocalDefinition{
			Dir: dir,
		},
	}
}

func newImportRedirectGitDefinition(prefix string, url string, ref string) *importRedirectDefinition {
	return &importRedirectDefinition{
		Prefix: prefix,
		Git: &importGitDefinition{
			Url: url,
			Ref: ref,
		},
	}
}
