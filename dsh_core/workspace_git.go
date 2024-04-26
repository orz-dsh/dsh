package dsh_core

import (
	"dsh/dsh_utils"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pkg/errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type GitRefType string

const (
	GitRefTypeBranch GitRefType = "branch"
	GitRefTypeTag    GitRefType = "tag"
)

type GitRef struct {
	Type          GitRefType
	PathPostfix   string
	ReferenceName plumbing.ReferenceName
}

func ParseGitRef(rawRef string) *GitRef {
	if strings.HasPrefix(rawRef, "tags/") {
		tag := strings.TrimPrefix(rawRef, "tags/")
		return &GitRef{
			Type:          GitRefTypeTag,
			PathPostfix:   "tag-" + tag,
			ReferenceName: plumbing.NewTagReferenceName(tag),
		}
	}
	return &GitRef{
		Type:          GitRefTypeBranch,
		PathPostfix:   "branch-" + rawRef,
		ReferenceName: plumbing.NewBranchReferenceName(rawRef),
	}
}

func (w *Workspace) GetGitProjectPath(parsedUrl *url.URL, parsedRef *GitRef) string {
	path1 := strings.ReplaceAll(parsedUrl.Host, ":", "@")
	path2 := strings.ReplaceAll(strings.TrimSuffix(strings.TrimPrefix(parsedUrl.Path, "/"), ".git"), "/", "@") + "@" + parsedRef.PathPostfix
	projectPath := filepath.Join(w.Path, "project", path1, path2)
	return projectPath
}

func (w *Workspace) DownloadGitProject(projectPath string, rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *GitRef) (err error) {
	if err = os.MkdirAll(projectPath, os.ModePerm); err != nil {
		return dsh_utils.WrapError(err, "git mkdir failed", map[string]interface{}{
			"url":         rawUrl,
			"ref":         rawRef,
			"projectPath": projectPath,
		})
	}
	repo, err := git.PlainOpen(projectPath)
	if errors.Is(err, git.ErrRepositoryNotExists) {
		startTime := time.Now()
		w.Logger.Info("clone project start: path=%s, url=%s, ref=%s", projectPath, rawUrl, rawRef)
		cloneOptions := &git.CloneOptions{
			URL:           rawUrl,
			ReferenceName: parsedRef.ReferenceName,
			SingleBranch:  true,
			Depth:         1,
		}
		if w.Logger.IsDebugEnabled() {
			cloneOptions.Progress = w.Logger.GetDebugWriter()
		}
		repo, err = git.PlainClone(projectPath, false, cloneOptions)
		if err != nil {
			return dsh_utils.WrapError(err, "git clone failed", map[string]interface{}{
				"url":         rawUrl,
				"ref":         rawRef,
				"projectPath": projectPath,
			})
		}
		w.Logger.Info("clone project finish: elapsed=%s", time.Since(startTime))
	} else if err != nil {
		return dsh_utils.WrapError(err, "git open failed", map[string]interface{}{
			"url":         rawUrl,
			"ref":         rawRef,
			"projectPath": projectPath,
		})
	} else {
		startTime := time.Now()
		w.Logger.Info("pull project start: path=%s, url=%s, ref=%s", projectPath, rawUrl, rawRef)
		worktree, err := repo.Worktree()
		if err != nil {
			return dsh_utils.WrapError(err, "git worktree get failed", map[string]interface{}{
				"url":         rawUrl,
				"ref":         rawRef,
				"projectPath": projectPath,
			})
		}
		err = worktree.Reset(&git.ResetOptions{
			Mode: git.HardReset,
		})
		if err != nil {
			return dsh_utils.WrapError(err, "git worktree reset failed", map[string]interface{}{
				"url":         rawUrl,
				"ref":         rawRef,
				"projectPath": projectPath,
			})
		}
		err = worktree.Clean(&git.CleanOptions{
			Dir: true,
		})
		if err != nil {
			return dsh_utils.WrapError(err, "git worktree clean failed", map[string]interface{}{
				"url":         rawUrl,
				"ref":         rawRef,
				"projectPath": projectPath,
			})
		}
		pullOptions := &git.PullOptions{
			ReferenceName: parsedRef.ReferenceName,
			SingleBranch:  true,
			Depth:         1,
		}
		if w.Logger.IsDebugEnabled() {
			pullOptions.Progress = w.Logger.GetDebugWriter()
		}
		err = worktree.Pull(pullOptions)
		if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return dsh_utils.WrapError(err, "git worktree pull failed", map[string]interface{}{
				"url":         rawUrl,
				"ref":         rawRef,
				"projectPath": projectPath,
			})
		}
		w.Logger.Info("pull project finish: elapsed=%s", time.Since(startTime))
	}
	return nil
}
