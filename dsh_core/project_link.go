package dsh_core

import (
	"github.com/go-git/go-git/v5/plumbing"
	"net/url"
	"path/filepath"
	"strings"
)

type projectLink struct {
	Raw        string
	Normalized string
	Type       projectLinkType
	Registry   *projectLinkRegistry
	Dir        *projectLinkDir
	Git        *projectLinkGit
}

type projectLinkRegistry struct {
	Name string
	Path string
	Ref  string
	ref  *projectLinkGitRef
}

type projectLinkDir struct {
	Raw  string
	Path string
}

type projectLinkGit struct {
	Url       string
	Ref       string
	parsedUrl *url.URL
	parsedRef *projectLinkGitRef
}

type projectLinkGitRef struct {
	Raw           string
	Normalized    string
	Type          projectLinkGitRefType
	Name          string
	ReferenceName plumbing.ReferenceName
}

type projectLinkType string

const (
	ProjectLinkTypeRegistry projectLinkType = "registry"
	ProjectLinkTypeDir      projectLinkType = "dir"
	ProjectLinkTypeGit      projectLinkType = "git"
)

type projectLinkGitRefType string

const (
	projectLinkGitRefTypeBranch projectLinkGitRefType = "branch"
	projectLinkGitRefTypeTag    projectLinkGitRefType = "tag"
)

const (
	projectLinkPrefixRegistry     = "registry:"
	projectLinkPrefixRegistryAbbr = "@"
	projectLinkPrefixDir          = "dir:"
	projectLinkPrefixGit          = "git:"
	projectLinkGitRefPrefixTag    = "tag/"
	projectLinkGitRefPrefixBranch = "branch/"
	projectLinkRefSeparator       = "#ref="
	projectLinkRefSeparatorLen    = len(projectLinkRefSeparator)
)

type projectLinkTarget struct {
	Link *projectLink
	Path string
	Git  *projectLinkGit
}

func parseProjectLink(rawLink string) (*projectLink, error) {
	var content string
	var matched bool
	if content, matched = strings.CutPrefix(rawLink, projectLinkPrefixRegistry); matched {
		return parseProjectLinkRegistry(rawLink, content)
	} else if content, matched = strings.CutPrefix(rawLink, projectLinkPrefixRegistryAbbr); matched {
		return parseProjectLinkRegistry(rawLink, content)
	} else if content, matched = strings.CutPrefix(rawLink, projectLinkPrefixDir); matched {
		return parseProjectLinkDir(rawLink, content)
	} else if content, matched = strings.CutPrefix(rawLink, projectLinkPrefixGit); matched {
		return parseProjectLinkGit(rawLink, content)
	} else {
		return nil, errN("parse project link error",
			reason("unsupported link"),
			kv("rawLink", rawLink),
		)
	}
}

func parseProjectLinkRegistry(rawLink string, content string) (link *projectLink, err error) {
	name, path, rawRef := "", "", ""
	slashIndex := strings.Index(content, "/")
	if slashIndex < 0 {
		name = content
	} else {
		name = content[:slashIndex]
		path = content[slashIndex+1:]
	}
	if path != "" {
		refIndex := strings.Index(path, projectLinkRefSeparator)
		if refIndex >= 0 {
			rawRef = path[refIndex+projectLinkRefSeparatorLen:]
			path = path[:refIndex]
		} else {
			rawRef = "main"
		}
	} else if name != "" {
		refIndex := strings.Index(name, projectLinkRefSeparator)
		if refIndex >= 0 {
			rawRef = name[refIndex+projectLinkRefSeparatorLen:]
			name = name[:refIndex]
		} else {
			rawRef = "main"
		}
	}
	if name == "" {
		return nil, errN("parse project link error",
			reason("name is empty"),
			kv("rawLink", rawLink),
		)
	}
	if rawRef == "" {
		return nil, errN("parse project link error",
			reason("ref is empty"),
			kv("rawLink", rawLink),
		)
	}
	parsedRef, err := parseProjectLinkGitRef(rawRef)
	if err != nil {
		return nil, errW(err, "parse project link error",
			reason("parse ref error"),
			kv("rawLink", rawLink),
			kv("rawRef", rawRef),
		)
	}
	normalizeLink := projectLinkPrefixRegistry + name
	if path != "" {
		normalizeLink += "/" + path
	}
	normalizeLink += projectLinkRefSeparator + parsedRef.Normalized
	link = &projectLink{
		Raw:        rawLink,
		Normalized: normalizeLink,
		Type:       ProjectLinkTypeRegistry,
		Registry: &projectLinkRegistry{
			Name: name,
			Path: path,
			Ref:  parsedRef.Normalized,
			ref:  parsedRef,
		},
	}
	return link, nil
}

