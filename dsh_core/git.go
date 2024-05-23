package dsh_core

import (
	"github.com/go-git/go-git/v5/plumbing"
	"strings"
)

type gitRef struct {
	Raw           string
	Type          gitRefType
	ReferenceName plumbing.ReferenceName
}

type gitRefType string

const (
	gitRefTypeBranch gitRefType = "branch"
	gitRefTypeTag    gitRefType = "tag"
)

func parseGitRef(rawRef string) *gitRef {
	if strings.HasPrefix(rawRef, "tags/") {
		tag := strings.TrimPrefix(rawRef, "tags/")
		return &gitRef{
			Raw:           rawRef,
			Type:          gitRefTypeTag,
			ReferenceName: plumbing.NewTagReferenceName(tag),
		}
	}
	return &gitRef{
		Raw:           rawRef,
		Type:          gitRefTypeBranch,
		ReferenceName: plumbing.NewBranchReferenceName(rawRef),
	}
}
