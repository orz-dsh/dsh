package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

func (w *Workspace) buildOutputDirName(projectName string) (dirName string, err error) {
	random, err := dsh_utils.RandomString(8)
	if err != nil {
		return "", errW(err, "build output dir name error",
			reason("random string error"),
			kv("projectName", projectName),
		)
	}
	createTime := time.Now().Format("060102150405")
	dirName = fmt.Sprintf("%s-%s-%s", projectName, createTime, random)
	return dirName, nil
}

func (w *Workspace) parseOutputDirName(dirName string) (projectName string, createTime time.Time, err error) {
	fields := strings.Split(dirName, "-")
	if len(fields) < 2 {
		return "", time.Time{}, errN("parse output dir name error",
			reason("invalid format"),
			kv("dirName", dirName),
		)
	}
	projectName = fields[0]
	if projectName == "" {
		return "", time.Time{}, errN("parse output dir name error",
			reason("project dirName empty"),
			kv("dirName", dirName),
		)
	}
	createTime, err = time.ParseInLocation("060102150405", fields[1], time.Local)
	if err != nil {
		return "", time.Time{}, errW(err, "parse output dir name error",
			reason("parse create time error"),
			kv("dirName", dirName),
		)
	}
	return projectName, createTime, nil
}

func (w *Workspace) makeOutputDir(projectName string) (string, error) {
	for i := 0; i < 10; i++ {
		name, err := w.buildOutputDirName(projectName)
		if err != nil {
			return "", errW(err, "make output dir error",
				reason("build dir name error"),
				kv("projectName", projectName),
			)
		}
		path := filepath.Join(w.path, "output", name)
		if !dsh_utils.IsDirExists(path) {
			if err = os.MkdirAll(path, os.ModePerm); err != nil {
				return "", errW(err, "make output dir error",
					reason("make dir error"),
					kv("projectName", projectName),
					kv("path", path),
				)
			}
			return path, nil
		}
	}
	return "", errN("make output path error",
		reason("retry too many times"),
		kv("projectName", projectName),
		kv("workspacePath", w.path),
	)
}

func (w *Workspace) cleanOutputDir(excludeOutputPath string) error {
	outputPath := filepath.Join(w.path, "output")
	if !dsh_utils.IsDirExists(outputPath) {
		return nil
	}
	dirNames, err := dsh_utils.ListChildDirs(outputPath)
	if err != nil {
		w.logger.Warn("cleanup workspace output dir error",
			reason("list child dirs error"),
			kv("outputPath", outputPath),
		)
		return errW(err, "cleanup workspace dir error",
			reason("list child dirs error"),
			kv("outputPath", outputPath),
		)
	}
	slices.Sort(dirNames)
	slices.Reverse(dirNames)

	var errorDirNames []string
	var removeDirNames []string
	var projectCounts = make(map[string]int)
	now := time.Now()
	for i := 0; i < len(dirNames); i++ {
		dirName := dirNames[i]
		projectName, createTime, err := w.parseOutputDirName(dirName)
		if err != nil {
			w.logger.Warn("cleanup workspace output dir error",
				reason("parse dir name error"),
				kv("dirName", dirName),
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
		if projectCount > w.manifest.Clean.Output.count {
			removeDirNames = append(removeDirNames, dirName)
		} else if now.Sub(createTime) > w.manifest.Clean.Output.expires {
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
			w.logger.Warn("cleanup workspace output dir error",
				reason("remove dir error"),
				kv("dirPath", dirPath),
			)
			errorDirNames = append(errorDirNames, dirName)
		} else {
			w.logger.DebugDesc("cleanup workspace output dir",
				kv("dirPath", dirPath),
			)
		}
	}
	if len(errorDirNames) > 0 {
		return errN("cleanup workspace error",
			reason("remove dirs error"),
			kv("outputPath", outputPath),
			kv("errorDirNames", errorDirNames),
		)
	}
	return nil
}
