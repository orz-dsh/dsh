package internal

import (
	"fmt"
	. "github.com/orz-dsh/dsh/core/common"
	. "github.com/orz-dsh/dsh/utils"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

func (w *WorkspaceCore) Clean(options WorkspaceCleanOptions) error {
	return w.CleanOutputDir(options.ExcludeOutputDir)
}

func (w *WorkspaceCore) BuildOutputDirName(projectName string) (dirName string, err error) {
	random, err := RandomString(8)
	if err != nil {
		return "", ErrW(err, "build output dir name error",
			Reason("random string error"),
			KV("projectName", projectName),
		)
	}
	createTime := time.Now().Format("060102150405")
	dirName = fmt.Sprintf("%s-%s-%s", projectName, createTime, random)
	return dirName, nil
}

func (w *WorkspaceCore) ParseOutputDirName(dirName string) (projectName string, createTime time.Time, err error) {
	fields := strings.Split(dirName, "-")
	if len(fields) < 2 {
		return "", time.Time{}, ErrN("parse output dir name error",
			Reason("invalid format"),
			KV("dirName", dirName),
		)
	}
	projectName = fields[0]
	if projectName == "" {
		return "", time.Time{}, ErrN("parse output dir name error",
			Reason("project dirName empty"),
			KV("dirName", dirName),
		)
	}
	createTime, err = time.ParseInLocation("060102150405", fields[1], time.Local)
	if err != nil {
		return "", time.Time{}, ErrW(err, "parse output dir name error",
			Reason("parse create time error"),
			KV("dirName", dirName),
		)
	}
	return projectName, createTime, nil
}

func (w *WorkspaceCore) MakeOutputDir(projectName string) (string, error) {
	for i := 0; i < 10; i++ {
		name, err := w.BuildOutputDirName(projectName)
		if err != nil {
			return "", ErrW(err, "make output dir error",
				Reason("build dir name error"),
				KV("projectName", projectName),
			)
		}
		path := filepath.Join(w.Dir, "output", name)
		if !IsDirExists(path) {
			if err = os.MkdirAll(path, os.ModePerm); err != nil {
				return "", ErrW(err, "make output dir error",
					Reason("make dir error"),
					KV("projectName", projectName),
					KV("path", path),
				)
			}
			return path, nil
		}
	}
	return "", ErrN("make output path error",
		Reason("retry too many times"),
		KV("projectName", projectName),
		KV("workspaceDir", w.Dir),
	)
}

func (w *WorkspaceCore) CleanOutputDir(excludeOutputPath string) error {
	outputPath := filepath.Join(w.Dir, "output")
	if !IsDirExists(outputPath) {
		return nil
	}
	dirNames, err := ListChildDirs(outputPath)
	if err != nil {
		w.Logger.Warn("cleanup workspace output dir error",
			Reason("list child dirs error"),
			KV("outputPath", outputPath),
		)
		return ErrW(err, "cleanup workspace dir error",
			Reason("list child dirs error"),
			KV("outputPath", outputPath),
		)
	}
	slices.Sort(dirNames)
	slices.Reverse(dirNames)

	var errorDirNames []string
	var removeDirNames []string
	var projectCounts = map[string]int{}
	now := time.Now()
	for i := 0; i < len(dirNames); i++ {
		dirName := dirNames[i]
		projectName, createTime, err := w.ParseOutputDirName(dirName)
		if err != nil {
			w.Logger.Warn("cleanup workspace output dir error",
				Reason("parse dir name error"),
				KV("dirName", dirName),
			)
			errorDirNames = append(errorDirNames, dirName)
			continue
		}
		projectCount, exist := projectCounts[projectName]
		if exist {
			projectCount = projectCount + 1
		} else {
			projectCount = 1
		}
		if projectCount > w.Setting.Clean.OutputCount {
			removeDirNames = append(removeDirNames, dirName)
		} else if now.Sub(createTime) > w.Setting.Clean.OutputExpires {
			removeDirNames = append(removeDirNames, dirName)
		}
		projectCounts[projectName] = projectCount
	}
	for i := 0; i < len(removeDirNames); i++ {
		dirName := removeDirNames[i]
		dirPath := filepath.Join(outputPath, dirName)
		if dirPath == excludeOutputPath {
			continue
		}
		if err = os.RemoveAll(dirPath); err != nil {
			w.Logger.Warn("cleanup workspace output dir error",
				Reason("remove dir error"),
				KV("dirPath", dirPath),
			)
			errorDirNames = append(errorDirNames, dirName)
		} else {
			w.Logger.DebugDesc("cleanup workspace output dir",
				KV("dirPath", dirPath),
			)
		}
	}
	if len(errorDirNames) > 0 {
		return ErrN("cleanup workspace error",
			Reason("remove dirs error"),
			KV("outputPath", outputPath),
			KV("errorDirNames", errorDirNames),
		)
	}
	return nil
}
