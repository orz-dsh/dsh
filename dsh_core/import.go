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
