package dsh_core

import (
	"dsh/dsh_utils"
	"path/filepath"
	"text/template"
)

func newTemplateFuncs() template.FuncMap {
	return template.FuncMap{}
}

func makeTemplate(config map[string]any, funcs template.FuncMap, templateSourcePath string, templateLibSourcePaths []string, outputTargetPath string) error {
	templateFiles := append([]string{templateSourcePath}, templateLibSourcePaths...)
	tpl, err := template.New(filepath.Base(templateSourcePath)).Funcs(funcs).Option("missingkey=error").ParseFiles(templateFiles...)
	if err != nil {
		return dsh_utils.WrapError(err, "template parse failed", map[string]any{
			"templateSourcePath":     templateSourcePath,
			"templateLibSourcePaths": templateLibSourcePaths,
		})
	}
	if err = dsh_utils.WriteTemplate(tpl, config, outputTargetPath); err != nil {
		return err
	}
	return nil
}
