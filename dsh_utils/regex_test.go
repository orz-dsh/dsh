package dsh_utils

import (
	"regexp"
	"testing"
)

func TestRegexMatch(t *testing.T) {
	re := regexp.MustCompile("^git:https://github.com/group/(?P<path>.+).git#ref=(?P<ref>.+)$")
	matched, values := RegexMatch(re, "git:https://github.com/group/project.git#ref=master")
	if !matched {
		t.Fatal("not matched")
	}
	t.Log(values)

	re = regexp.MustCompile("^dir:C:/some/../path/(?P<path>.+)$")
	matched, values = RegexMatch(re, "dir:C:/some/../path/a/b/c")
	if !matched {
		t.Fatal("not matched")
	}
	t.Log(values)
}
