package dsh_utils

import (
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func IsFileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func IsDirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func RemakeDir(path string) (err error) {
	if err = os.RemoveAll(path); err != nil {
		return WrapError(err, "dir remove failed", map[string]interface{}{
			"path": path,
		})
	}
	if err = os.MkdirAll(path, os.ModePerm); err != nil {
		return WrapError(err, "dir make failed", map[string]interface{}{
			"path": path,
		})
	}
	return nil
}

func LinkFile(sourcePath string, targetPath string) (err error) {
	if err = os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
		return WrapError(err, "dir make failed", map[string]interface{}{
			"path": targetPath,
		})
	}
	return os.Link(sourcePath, targetPath)
}

func CopyFile(sourcePath string, targetPath string) (err error) {
	if err = os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
		return WrapError(err, "dir make failed", map[string]interface{}{
			"path": targetPath,
		})
	}

	targetFile, err := os.Create(targetPath)
	if err != nil {
		return WrapError(err, "file create failed", map[string]interface{}{
			"path": targetPath,
		})
	}
	defer targetFile.Close()

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return WrapError(err, "file open failed", map[string]interface{}{
			"path": sourcePath,
		})
	}
	defer sourceFile.Close()

	_, err = io.Copy(targetFile, sourceFile)
	if err != nil {
		return WrapError(err, "file copy failed", map[string]interface{}{
			"targetFile": targetFile,
			"sourceFile": sourceFile,
		})
	}
	return nil
}

func LinkOrCopyFile(sourcePath string, targetPath string) (err error) {
	err = LinkFile(sourcePath, targetPath)
	if err != nil {
		err = CopyFile(sourcePath, targetPath)
		if err != nil {
			return WrapError(err, "link or copy failed", map[string]interface{}{
				"sourcePath": sourcePath,
				"targetPath": targetPath,
			})
		}
	}
	return nil
}

func ReadYaml(path string, model interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return WrapError(err, "file read failed", map[string]interface{}{
			"path": path,
		})
	}
	err = yaml.Unmarshal(data, model)
	if err != nil {
		return WrapError(err, "yaml unmarshal failed", map[string]interface{}{
			"path": path,
		})
	}
	return nil
}

func IsYaml(path string) bool {
	return strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml")
}

func ScanScriptSources(sourceDir string, includeFiles []string) (plainSourcePaths []string, templateSourcePaths []string, templateLibSourcePaths []string, err error) {
	var includeFileMap = make(map[string]bool)
	for i := 0; i < len(includeFiles); i++ {
		includeFileMap[filepath.Join(sourceDir, includeFiles[i])] = true
	}
	err = filepath.WalkDir(sourceDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return WrapError(err, "dir walk failed", map[string]interface{}{
				"dir": sourceDir,
			})
		}
		if !d.IsDir() {
			if len(includeFileMap) > 0 {
				if _, exist := includeFileMap[path]; !exist {
					return nil
				}
			}
			relPath, err := filepath.Rel(sourceDir, path)
			if err != nil {
				return WrapError(err, "file rel-path get failed", map[string]interface{}{
					"dir":  sourceDir,
					"path": path,
				})
			}
			isTemplate := strings.HasSuffix(relPath, ".dtpl")
			if isTemplate {
				templateSourcePaths = append(templateSourcePaths, relPath)
			} else {
				isTemplateLib := strings.HasSuffix(relPath, ".dtpl.lib")
				if isTemplateLib {
					templateLibSourcePaths = append(templateLibSourcePaths, relPath)
				} else {
					plainSourcePaths = append(plainSourcePaths, relPath)
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, nil, WrapError(err, "script sources scan failed", map[string]interface{}{
			"sourceDir": sourceDir,
		})
	}
	return plainSourcePaths, templateSourcePaths, templateLibSourcePaths, nil
}

func ScanConfigSources(sourceDir string, includeFiles []string) (yamlSourcePaths []string, err error) {
	var includeFileMap = make(map[string]bool)
	for i := 0; i < len(includeFiles); i++ {
		includeFileMap[filepath.Join(sourceDir, includeFiles[i])] = true
	}
	err = filepath.WalkDir(sourceDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return WrapError(err, "dir walk failed", map[string]interface{}{
				"dir": sourceDir,
			})
		}
		if !d.IsDir() {
			if len(includeFileMap) > 0 {
				if _, exist := includeFileMap[path]; !exist {
					return nil
				}
			}
			relPath, err := filepath.Rel(sourceDir, path)
			if err != nil {
				return WrapError(err, "file rel-path get failed", map[string]interface{}{
					"dir":  sourceDir,
					"path": path,
				})
			}
			if IsYaml(relPath) {
				yamlSourcePaths = append(yamlSourcePaths, relPath)
			}
		}
		return nil
	})
	if err != nil {
		return nil, WrapError(err, "config sources scan failed", map[string]interface{}{
			"sourceDir": sourceDir,
		})
	}
	return yamlSourcePaths, nil
}

func WriteTemplate(t *template.Template, env any, targetPath string) (err error) {
	if err = os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
		return WrapError(err, "dir make failed", map[string]interface{}{
			"path": targetPath,
		})
	}

	targetFile, err := os.Create(targetPath)
	if err != nil {
		return WrapError(err, "file create failed", map[string]interface{}{
			"path": targetPath,
		})
	}
	defer targetFile.Close()

	err = t.Execute(targetFile, env)
	if err != nil {
		return WrapError(err, "template execute failed", map[string]interface{}{
			"targetFile": targetFile,
		})
	}
	return nil
}
