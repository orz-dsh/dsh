package dsh_core

import (
	"os"
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

func executeFileTemplate(sourcePath string, libPaths []string, targetPath string, data map[string]any, funcs template.FuncMap) error {
	tpl := template.New(filepath.Base(sourcePath)).Option("missingkey=error")
	if funcs != nil {
		tpl = tpl.Funcs(funcs)
	}
	files := append([]string{sourcePath}, libPaths...)
	tpl, err := tpl.ParseFiles(files...)
	if err != nil {
		return errW(err, "execute file template error",
			reason("parse template error"),
			kv("sourcePath", sourcePath),
			kv("libPaths", libPaths),
		)
	}

	if err = os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
		return errW(err, "execute file template error",
			reason("make target dir error"),
			kv("targetPath", targetPath),
		)
	}

	targetFile, err := os.Create(targetPath)
	if err != nil {
		return errW(err, "execute file template error",
			reason("create target file error"),
			kv("targetPath", targetPath),
		)
	}
	defer targetFile.Close()

	err = tpl.Execute(targetFile, data)
	if err != nil {
		return errW(err, "execute file template error",
			reason("execute template error"),
			kv("sourcePath", sourcePath),
			kv("libPaths", libPaths),
			kv("targetPath", targetPath),
			kv("data", data),
			kv("funcs", funcs),
		)
	}
	return nil
}

func executeStringTemplate(str string, data map[string]any, funcs template.FuncMap) (string, error) {
	tpl := template.New("StringTemplate").Option("missingkey=error")
	if funcs != nil {
		tpl = tpl.Funcs(funcs)
	}
	tpl, err := tpl.Parse(str)
	if err != nil {
		return "", errW(err, "execute string template error",
			reason("parse template error"),
			kv("str", str),
			kv("data", data),
			kv("funcs", funcs),
		)
	}
	var writer strings.Builder
	err = tpl.Execute(&writer, data)
	if err != nil {
		return "", errW(err, "execute string template error",
			reason("execute template error"),
			kv("str", str),
			kv("data", data),
			kv("funcs", funcs),
		)
	}
	return writer.String(), nil
}
