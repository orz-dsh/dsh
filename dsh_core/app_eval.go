package dsh_core

import "dsh/dsh_utils"

type appEvalData = dsh_utils.EvalData

func makeAppEvalData(workspace *Workspace) (*appEvalData, error) {
	local, err := initAppEvalDataLocal(workspace)
	if err != nil {
		return nil, err
	}

	data := dsh_utils.NewEvalData().Data("local", local)
	return data, nil
}

func initAppEvalDataLocal(workspace *Workspace) (map[string]any, error) {
	workingDir, err := dsh_utils.GetWorkingDir()
	if err != nil {
		return nil, err
	}

	data := map[string]any{
		"working_dir":          workingDir,
		"workspace_dir":        workspace.path,
		"runtime_version":      dsh_utils.GetRuntimeVersion(),
		"runtime_version_code": dsh_utils.GetRuntimeVersionCode(),
	}
	return data, nil
}
