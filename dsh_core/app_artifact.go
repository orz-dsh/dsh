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
	shellName := a.app.context.Option.GenericItems.getShell()
	entity, err := a.context.profile.getWorkspaceShellEntity(shellName)
	if err != nil {
		return nil, errW(err, "create artifact executor error",
			reason("get workspace shell entity error"),
			kv("shellName", shellName),
		)
	}

	targetName, err := a.getTargetName(entity, targetGlob)
	if err != nil {
		return nil, errW(err, "create artifact executor error",
			reason("get target name error"),
			kv("shellName", shellName),
			kv("targetGlob", targetGlob),
		)
	}
	targetPath := filepath.Join(a.OutputPath, targetName)

	shellArgs, err := a.getShellArgs(entity, targetGlob, targetName, targetPath)
	if err != nil {
		return nil, errW(err, "create artifact executor error",
			reason("get shell args error"),
			kv("entity", entity),
			kv("targetGlob", targetGlob),
			kv("targetName", targetName),
			kv("targetPath", targetPath),
		)
	}

	executor = &appArtifactExecutor{
		context:    a.context,
		ShellName:  entity.Name,
		ShellPath:  entity.Path,
		ShellArgs:  shellArgs,
		TargetGlob: targetGlob,
		TargetName: targetName,
		TargetPath: targetPath,
	}
	return executor, nil
}

func (a *AppArtifact) getTargetName(entity *workspaceShellEntity, targetGlob string) (targetName string, err error) {
	if targetGlob == "" {
		return "", errN("get target name error",
			reason("target glob empty"),
		)
	}
	targetGlob = strings.ReplaceAll(targetGlob, "\\", "/")
	slashCount := strings.Count(targetGlob, "/")
	if slashCount == 0 {
		targetGlob = a.app.project.Manifest.Name + "/" + targetGlob
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

func (a *AppArtifact) getShellArgs(entity *workspaceShellEntity, targetGlob string, targetName string, targetPath string) (shellArgs []string, err error) {
	args := entity.Args
	if len(args) == 0 {
		shellArgs = []string{targetPath}
	} else {
		evaluator := a.context.evaluator.SetRootData("executor", map[string]any{
			"shell": map[string]any{
				"name": entity.Name,
				"path": entity.Path,
			},
			"target": map[string]any{
				"glob": targetGlob,
				"name": targetName,
				"path": targetPath,
			},
		})
		shellArgs, err = entity.Args.getArgs(evaluator)
		if err != nil {
			return nil, errW(err, "get shell args error",
				reason("eval shell args error"),
				kv("entity", entity),
				kv("targetGlob", targetGlob),
				kv("targetName", targetName),
				kv("targetPath", targetPath),
			)
		}
	}
	return shellArgs, nil
}

// endregion

// region executor

type appArtifactExecutor struct {
	context    *appContext
	ShellName  string
	ShellPath  string
	ShellArgs  []string
	TargetGlob string
	TargetName string
	TargetPath string
}

func (e *appArtifactExecutor) executeInChildProcess() (exitCode int, err error) {
	startTime := time.Now()
	cmd := exec.Command(e.ShellPath, e.ShellArgs...)
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
	execArgs := append([]string{e.ShellName}, e.ShellArgs...)
	e.context.logger.InfoDesc("execute artifact in this process start",
		kv("executor", e),
		kv("execArgs", execArgs),
	)
	err = syscall.Exec(e.ShellPath, execArgs, os.Environ())
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