func parseProjectLinkDir(rawLink string, content string) (link *projectLink, err error) {
	dir := content
	if dir == "" {
		return nil, errN("parse project link error",
			reason("dir is empty"),
			kv("rawLink", rawLink),
		)
	}
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return nil, errW(err, "parse project link error",
			reason("get abs-path error"),
			kv("rawLink", rawLink),
			kv("dir", dir),
		)
	}
	normalizeLink := projectLinkPrefixDir + absPath
	link = &projectLink{
		Raw:        rawLink,
		Normalized: normalizeLink,
		Type:       ProjectLinkTypeDir,
		Dir: &projectLinkDir{
			Raw:  dir,
			Path: absPath,
		},
	}
	return link, nil
}

func parseProjectLinkGit(rawLink string, content string) (link *projectLink, err error) {
	rawUrl, rawRef := content, ""
	if rawUrl != "" {
		refIndex := strings.Index(rawUrl, projectLinkRefSeparator)
		if refIndex >= 0 {
			rawRef = rawUrl[refIndex+projectLinkRefSeparatorLen:]
			rawUrl = rawUrl[:refIndex]
		} else {
			rawRef = "main"
		}
	}
	if rawUrl == "" {
		return nil, errN("parse project link error",
			reason("url is empty"),
			kv("rawLink", rawLink),
		)
	}
	if rawRef == "" {
		return nil, errN("parse project link error",
			reason("ref is empty"),
			kv("rawLink", rawLink),
		)
	}
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return nil, errW(err, "parse project link error",
			reason("parse url error"),
			kv("rawLink", rawLink),
			kv("rawUrl", rawUrl),
		)
	}
	parsedRef, err := parseProjectLinkGitRef(rawRef)
	if err != nil {
		return nil, errW(err, "parse project link error",
			reason("parse ref error"),
			kv("rawLink", rawLink),
			kv("rawRef", rawRef),
		)
	}
	normalizeLink := projectLinkPrefixGit + rawUrl + projectLinkRefSeparator + parsedRef.Normalized
	link = &projectLink{
		Raw:        rawLink,
		Normalized: normalizeLink,
		Type:       ProjectLinkTypeGit,
		Git: &projectLinkGit{
			Url:       rawUrl,
			Ref:       parsedRef.Normalized,
			parsedUrl: parsedUrl,
			parsedRef: parsedRef,
		},
	}
	return link, nil
}

func parseProjectLinkGitRef(rawRef string) (ref *projectLinkGitRef, err error) {
	if rawRef == "" {
		return nil, errN("parse project link git ref error",
			reason("ref is empty"),
		)
	}
	var name string
	var matched bool
	if name, matched = strings.CutPrefix(rawRef, projectLinkGitRefPrefixTag); matched {
		if name == "" {
			return nil, errN("parse project link git ref error",
				reason("tag name is empty"),
				kv("rawRef", rawRef),
			)
		}
		ref = &projectLinkGitRef{
			Raw:           rawRef,
			Normalized:    projectLinkGitRefPrefixTag + name,
			Type:          projectLinkGitRefTypeTag,
			Name:          name,
			ReferenceName: plumbing.NewTagReferenceName(name),
		}
	} else if name, matched = strings.CutPrefix(rawRef, projectLinkGitRefPrefixBranch); matched {
		if name == "" {
			return nil, errN("parse project link git ref error",
				reason("branch name is empty"),
				kv("rawRef", rawRef),
			)
		}
		ref = &projectLinkGitRef{
			Raw:           rawRef,
			Normalized:    projectLinkGitRefPrefixBranch + name,
			Type:          projectLinkGitRefTypeBranch,
			Name:          name,
			ReferenceName: plumbing.NewBranchReferenceName(name),
		}
	} else {
		name = rawRef
		ref = &projectLinkGitRef{
			Raw:           rawRef,
			Normalized:    projectLinkGitRefPrefixBranch + name,
			Type:          projectLinkGitRefTypeBranch,
			Name:          name,
			ReferenceName: plumbing.NewBranchReferenceName(name),
		}
	}

	return ref, nil
}
