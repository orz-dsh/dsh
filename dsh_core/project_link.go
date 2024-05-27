package dsh_core

import (
	"github.com/go-git/go-git/v5/plumbing"
	"net/url"
	"strings"
)

type ProjectLink struct {
	Raw        string
	Normalized string
	Type       ProjectLinkType
	Registry   *ProjectLinkRegistry
	Dir        *ProjectLinkDir
	Git        *ProjectLinkGit
}

type ProjectLinkRegistry struct {
	Name string
	Path string
	Ref  string
	ref  *ProjectLinkGitRef
}

type ProjectLinkDir struct {
	Dir string
}

type ProjectLinkGit struct {
	Url       string
	Ref       string
	parsedUrl *url.URL
	parsedRef *ProjectLinkGitRef
}

type ProjectLinkGitRef struct {
	Raw           string
	Normalized    string
	Type          ProjectLinkGitRefType
	Name          string
	ReferenceName plumbing.ReferenceName
}

type ProjectLinkType string

const (
	ProjectLinkTypeRegistry ProjectLinkType = "registry"
	ProjectLinkTypeDir      ProjectLinkType = "dir"
	ProjectLinkTypeGit      ProjectLinkType = "git"
)

type ProjectLinkGitRefType string

const (
	ProjectLinkGitRefTypeBranch ProjectLinkGitRefType = "branch"
	ProjectLinkGitRefTypeTag    ProjectLinkGitRefType = "tag"
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

type projectResolvedLink struct {
	Link *ProjectLink
	Path string
	Git  *ProjectLinkGit
}

func ParseProjectLink(rawLink string) (*ProjectLink, error) {
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

func parseProjectLinkRegistry(rawLink string, content string) (link *ProjectLink, err error) {
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
	link = &ProjectLink{
		Raw:        rawLink,
		Normalized: normalizeLink,
		Type:       ProjectLinkTypeRegistry,
		Registry: &ProjectLinkRegistry{
			Name: name,
			Path: path,
			Ref:  parsedRef.Normalized,
			ref:  parsedRef,
		},
	}
	return link, nil
}

func parseProjectLinkDir(rawLink string, content string) (link *ProjectLink, err error) {
	dir := content
	if dir == "" {
		return nil, errN("parse project link error",
			reason("dir is empty"),
			kv("rawLink", rawLink),
		)
	}
	normalizeLink := projectLinkPrefixDir + dir
	link = &ProjectLink{
		Raw:        rawLink,
		Normalized: normalizeLink,
		Type:       ProjectLinkTypeDir,
		Dir: &ProjectLinkDir{
			Dir: dir,
		},
	}
	return link, nil
}

func parseProjectLinkGit(rawLink string, content string) (link *ProjectLink, err error) {
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
	link = &ProjectLink{
		Raw:        rawLink,
		Normalized: normalizeLink,
		Type:       ProjectLinkTypeGit,
		Git: &ProjectLinkGit{
			Url:       rawUrl,
			Ref:       parsedRef.Normalized,
			parsedUrl: parsedUrl,
			parsedRef: parsedRef,
		},
	}
	return link, nil
}

func parseProjectLinkGitRef(rawRef string) (ref *ProjectLinkGitRef, err error) {
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
		ref = &ProjectLinkGitRef{
			Raw:           rawRef,
			Normalized:    projectLinkGitRefPrefixTag + name,
			Type:          ProjectLinkGitRefTypeTag,
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
		ref = &ProjectLinkGitRef{
			Raw:           rawRef,
			Normalized:    projectLinkGitRefPrefixBranch + name,
			Type:          ProjectLinkGitRefTypeBranch,
			Name:          name,
			ReferenceName: plumbing.NewBranchReferenceName(name),
		}
	} else {
		name = rawRef
		ref = &ProjectLinkGitRef{
			Raw:           rawRef,
			Normalized:    projectLinkGitRefPrefixBranch + name,
			Type:          ProjectLinkGitRefTypeBranch,
			Name:          name,
			ReferenceName: plumbing.NewBranchReferenceName(name),
		}
	}

	return ref, nil
}
