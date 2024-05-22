package dsh_core

import (
	"github.com/expr-lang/expr/vm"
	"net/url"
)

// region workspace definition

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

// endregion

// region project definition

type projectSourceDefinition struct {
	Dir   string
	Files []string
	Match string
	match *vm.Program
}

type projectImportDefinition struct {
	Registry *projectImportRegistryDefinition
	Local    *projectImportLocalDefinition
	Git      *projectImportGitDefinition
	Match    string
	match    *vm.Program
}

type projectImportRegistryDefinition struct {
	Name string
	Path string
	Ref  string
}

type projectImportLocalDefinition struct {
	Dir string
}

type projectImportGitDefinition struct {
	Url string
	Ref string
	url *url.URL
	ref *gitRef
}

func newProjectSourceDefinition(dir string, files []string, match string, matchExpr *vm.Program) *projectSourceDefinition {
	return &projectSourceDefinition{
		Dir:   dir,
		Files: files,
		Match: match,
		match: matchExpr,
	}
}

func newProjectImportDefinition(registry *projectImportRegistryDefinition, local *projectImportLocalDefinition, git *projectImportGitDefinition, match string, matchExpr *vm.Program) *projectImportDefinition {
	return &projectImportDefinition{
		Registry: registry,
		Local:    local,
		Git:      git,
		Match:    match,
		match:    matchExpr,
	}
}

func newProjectImportRegistryDefinition(name string, path string, ref string) *projectImportRegistryDefinition {
	return &projectImportRegistryDefinition{
		Name: name,
		Path: path,
		Ref:  ref,
	}
}

func newProjectImportLocalDefinition(dir string) *projectImportLocalDefinition {
	return &projectImportLocalDefinition{
		Dir: dir,
	}
}

func newProjectImportGitDefinition(rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *gitRef) *projectImportGitDefinition {
	return &projectImportGitDefinition{
		Url: rawUrl,
		Ref: rawRef,
		url: parsedUrl,
		ref: parsedRef,
	}
}

// endregion
