package dsh_core

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// region artifact

type AppArtifact struct {
	app             *App
	context         *appContext
	targetNames     []string
	targetNamesDict map[string]bool
	OutputPath      string
}

func newAppArtifact(app *App, targetNames []string, outputPath string) *AppArtifact {
	var targetNamesDict = map[string]bool{}
	for i := 0; i < len(targetNames); i++ {
		targetNamesDict[targetNames[i]] = true
	}
	return &AppArtifact{
		app:             app,
		context:         app.context,
		targetNames:     targetNames,
		targetNamesDict: targetNamesDict,
		OutputPath:      outputPath,
	}
}

func (a *AppArtifact) DescExtraKeyValues() KVS {
	return KVS{
		kv("targetNames", a.targetNames),
	}
}

func (a *AppArtifact) ExecuteInChildProcess(targetGlob string) (exitCode int, err error) {
	executor, err := a.createExecutor(targetGlob)
	if err != nil {
		return -1, errW(err, "execute artifact in child process error",
			reason("create executor error"),
			kv("targetGlob", targetGlob),
		)
	}
	exitCode, err = executor.executeInChildProcess()
	if err != nil {
		return -1, err
	}
	return exitCode, nil
}

func (a *AppArtifact) ExecuteInThisProcess(targetGlob string) (err error) {
	executor, err := a.createExecutor(targetGlob)
	if err != nil {
		return errW(err, "execute artifact in this process error",
			reason("create executor error"),
			kv("targetGlob", targetGlob),
		)
	}
	err = executor.executeInThisProcess()
	if err != nil {
		return err
	}
	return nil
}

func (a *AppArtifact) createExecutor(targetGlob string) (executor *appArtifactExecutor, err error) {
	name := a.app.context.option.GenericItems.getExecutor()
	setting, err := a.context.profile.getWorkspaceExecutorSetting(name)
	if err != nil {
		return nil, errW(err, "create artifact executor error",
			reason("get workspace executor setting error"),
			kv("name", name),
		)
	}

	targetName, err := a.getTargetName(setting, targetGlob)
	if err != nil {
		return nil, errW(err, "create artifact executor error",
			reason("get target name error"),
			kv("name", name),
			kv("targetGlob", targetGlob),
		)
	}
	targetPath := filepath.Join(a.OutputPath, targetName)

	args, err := a.getExecutorArgs(setting, targetGlob, targetName, targetPath)
	if err != nil {
		return nil, errW(err, "create artifact executor error",
			reason("get executor args error"),
			kv("setting", setting),
			kv("targetGlob", targetGlob),
			kv("targetName", targetName),
			kv("targetPath", targetPath),
		)
	}

	executor = &appArtifactExecutor{
		context:    a.context,
		Name:       setting.Name,
		Path:       setting.Path,
		Args:       args,
		TargetGlob: targetGlob,
		TargetName: targetName,
		TargetPath: targetPath,
	}
	return executor, nil
}

func (a *AppArtifact) getTargetName(entity *workspaceExecutorSetting, targetGlob string) (targetName string, err error) {
	if targetGlob == "" {
		return "", errN("get target name error",
			reason("target glob empty"),
		)
	}
	targetGlob = strings.ReplaceAll(targetGlob, "\\", "/")
	slashCount := strings.Count(targetGlob, "/")
	if slashCount == 0 {
		targetGlob = a.app.mainProjectName + "/" + targetGlob
	} else if slashCount > 1 {
		return "", errN("get target name error",
			reason("target glob invalid"),
			kv("targetGlob", targetGlob),
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

	return "", errN("get target name error",
		reason("target name not found"),
		kv("entity", entity),
		kv("targetGlob", targetGlob),
	)
}

func (a *AppArtifact) getExecutorArgs(setting *workspaceExecutorSetting, targetGlob string, targetName string, targetPath string) (executorArgs []string, err error) {
	args := setting.Args
	if len(args) == 0 {
		executorArgs = []string{targetPath}
	} else {
		evaluator := a.context.evaluator.SetRootData("executor", map[string]any{
			"name": setting.Name,
			"path": setting.Path,
			"target": map[string]any{
				"glob": targetGlob,
				"name": targetName,
				"path": targetPath,
			},
		})
		executorArgs, err = setting.getArgs(evaluator)
		if err != nil {
			return nil, errW(err, "get executor args error",
				reason("eval executor args error"),
				kv("setting", setting),
				kv("targetGlob", targetGlob),
				kv("targetName", targetName),
				kv("targetPath", targetPath),
			)
		}
	}
	return executorArgs, nil
}

// endregion

// region executor

type appArtifactExecutor struct {
	context    *appContext
	Name       string
	Path       string
	Args       []string
	TargetGlob string
	TargetName string
	TargetPath string
}

func (e *appArtifactExecutor) executeInChildProcess() (exitCode int, err error) {
	startTime := time.Now()
	cmd := exec.Command(e.Path, e.Args...)
	cmd.Stdout = e.context.logger.GetInfoWriter()
	cmd.Stderr = e.context.logger.GetErrorWriter()
	err = cmd.Start()
	if err != nil {
		return -1, errW(err, "execute artifact in child process error",
			reason("start command error"),
			kv("executor", e),
		)
	}
	pid := cmd.Process.Pid
	e.context.logger.InfoDesc("execute artifact in child process start",
		kv("executor", e),
		kv("pid", pid),
	)
	err = cmd.Wait()
	exitCode = 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			return -1, errW(err, "execute artifact in child process error",
				reason("wait command exit error"),
				kv("executor", e),
				kv("pid", pid),
			)
		}
	}
	e.context.logger.InfoDesc("execute artifact in child process finish",
		kv("elapsed", time.Since(startTime)),
		kv("exitCode", exitCode),
	)
	return exitCode, nil
}

func (e *appArtifactExecutor) executeInThisProcess() (err error) {
	execArgs := append([]string{e.Name}, e.Args...)
	e.context.logger.InfoDesc("execute artifact in this process start",
		kv("executor", e),
		kv("execArgs", execArgs),
	)
	err = syscall.Exec(e.Path, execArgs, os.Environ())
	if err != nil {
		return errW(err, "execute artifact in this process error",
			reason("system exec error"),
			kv("executor", e),
			kv("execArgs", execArgs),
		)
	}
	return nil
}

// endregion
