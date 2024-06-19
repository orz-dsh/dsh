package dsh_core

import (
	"dsh/dsh_utils"
	"strings"
)

func projectScriptTemplateFuncShInitApp() string {
	return `if [ -z "${DSH_APP_DIR}" ]; then
  DSH_APP_DIR="$(dirname "$(dirname "$(readlink -f "$0")")")"
  export DSH_APP_DIR
fi`
}

func projectScriptTemplateFuncShImport(importName string) string {
	importEnvVar := "DSH_IMPORT_" + strings.ReplaceAll(strings.ToUpper(importName), "-", "_")
	return `if [ -z "${` + importEnvVar + `}" ]; then
  . "${DSH_APP_DIR}/` + importName + `/lib.sh"
  export ` + importEnvVar + `="true"
fi`
}

func newProjectScriptTemplateFuncs() dsh_utils.EvalFuncs {
	return dsh_utils.EvalFuncs{
		"SH_INIT_APP": projectScriptTemplateFuncShInitApp,
		"SH_IMPORT":   projectScriptTemplateFuncShImport,
	}
}
