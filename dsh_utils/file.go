package dsh_utils

import (
	"encoding/json"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
	"strings"
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
		return errW(err, "remake dir error",
			reason("remove dir error"),
			kv("path", path),
		)
	}
	if err = os.MkdirAll(path, os.ModePerm); err != nil {
		return errW(err, "remake dir error",
			reason("make dir error"),
			kv("path", path),
		)
	}
	return nil
}

func LinkFile(sourcePath string, targetPath string) (err error) {
	if err = os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
		return errW(err, "link file error",
			reason("make dir error"),
			kv("path", targetPath),
		)
	}
	return os.Link(sourcePath, targetPath)
}

func CopyFile(sourcePath string, targetPath string) (err error) {
	if err = os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
		return errW(err, "copy file error",
			reason("make dir error"),
			kv("path", targetPath),
		)
	}

	targetFile, err := os.Create(targetPath)
	if err != nil {
		return errW(err, "copy file error",
			reason("create target file error"),
			kv("path", targetPath),
		)
	}
	defer targetFile.Close()

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return errW(err, "copy file error",
			reason("open source file error"),
			kv("path", sourcePath),
		)
	}
	defer sourceFile.Close()

	_, err = io.Copy(targetFile, sourceFile)
	if err != nil {
		return errW(err, "copy file error",
			reason("io copy error"),
			kv("targetFile", targetFile),
			kv("sourceFile", sourceFile),
		)
	}
	return nil
}

func LinkOrCopyFile(sourcePath string, targetPath string) (err error) {
	err = LinkFile(sourcePath, targetPath)
	if err != nil {
		err = CopyFile(sourcePath, targetPath)
		if err != nil {
			return errW(err, "link or copy file error",
				reason("copy file error"),
				kv("sourcePath", sourcePath),
				kv("targetPath", targetPath),
			)
		}
	}
	return nil
}

func ReadYamlFile(path string, model any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return errW(err, "read yaml file error",
			reason("read file error"),
			kv("path", path),
		)
	}
	err = yaml.Unmarshal(data, model)
	if err != nil {
		return errW(err, "read yaml file error",
			reason("yaml unmarshal error"),
			kv("path", path),
		)
	}
	return nil
}

func ReadTomlFile(path string, model any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return errW(err, "read toml file error",
			reason("read file error"),
			kv("path", path),
		)
	}
	err = toml.Unmarshal(data, model)
	if err != nil {
		return errW(err, "read toml file error",
			reason("toml unmarshal error"),
			kv("path", path),
		)
	}
	return nil
}

func ReadJsonFile(path string, model any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return errW(err, "read json file error",
			reason("read file error"),
			kv("path", path),
		)
	}
	err = json.Unmarshal(data, model)
	if err != nil {
		return errW(err, "read json file error",
			reason("json unmarshal error"),
			kv("path", path),
		)
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

func FindFileName(path string, fileNames []string, fileTypes []FileType) (string, FileType) {
	for i := 0; i < len(fileNames); i++ {
		filePath := filepath.Join(path, fileNames[i])
		if IsFileExists(filePath) {
			return filePath, GetFileType(filePath, fileTypes)
		}
	}
	return "", ""
}

func ScanFiles(sourceDir string, includeFiles []string, includeFileTypes []FileType) (filePaths []string, fileTypes []FileType, err error) {
	var includeFilePathsDict = make(map[string]bool)
	for i := 0; i < len(includeFiles); i++ {
		includeFilePathsDict[filepath.Join(sourceDir, includeFiles[i])] = true
	}
	err = filepath.WalkDir(sourceDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return errW(err, "scan files error",
				reason("walk dir error"),
				kv("dir", sourceDir),
			)
		}
		if !d.IsDir() {
			if len(includeFilePathsDict) > 0 {
				if _, exist := includeFilePathsDict[path]; !exist {
					return nil
				}
			}
			relPath, err := filepath.Rel(sourceDir, path)
			if err != nil {
				return errW(err, "scan files error",
					reason("get rel-path error"),
					kv("dir", sourceDir),
					kv("path", path),
				)
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
		return nil, nil, err
	}
	return filePaths, fileTypes, nil
}
