package dsh_core

import (
	"github.com/go-git/go-git/v5/plumbing"
	"strings"
)

type gitRef struct {
	raw           string
	refType       gitRefType
	referenceName plumbing.ReferenceName
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
			raw:           rawRef,
			refType:       gitRefTypeTag,
			referenceName: plumbing.NewTagReferenceName(tag),
		}
	}
	return &gitRef{
		raw:           rawRef,
		refType:       gitRefTypeBranch,
		referenceName: plumbing.NewBranchReferenceName(rawRef),
	}
}
