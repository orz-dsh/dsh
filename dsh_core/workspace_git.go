package dsh_core

import (
	"github.com/go-git/go-git/v5"
	"github.com/pkg/errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (w *Workspace) getGitProjectPath(parsedUrl *url.URL, parsedRef *projectLinkGitRef) string {
	path1 := strings.ReplaceAll(parsedUrl.Host, ":", "@")
	postfix := strings.ReplaceAll(parsedRef.Normalized, "/", "-")
	path2 := strings.ReplaceAll(strings.TrimSuffix(strings.TrimPrefix(parsedUrl.Path, "/"), ".git"), "/", "@") + "@" + postfix
	path := filepath.Join(w.path, "project", path1, path2)
	return path
}

func (w *Workspace) downloadGitProject(path string, rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *projectLinkGitRef) (err error) {
	if err = os.MkdirAll(path, os.ModePerm); err != nil {
		return errW(err, "download git project error",
			reason("make dir error"),
			kv("url", rawUrl),
			kv("ref", rawRef),
			kv("path", path),
		)
	}
	repo, err := git.PlainOpen(path)
	if errors.Is(err, git.ErrRepositoryNotExists) {
		startTime := time.Now()
		w.logger.InfoDesc("download git project start",
			kv("action", "clone project"),
			kv("path", path),
			kv("url", rawUrl),
			kv("ref", rawRef),
		)
		cloneOptions := &git.CloneOptions{
			URL:           rawUrl,
			ReferenceName: parsedRef.ReferenceName,
			SingleBranch:  true,
			Depth:         1,
		}
		if w.logger.IsDebugEnabled() {
			cloneOptions.Progress = w.logger.GetDebugWriter()
		}
		repo, err = git.PlainClone(path, false, cloneOptions)
		if err != nil {
			return errW(err, "download git project error",
				reason("clone repository error"),
				kv("url", rawUrl),
				kv("ref", rawRef),
				kv("path", path),
			)
		}
		w.logger.InfoDesc("download git project finish",
			kv("action", "clone project"),
			kv("elapsed", time.Since(startTime)),
		)
	} else if err != nil {
		return errW(err, "download git project error",
			reason("open repository error"),
			kv("url", rawUrl),
			kv("ref", rawRef),
			kv("path", path),
		)
	} else {
		startTime := time.Now()
		w.logger.InfoDesc("download git project start",
			kv("action", "pull project"),
			kv("path", path),
			kv("url", rawUrl),
			kv("ref", rawRef),
		)
		worktree, err := repo.Worktree()
		if err != nil {
			return errW(err, "download git project error",
				reason("get worktree error"),
				kv("url", rawUrl),
				kv("ref", rawRef),
				kv("path", path),
			)
		}
		err = worktree.Reset(&git.ResetOptions{
			Mode: git.HardReset,
		})
		if err != nil {
			return errW(err, "download git project error",
				reason("reset worktree error"),
				kv("url", rawUrl),
				kv("ref", rawRef),
				kv("path", path),
			)
		}
		err = worktree.Clean(&git.CleanOptions{
			Dir: true,
		})
		if err != nil {
			return errW(err, "download git project error",
				reason("clean worktree error"),
				kv("url", rawUrl),
				kv("ref", rawRef),
				kv("path", path),
			)
		}
		pullOptions := &git.PullOptions{
			ReferenceName: parsedRef.ReferenceName,
			SingleBranch:  true,
			Depth:         1,
		}
		if w.logger.IsDebugEnabled() {
			pullOptions.Progress = w.logger.GetDebugWriter()
		}
		err = worktree.Pull(pullOptions)
		if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return errW(err, "download git project error",
				reason("pull worktree error"),
				kv("url", rawUrl),
				kv("ref", rawRef),
				kv("path", path),
			)
		}
		w.logger.InfoDesc("download git project finish",
			kv("action", "pull project"),
			kv("elapsed", time.Since(startTime)),
		)
	}
	return nil
}
