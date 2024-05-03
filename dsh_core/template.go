package dsh_core

import (
	"dsh/dsh_utils"
	"path/filepath"
	"text/template"
)

func newTemplateFuncs() template.FuncMap {
	return template.FuncMap{}
}

func makeTemplate(env map[string]any, funcs template.FuncMap, templateSourcePath string, templateLibSourcePaths []string, outputTargetPath string) error {
	templateFiles := append([]string{templateSourcePath}, templateLibSourcePaths...)
	tpl, err := template.New(filepath.Base(templateSourcePath)).Funcs(funcs).Option("missingkey=error").ParseFiles(templateFiles...)
	if err != nil {
		return errW(err, "make template error",
			reason("parse template error"),
			kv("templateSourcePath", templateSourcePath),
			kv("templateLibSourcePaths", templateLibSourcePaths),
		)
	}
	if err = dsh_utils.WriteTemplate(tpl, env, outputTargetPath); err != nil {
		return errW(err, "make template error",
			reason("write template error"),
			kv("outputTargetPath", outputTargetPath),
		)
	}
	return nil
}
