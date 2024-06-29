package common

import (
	"github.com/go-git/go-git/v5/plumbing"
	. "github.com/orz-dsh/dsh/utils"
	"net/url"
	"path/filepath"
	"regexp"
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
	Name      string
	Path      string
	Ref       string
	ParsedRef *ProjectLinkGitRef
}

type ProjectLinkDir struct {
	Raw  string
	Path string
}

type ProjectLinkGit struct {
	Url       string
	Ref       string
	ParsedUrl *url.URL
	ParsedRef *ProjectLinkGitRef
}

type ProjectLinkGitRef struct {
	Raw           string
	Normalized    string
	Type          ProjectLinkGitRefType
	Name          string
	ReferenceName plumbing.ReferenceName
}

type ProjectLinkTarget struct {
	Link *ProjectLink
	Dir  string
	Git  *ProjectLinkGit
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

var projectLinkRegistryNameCheckRegex = regexp.MustCompile("^[a-z][a-z0-9-]*[a-z0-9]$")

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
		return nil, ErrN("parse project link error",
			Reason("unsupported link"),
			KV("rawLink", rawLink),
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
		return nil, ErrN("parse project link error",
			Reason("name is empty"),
			KV("rawLink", rawLink),
		)
	}
	if !projectLinkRegistryNameCheckRegex.MatchString(name) {
		return nil, ErrN("parse project link error",
			Reason("name is invalid"),
			KV("rawLink", rawLink),
			KV("name", name),
		)
	}
	if rawRef == "" {
		return nil, ErrN("parse project link error",
			Reason("ref is empty"),
			KV("rawLink", rawLink),
		)
	}
	parsedRef, err := ParseProjectLinkGitRef(rawRef)
	if err != nil {
		return nil, ErrW(err, "parse project link error",
			Reason("parse ref error"),
			KV("rawLink", rawLink),
			KV("rawRef", rawRef),
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
			Name:      name,
			Path:      path,
			Ref:       parsedRef.Normalized,
			ParsedRef: parsedRef,
		},
	}
	return link, nil
}

func parseProjectLinkDir(rawLink string, content string) (link *ProjectLink, err error) {
	dir := content
	if dir == "" {
		return nil, ErrN("parse project link error",
			Reason("dir is empty"),
			KV("rawLink", rawLink),
		)
	}
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return nil, ErrW(err, "parse project link error",
			Reason("get abs-path error"),
			KV("rawLink", rawLink),
			KV("dir", dir),
		)
	}
	normalizeLink := projectLinkPrefixDir + absPath
	link = &ProjectLink{
		Raw:        rawLink,
		Normalized: normalizeLink,
		Type:       ProjectLinkTypeDir,
		Dir: &ProjectLinkDir{
			Raw:  dir,
			Path: absPath,
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
		return nil, ErrN("parse project link error",
			Reason("url is empty"),
			KV("rawLink", rawLink),
		)
	}
	if rawRef == "" {
		return nil, ErrN("parse project link error",
			Reason("ref is empty"),
			KV("rawLink", rawLink),
		)
	}
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return nil, ErrW(err, "parse project link error",
			Reason("parse url error"),
			KV("rawLink", rawLink),
			KV("rawUrl", rawUrl),
		)
	}
	parsedRef, err := ParseProjectLinkGitRef(rawRef)
	if err != nil {
		return nil, ErrW(err, "parse project link error",
			Reason("parse ref error"),
			KV("rawLink", rawLink),
			KV("rawRef", rawRef),
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
			ParsedUrl: parsedUrl,
			ParsedRef: parsedRef,
		},
	}
	return link, nil
}

func ParseProjectLinkGitRef(rawRef string) (ref *ProjectLinkGitRef, err error) {
	if rawRef == "" {
		return nil, ErrN("parse project link git ref error",
			Reason("ref is empty"),
		)
	}
	var name string
	var matched bool
	if name, matched = strings.CutPrefix(rawRef, projectLinkGitRefPrefixTag); matched {
		if name == "" {
			return nil, ErrN("parse project link git ref error",
				Reason("tag name is empty"),
				KV("rawRef", rawRef),
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
			return nil, ErrN("parse project link git ref error",
				Reason("branch name is empty"),
				KV("rawRef", rawRef),
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
