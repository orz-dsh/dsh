package internal

import (
	. "github.com/orz-dsh/dsh/core/internal/setting"
	. "github.com/orz-dsh/dsh/utils"
	"path/filepath"
	"strings"
)

// region ArtifactCore

type ArtifactCore struct {
	Application     *ApplicationCore
	OutputDir       string
	TargetNames     []string
	targetNamesDict map[string]bool
}

func NewArtifactCore(application *ApplicationCore, outputDir string, targetNames []string, targetNamesDict map[string]bool) *ArtifactCore {
	return &ArtifactCore{
		Application:     application,
		OutputDir:       outputDir,
		TargetNames:     targetNames,
		targetNamesDict: targetNamesDict,
	}
}

func (a *ArtifactCore) ExecuteInChildProcess(targetGlob string) (exitCode int, err error) {
	executor, err := a.createExecutor(targetGlob)
	if err != nil {
		return -1, ErrW(err, "execute artifact in child process error",
			Reason("create executor error"),
			KV("targetGlob", targetGlob),
		)
	}
	exitCode, err = executor.ExecuteInChildProcess()
	if err != nil {
		return -1, err
	}
	return exitCode, nil
}

func (a *ArtifactCore) ExecuteInThisProcess(targetGlob string) (err error) {
	executor, err := a.createExecutor(targetGlob)
	if err != nil {
		return ErrW(err, "execute artifact in this process error",
			Reason("create executor error"),
			KV("targetGlob", targetGlob),
		)
	}
	err = executor.ExecuteInThisProcess()
	if err != nil {
		return err
	}
	return nil
}

func (a *ArtifactCore) createExecutor(targetGlob string) (executor *ArtifactExecutor, err error) {
	name := a.Application.Option.Common.Executor
	setting, err := a.Application.Setting.GetExecutorItemSetting(name)
	if err != nil {
		return nil, ErrW(err, "create artifact executor error",
			Reason("get workspace executor setting error"),
			KV("name", name),
		)
	}

	targetName, err := a.getTargetName(setting, targetGlob)
	if err != nil {
		return nil, ErrW(err, "create artifact executor error",
			Reason("get target name error"),
			KV("name", name),
			KV("targetGlob", targetGlob),
		)
	}
	targetFile := filepath.Join(a.OutputDir, targetName)

	args, err := a.getExecutorArgs(setting, targetGlob, targetName, targetFile)
	if err != nil {
		return nil, ErrW(err, "create artifact executor error",
			Reason("get executor args error"),
			KV("setting", setting),
			KV("targetGlob", targetGlob),
			KV("targetName", targetName),
			KV("targetFile", targetFile),
		)
	}

	executor = NewArtifactExecutor(a.Application, setting.Name, setting.File, args, targetGlob, targetName, targetFile)
	return executor, nil
}

func (a *ArtifactCore) getTargetName(entity *ExecutorItemSetting, targetGlob string) (targetName string, err error) {
	if targetGlob == "" {
		return "", ErrN("get target name error",
			Reason("target glob empty"),
		)
	}
	targetGlob = strings.ReplaceAll(targetGlob, "\\", "/")
	slashCount := strings.Count(targetGlob, "/")
	if slashCount == 0 {
		targetGlob = a.Application.MainProject.Name + "/" + targetGlob
	} else if slashCount > 1 {
		return "", ErrN("get target name error",
			Reason("target glob invalid"),
			KV("targetGlob", targetGlob),
		)
	}

	targetName = targetGlob
	if a.targetNamesDict[targetName] {
		return targetName, nil
	}

	exts := entity.Exts
	for i := 0; i < len(exts); i++ {
		targetName = targetGlob + exts[i]
		if a.targetNamesDict[targetName] {
			return targetName, nil
		}
	}

	return "", ErrN("get target name error",
		Reason("target name not found"),
		KV("entity", entity),
		KV("targetGlob", targetGlob),
	)
}

func (a *ArtifactCore) getExecutorArgs(setting *ExecutorItemSetting, targetGlob, targetName, targetFile string) (executorArgs []string, err error) {
	args := setting.Args
	if len(args) == 0 {
		executorArgs = []string{targetFile}
	} else {
		evaluator := a.Application.Evaluator.SetRootData("executor", map[string]any{
			"executor_name": setting.Name,
			"executor_file": setting.File,
			"target_glob":   targetGlob,
			"target_name":   targetName,
			"target_file":   targetFile,
		})
		executorArgs, err = setting.GetArgs(evaluator)
		if err != nil {
			return nil, ErrW(err, "get executor args error",
				Reason("eval executor args error"),
				KV("setting", setting),
				KV("targetGlob", targetGlob),
				KV("targetName", targetName),
				KV("targetFile", targetFile),
			)
		}
	}
	return executorArgs, nil
}

// endregion
