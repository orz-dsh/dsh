package dsh_core

import "dsh/dsh_utils"

type appArtifactEvaluator struct {
	data *appProfileEvalData
}

func newAppArtifactEvaluator(data *appProfileEvalData) *appArtifactEvaluator {
	return &appArtifactEvaluator{
		data: data,
	}
}

func (e *appArtifactEvaluator) evalShellArgs(shellName, shellPath, targetGlob, targetName, targetPath string, args []string) ([]string, error) {
	evalData := e.data.mergeMap(map[string]any{
		"shell": map[string]any{
			"name": shellName,
			"path": shellPath,
		},
		"target": map[string]any{
			"glob": targetGlob,
			"name": targetName,
			"path": targetPath,
		},
	})
	var shellArgs []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		shellArg, err := dsh_utils.EvalStringTemplate(arg, evalData, nil)
		if err != nil {
			return nil, errW(err, "eval shell args error",
				reason("execute arg template error"),
				kv("index", i),
				kv("args", args),
			)
		}
		shellArgs = append(shellArgs, shellArg)
	}
	return shellArgs, nil
}
