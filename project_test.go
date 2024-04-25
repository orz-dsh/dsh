package main

import (
	"dsh/dsh_core"
	"log"
	"testing"
)

func TestProject1(t *testing.T) {
	workspace, err := dsh_core.LoadWorkspace("")
	if err != nil {
		log.Panicf("%+v", err)
	}
	project, err := workspace.LoadLocalProject("./.test/app1")
	if err != nil {
		log.Panicf("%+v", err)
	}
	builder := project.NewBuilder()
	err = builder.Build("")
	if err != nil {
		log.Panicf("%+v", err)
	}
}

func TestProject2(t *testing.T) {
	workspace, err := dsh_core.LoadWorkspace("")
	if err != nil {
		log.Panicf("%+v", err)
	}
	_, err = workspace.LoadGitProject("", "https://github.com/orz-dsh/not-exist-project.git", nil, "main", nil)
	if err != nil {
		log.Panicf("%+v", err)
	}
}
