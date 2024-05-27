package dsh_core

import "dsh/dsh_utils"

type appArtifactEvaluator struct {
	data *appEvalData
}

func newAppArtifactEvaluator(data *appEvalData) *appArtifactEvaluator {
	return &appArtifactEvaluator{
		data: data,
	}
}

func (e *appArtifactEvaluator) evalShellArgs(shellName, shellPath, targetGlob, targetName, targetPath string, args []string) ([]string, error) {
	evalData := e.data.MainData("executor", map[string]any{
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
	replacer := dsh_utils.NewEvalReplacer(evalData, nil)
	var shellArgs []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		shellArg, err := replacer.Replace(arg)
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
