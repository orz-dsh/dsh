package main

import (
	"dsh/dsh_core"
	"dsh/dsh_utils"
	"testing"
)

func TestProject1(t *testing.T) {
	logger := dsh_utils.NewLogger(dsh_utils.LogLevelAll)
	workspace, err := dsh_core.OpenWorkspace("", logger)
	if err != nil {
		logger.Panic("%+v", err)
	}
	project, err := workspace.OpenLocalProject("./.test/app1", map[string]string{
		"_os":  "linux",
		"test": "a",
	})
	if err != nil {
		logger.Panic("%+v", err)
	}
	err = project.MakeScript("")
	if err != nil {
		logger.Panic("%+v", err)
	}
}

func TestProject2(t *testing.T) {
	logger := dsh_utils.NewLogger(dsh_utils.LogLevelAll)
	workspace, err := dsh_core.OpenWorkspace("", logger)
	if err != nil {
		logger.Panic("%+v", err)
	}
	_, err = workspace.OpenGitProject("https://github.com/orz-dsh/not-exist-project.git", "main", map[string]string{
		"option1": "value1",
	})
	if err != nil {
		logger.Panic("%+v", err)
	}
}
