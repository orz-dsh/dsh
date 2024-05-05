package dsh_core

import (
	"dsh/dsh_utils"
	"path/filepath"
	"strings"
	"text/template"
)

func shInitApp() string {
	return `if [ -z "${DSH_APP_DIR}" ]; then
  DSH_APP_DIR="$(dirname "$(dirname "$(readlink -f "$0")")")"
  export DSH_APP_DIR
fi`
}

func shImport(importName string) string {
	importEnvVar := "DSH_IMPORT_" + strings.ToUpper(importName)
	return `if [ -z "${DSH_IMPORT_` + strings.ToUpper(importName) + `}" ]; then
  . "${DSH_APP_DIR}/` + importName + `/lib.sh"
  ` + importEnvVar + `="true"
  export ` + importEnvVar + `
fi`
}

func newTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"SH_INIT_APP": shInitApp,
		"SH_IMPORT":   shImport,
	}
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
