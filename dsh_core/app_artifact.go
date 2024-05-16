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

type AppArtifact struct {
	app             *App
	context         *appContext
	targetNames     []string
	targetNamesDict map[string]bool
	OutputPath      string
}

type appArtifactExecutor struct {
	context    *appContext
	shellName  string
	shellPath  string
	shellArgs  []string
	targetGlob string
	targetName string
	targetPath string
}

func newAppArtifact(app *App, targetNames []string, outputPath string) *AppArtifact {
	var targetNamesDict = make(map[string]bool)
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
	shellName := a.app.context.option.getGlobalOptionsShell()
	shellPath, err := a.getShellPath(shellName)
	if err != nil {
		return nil, errW(err, "create artifact executor error",
			reason("get shell path error"),
			kv("shellName", shellName),
		)
	}

	targetName, err := a.getTargetName(shellName, targetGlob)
	if err != nil {
		return nil, errW(err, "create artifact executor error",
			reason("get target name error"),
			kv("shellName", shellName),
			kv("targetGlob", targetGlob),
		)
	}
	targetPath := filepath.Join(a.OutputPath, targetName)

	shellArgs, err := a.getShellArgs(shellName, shellPath, targetGlob, targetName, targetPath)
	if err != nil {
		return nil, errW(err, "create artifact executor error",
			reason("get shell args error"),
			kv("shellName", shellName),
			kv("shellPath", shellPath),
			kv("targetGlob", targetGlob),
			kv("targetName", targetName),
			kv("targetPath", targetPath),
		)
	}

	executor = &appArtifactExecutor{
		context:    a.context,
		shellName:  shellName,
		shellPath:  shellPath,
		shellArgs:  shellArgs,
		targetGlob: targetGlob,
		targetName: targetName,
		targetPath: targetPath,
	}
	return executor, nil
}

func (a *AppArtifact) getShellPath(shellName string) (shellPath string, err error) {
	shellPath = a.context.workspace.manifest.getShellPath(shellName)
	if shellPath != "" {
		return shellPath, nil
	}

	shellPath, err = exec.LookPath(shellName)
	if err != nil {
		return "", errW(err, "get shell path error",
			reason("look path error"),
			kv("shellName", shellName),
		)
	}
	return shellPath, nil
}

func (a *AppArtifact) getTargetName(shellName string, targetGlob string) (targetName string, err error) {
	if targetGlob == "" {
		return "", errN("get target name error",
			reason("target glob empty"),
		)
	}
	targetGlob = strings.ReplaceAll(targetGlob, "\\", "/")
	slashCount := strings.Count(targetGlob, "/")
	if slashCount == 0 {
		targetGlob = a.app.project.manifest.Name + "/" + targetGlob
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

	exts := a.context.workspace.manifest.getShellExts(shellName)
	for i := 0; i < len(exts); i++ {
		targetName = targetGlob + exts[i]
		if a.targetNamesDict[targetName] {
			return targetName, nil
		}
	}

	return "", errN("get target name error",
		reason("target name not found"),
		kv("shellName", shellName),
		kv("targetGlob", targetGlob),
	)
}

func (a *AppArtifact) getShellArgs(shellName string, shellPath string, targetGlob string, targetName string, targetPath string) (shellArgs []string, err error) {
	args := a.context.workspace.manifest.getShellArgs(shellName)
	if len(args) == 0 {
		shellArgs = []string{targetPath}
	} else {
		tplData := map[string]any{
			"shell": map[string]any{
				"name": shellName,
				"path": shellPath,
			},
			"target": map[string]any{
				"glob": targetGlob,
				"name": targetName,
				"path": targetPath,
			},
		}
		for i := 0; i < len(args); i++ {
			arg := args[i]
			shellArg, err := executeStringTemplate(arg, tplData, nil)
			if err != nil {
				return nil, errW(err, "get shell args error",
					reason("execute arg template error"),
					kv("index", i),
					kv("arg", arg),
				)
			}
			shellArgs = append(shellArgs, shellArg)
		}
	}
	return shellArgs, nil
}

func (e *appArtifactExecutor) executeInChildProcess() (exitCode int, err error) {
	startTime := time.Now()
	cmd := exec.Command(e.shellPath, e.shellArgs...)
	cmd.Stdout = e.context.logger.GetInfoWriter()
	cmd.Stderr = e.context.logger.GetErrorWriter()
	err = cmd.Start()
	if err != nil {
		return -1, errW(err, "execute artifact in child process error",
			reason("start command error"),
			kv("path", e.shellPath),
			kv("args", e.shellArgs),
		)
	}
	pid := cmd.Process.Pid
	e.context.logger.InfoDesc("execute artifact in child process start",
		kv("shellName", e.shellName),
		kv("shellPath", e.shellPath),
		kv("shellArgs", e.shellArgs),
		kv("targetGlob", e.targetGlob),
		kv("targetName", e.targetName),
		kv("targetPath", e.targetPath),
		kv("childPid", pid),
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
				kv("path", e.shellPath),
				kv("args", e.shellArgs),
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
	execArgs := append([]string{e.shellName}, e.shellArgs...)
	e.context.logger.InfoDesc("execute artifact in this process start",
		kv("shellName", e.shellName),
		kv("shellPath", e.shellPath),
		kv("shellArgs", e.shellArgs),
		kv("targetGlob", e.targetGlob),
		kv("targetName", e.targetName),
		kv("targetPath", e.targetPath),
		kv("execArgs", execArgs),
	)
	err = syscall.Exec(e.shellPath, execArgs, os.Environ())
	if err != nil {
		return errW(err, "execute artifact in this process error",
			reason("system exec error"),
			kv("path", e.shellArgs),
			kv("args", execArgs),
		)
	}
	return nil
}
