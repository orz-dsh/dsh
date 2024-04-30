package dsh_utils

import (
	"encoding/json"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type FileType string

const (
	FileTypeYaml        FileType = "yaml"
	FileTypeToml        FileType = "toml"
	FileTypeJson        FileType = "json"
	FileTypeTemplate    FileType = "template"
	FileTypeTemplateLib FileType = "template-lib"
	FileTypePlain       FileType = "plain"
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
		return WrapError(err, "dir remove failed", map[string]any{
			"path": path,
		})
	}
	if err = os.MkdirAll(path, os.ModePerm); err != nil {
		return WrapError(err, "dir make failed", map[string]any{
			"path": path,
		})
	}
	return nil
}

func LinkFile(sourcePath string, targetPath string) (err error) {
	if err = os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
		return WrapError(err, "dir make failed", map[string]any{
			"path": targetPath,
		})
	}
	return os.Link(sourcePath, targetPath)
}

func CopyFile(sourcePath string, targetPath string) (err error) {
	if err = os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
		return WrapError(err, "dir make failed", map[string]any{
			"path": targetPath,
		})
	}

	targetFile, err := os.Create(targetPath)
	if err != nil {
		return WrapError(err, "file create failed", map[string]any{
			"path": targetPath,
		})
	}
	defer targetFile.Close()

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return WrapError(err, "file open failed", map[string]any{
			"path": sourcePath,
		})
	}
	defer sourceFile.Close()

	_, err = io.Copy(targetFile, sourceFile)
	if err != nil {
		return WrapError(err, "file copy failed", map[string]any{
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
			return WrapError(err, "link or copy failed", map[string]any{
				"sourcePath": sourcePath,
				"targetPath": targetPath,
			})
		}
	}
	return nil
}

func ReadYamlFile(path string, model any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return WrapError(err, "file read failed", map[string]any{
			"path": path,
		})
	}
	err = yaml.Unmarshal(data, model)
	if err != nil {
		return WrapError(err, "yaml unmarshal failed", map[string]any{
			"path": path,
		})
	}
	return nil
}

func ReadTomlFile(path string, model any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return WrapError(err, "file read failed", map[string]any{
			"path": path,
		})
	}
	err = toml.Unmarshal(data, model)
	if err != nil {
		return WrapError(err, "toml unmarshal failed", map[string]any{
			"path": path,
		})
	}
	return nil
}

func ReadJsonFile(path string, model any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return WrapError(err, "file read failed", map[string]any{
			"path": path,
		})
	}
	err = json.Unmarshal(data, model)
	if err != nil {
		return WrapError(err, "json unmarshal failed", map[string]any{
			"path": path,
		})
	}
	return nil
}

func IsYamlFile(path string) bool {
	return strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml")
}

func IsTomlFile(path string) bool {
	return strings.HasSuffix(path, ".toml")
}

func IsJsonFile(path string) bool {
	return strings.HasSuffix(path, ".json")
}

func IsTemplateFile(path string) bool {
	return strings.HasSuffix(path, ".dtpl")
}

func IsTemplateLibFile(path string) bool {
	return strings.HasSuffix(path, ".dtpl.lib")
}

func GetFileType(path string, fileTypes []FileType) FileType {
	includePlain := false
	for i := 0; i < len(fileTypes); i++ {
		switch fileTypes[i] {
		case FileTypeYaml:
			if IsYamlFile(path) {
				return FileTypeYaml
			}
		case FileTypeToml:
			if IsTomlFile(path) {
				return FileTypeToml
			}
		case FileTypeJson:
			if IsJsonFile(path) {
				return FileTypeJson
			}
		case FileTypeTemplate:
			if IsTemplateFile(path) {
				return FileTypeTemplate
			}
		case FileTypeTemplateLib:
			if IsTemplateLibFile(path) {
				return FileTypeTemplateLib
			}
		case FileTypePlain:
			includePlain = true
		}
	}
	if includePlain {
		return FileTypePlain
	}
	return ""
}

func RemoveFileExt(path string) string {
	ext := filepath.Ext(path)
	if ext == "" {
		return path
	}
	return path[:len(path)-len(ext)]
}

func SelectFile(sourceDir string, files []string, fileTypes []FileType) (string, FileType) {
	for i := 0; i < len(files); i++ {
		path := filepath.Join(sourceDir, files[i])
		if IsFileExists(path) {
			return path, GetFileType(path, fileTypes)
		}
	}
	return "", ""
}

func ScanFiles(sourceDir string, includeFiles []string, includeFileTypes []FileType) (filePaths []string, fileTypes []FileType, err error) {
	var includeFileMap = make(map[string]bool)
	for i := 0; i < len(includeFiles); i++ {
		includeFileMap[filepath.Join(sourceDir, includeFiles[i])] = true
	}
	err = filepath.WalkDir(sourceDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return WrapError(err, "dir walk failed", map[string]any{
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
				return WrapError(err, "file rel-path get failed", map[string]any{
					"dir":  sourceDir,
					"path": path,
				})
			}
			fileType := GetFileType(relPath, includeFileTypes)
			if fileType != "" {
				filePaths = append(filePaths, relPath)
				fileTypes = append(fileTypes, fileType)
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, WrapError(err, "files scan failed", map[string]any{
			"sourceDir": sourceDir,
		})
	}
	return filePaths, fileTypes, nil
}

func WriteTemplate(t *template.Template, env any, targetPath string) (err error) {
	if err = os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
		return WrapError(err, "dir make failed", map[string]any{
			"path": targetPath,
		})
	}

	targetFile, err := os.Create(targetPath)
	if err != nil {
		return WrapError(err, "file create failed", map[string]any{
			"path": targetPath,
		})
	}
	defer targetFile.Close()

	err = t.Execute(targetFile, env)
	if err != nil {
		return WrapError(err, "template execute failed", map[string]any{
			"targetPath": targetPath,
		})
	}
	return nil
}
