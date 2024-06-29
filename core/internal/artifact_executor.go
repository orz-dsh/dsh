package internal

import (
	"errors"
	. "github.com/orz-dsh/dsh/utils"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// region executor

type ArtifactExecutor struct {
	Application *ApplicationCore `desc:"-"`
	Name        string
	File        string
	Args        []string
	TargetGlob  string
	TargetName  string
	TargetFile  string
}

func NewArtifactExecutor(application *ApplicationCore, name, file string, args []string, targetGlob, targetName, targetFile string) *ArtifactExecutor {
	return &ArtifactExecutor{
		Application: application,
		Name:        name,
		File:        file,
		Args:        args,
		TargetGlob:  targetGlob,
		TargetName:  targetName,
		TargetFile:  targetFile,
	}
}

func (e *ArtifactExecutor) ExecuteInChildProcess() (exitCode int, err error) {
	startTime := time.Now()
	cmd := exec.Command(e.File, e.Args...)
	cmd.Stdout = e.Application.Logger.GetInfoWriter()
	cmd.Stderr = e.Application.Logger.GetErrorWriter()
	err = cmd.Start()
	if err != nil {
		return -1, ErrW(err, "execute artifact in child process error",
			Reason("start command error"),
			KV("executor", e),
		)
	}
	pid := cmd.Process.Pid
	e.Application.Logger.InfoDesc("execute artifact in child process start",
		KV("executor", e),
		KV("pid", pid),
	)
	err = cmd.Wait()
	exitCode = 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			return -1, ErrW(err, "execute artifact in child process error",
				Reason("wait command exit error"),
				KV("executor", e),
				KV("pid", pid),
			)
		}
	}
	e.Application.Logger.InfoDesc("execute artifact in child process finish",
		KV("elapsed", time.Since(startTime)),
		KV("exitCode", exitCode),
	)
	return exitCode, nil
}

func (e *ArtifactExecutor) ExecuteInThisProcess() (err error) {
	execArgs := append([]string{e.Name}, e.Args...)
	e.Application.Logger.InfoDesc("execute artifact in this process start",
		KV("executor", e),
		KV("execArgs", execArgs),
	)
	err = syscall.Exec(e.File, execArgs, os.Environ())
	if err != nil {
		return ErrW(err, "execute artifact in this process error",
			Reason("system exec error"),
			KV("executor", e),
			KV("execArgs", execArgs),
		)
	}
	return nil
}

// endregion
