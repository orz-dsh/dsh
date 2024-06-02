package dsh_core

import "regexp"

// region base

var projectNameCheckRegex = regexp.MustCompile("^[a-z][a-z0-9_]*$")

// endregion

// region project

type projectSchema struct {
	Name           string
	Path           string
	OptionDeclares projectSchemaOptionSet
	OptionVerifies projectSchemaOptionVerifySet
	ScriptSources  projectSchemaSourceSet
	ScriptImports  projectSchemaImportSet
	ConfigSources  projectSchemaSourceSet
	ConfigImports  projectSchemaImportSet
}

type projectSchemaSet []*projectSchema

func newProjectSchema(name string, path string, optionDeclares projectSchemaOptionSet, optionVerifies projectSchemaOptionVerifySet, scriptSources projectSchemaSourceSet, scriptImports projectSchemaImportSet, configSources projectSchemaSourceSet, configImports projectSchemaImportSet) *projectSchema {
	return &projectSchema{
		Name:           name,
		Path:           path,
		OptionDeclares: optionDeclares,
		OptionVerifies: optionVerifies,
		ScriptSources:  scriptSources,
		ScriptImports:  scriptImports,
		ConfigSources:  configSources,
		ConfigImports:  configImports,
	}
}

// endregion
