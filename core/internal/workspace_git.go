package internal

import (
	"errors"
	"github.com/go-git/go-git/v5"
	"github.com/orz-dsh/dsh/core/common"
	. "github.com/orz-dsh/dsh/utils"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (w *WorkspaceCore) GetGitProjectDir(parsedUrl *url.URL, parsedRef *common.ProjectLinkGitRef) string {
	path1 := strings.ReplaceAll(parsedUrl.Host, ":", "@")
	postfix := string(parsedRef.Type) + "-" + parsedRef.Name
	path2 := strings.ReplaceAll(strings.TrimSuffix(strings.TrimPrefix(parsedUrl.Path, "/"), ".git"), "/", "@") + "@" + postfix
	return filepath.Join(w.Dir, "project", path1, path2)
}

func (w *WorkspaceCore) DownloadGitProject(path string, rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *common.ProjectLinkGitRef) (err error) {
	if err = os.MkdirAll(path, os.ModePerm); err != nil {
		return ErrW(err, "download git project error",
			Reason("make dir error"),
			KV("url", rawUrl),
			KV("ref", rawRef),
			KV("path", path),
		)
	}
	repo, err := git.PlainOpen(path)
	if errors.Is(err, git.ErrRepositoryNotExists) {
		startTime := time.Now()
		w.Logger.InfoDesc("download git project start",
			KV("action", "clone project"),
			KV("path", path),
			KV("url", rawUrl),
			KV("ref", rawRef),
		)
		cloneOptions := &git.CloneOptions{
			URL:           rawUrl,
			ReferenceName: parsedRef.ReferenceName,
			SingleBranch:  true,
			Depth:         1,
		}
		if w.Logger.IsDebugEnabled() {
			cloneOptions.Progress = w.Logger.GetDebugWriter()
		}
		repo, err = git.PlainClone(path, false, cloneOptions)
		if err != nil {
			return ErrW(err, "download git project error",
				Reason("clone repository error"),
				KV("url", rawUrl),
				KV("ref", rawRef),
				KV("path", path),
			)
		}
		w.Logger.InfoDesc("download git project finish",
			KV("action", "clone project"),
			KV("elapsed", time.Since(startTime)),
		)
	} else if err != nil {
		return ErrW(err, "download git project error",
			Reason("open repository error"),
			KV("url", rawUrl),
			KV("ref", rawRef),
			KV("path", path),
		)
	} else {
		startTime := time.Now()
		w.Logger.InfoDesc("download git project start",
			KV("action", "pull project"),
			KV("path", path),
			KV("url", rawUrl),
			KV("ref", rawRef),
		)
		worktree, err := repo.Worktree()
		if err != nil {
			return ErrW(err, "download git project error",
				Reason("get worktree error"),
				KV("url", rawUrl),
				KV("ref", rawRef),
				KV("path", path),
			)
		}
		err = worktree.Reset(&git.ResetOptions{
			Mode: git.HardReset,
		})
		if err != nil {
			return ErrW(err, "download git project error",
				Reason("reset worktree error"),
				KV("url", rawUrl),
				KV("ref", rawRef),
				KV("path", path),
			)
		}
		err = worktree.Clean(&git.CleanOptions{
			Dir: true,
		})
		if err != nil {
			return ErrW(err, "download git project error",
				Reason("clean worktree error"),
				KV("url", rawUrl),
				KV("ref", rawRef),
				KV("path", path),
			)
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
			return ErrW(err, "download git project error",
				Reason("pull worktree error"),
				KV("url", rawUrl),
				KV("ref", rawRef),
				KV("path", path),
			)
		}
		w.Logger.InfoDesc("download git project finish",
			KV("action", "pull project"),
			KV("elapsed", time.Since(startTime)),
		)
	}
	return nil
}
