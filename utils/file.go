package utils

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
	FileTypeConfigYaml  FileType = "config-yaml"
	FileTypeConfigToml  FileType = "config-toml"
	FileTypeConfigJson  FileType = "config-json"
	FileTypeTemplate    FileType = "template"
	FileTypeTemplateLib FileType = "template-lib"
	FileTypeYaml        FileType = "yaml"
	FileTypeToml        FileType = "toml"
	FileTypeJson        FileType = "json"
	FileTypePlain       FileType = "plain"
)

func IsFileExists(file string) bool {
	info, err := os.Stat(file)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func IsDirExists(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func RemakeDir(dir string) (err error) {
	if err = os.RemoveAll(dir); err != nil {
		return errW(err, "remake dir error",
			reason("remove dir error"),
			kv("dir", dir),
		)
	}
	if err = os.MkdirAll(dir, os.ModePerm); err != nil {
		return errW(err, "remake dir error",
			reason("make dir error"),
			kv("dir", dir),
		)
	}
	return nil
}

func ClearDir(dir string) (err error) {
	children, err := os.ReadDir(dir)
	if err != nil {
		return errW(err, "clear dir error",
			reason("read dir error"),
			kv("dir", dir),
		)
	}
	for i := 0; i < len(children); i++ {
		child := filepath.Join(dir, children[i].Name())
		if children[i].IsDir() {
			if err = os.RemoveAll(child); err != nil {
				return errW(err, "clear dir error",
					reason("remove child dir error"),
					kv("child", child),
				)
			}
		} else {
			if err = os.Remove(child); err != nil {
				return errW(err, "clear dir error",
					reason("remove child file error"),
					kv("child", child),
				)
			}
		}
	}
	return nil
}

func LinkFile(sourceFile string, targetFile string) (err error) {
	targetDir := filepath.Dir(targetFile)
	if err = os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return errW(err, "link file error",
			reason("make target dir error"),
			kv("targetDir", targetDir),
		)
	}
	return os.Link(sourceFile, targetFile)
}

func CopyFile(sourceFile string, targetFile string) (err error) {
	targetDir := filepath.Dir(targetFile)
	if err = os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return errW(err, "copy file error",
			reason("make target dir error"),
			kv("targetDir", targetDir),
		)
	}

	targetWriter, err := os.Create(targetFile)
	if err != nil {
		return errW(err, "copy file error",
			reason("create target writer error"),
			kv("targetFile", targetFile),
		)
	}
	defer targetWriter.Close()

	sourceReader, err := os.Open(sourceFile)
	if err != nil {
		return errW(err, "copy file error",
			reason("open source reader error"),
			kv("sourceFile", sourceFile),
		)
	}
	defer sourceReader.Close()

	_, err = io.Copy(targetWriter, sourceReader)
	if err != nil {
		return errW(err, "copy file error",
			reason("io copy error"),
			kv("targetFile", targetFile),
			kv("sourceFile", sourceFile),
		)
	}
	return nil
}

func LinkOrCopyFile(sourceFile string, targetFile string) (err error) {
	err = LinkFile(sourceFile, targetFile)
	if err != nil {
		err = CopyFile(sourceFile, targetFile)
		if err != nil {
			return errW(err, "link or copy file error",
				reason("copy file error"),
				kv("sourceFile", sourceFile),
				kv("targetFile", targetFile),
			)
		}
	}
	return nil
}

func ReadYamlFile(file string, model any) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return errW(err, "read yaml file error",
			reason("read file error"),
			kv("file", file),
		)
	}
	err = yaml.Unmarshal(data, model)
	if err != nil {
		return errW(err, "read yaml file error",
			reason("yaml unmarshal error"),
			kv("file", file),
		)
	}
	return nil
}

func ReadTomlFile(file string, model any) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return errW(err, "read toml file error",
			reason("read file error"),
			kv("file", file),
		)
	}
	err = toml.Unmarshal(data, model)
	if err != nil {
		return errW(err, "read toml file error",
			reason("toml unmarshal error"),
			kv("file", file),
		)
	}
	return nil
}

func ReadJsonFile(file string, model any) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return errW(err, "read json file error",
			reason("read file error"),
			kv("file", file),
		)
	}
	err = json.Unmarshal(data, model)
	if err != nil {
		return errW(err, "read json file error",
			reason("json unmarshal error"),
			kv("file", file),
		)
	}
	return nil
}

func WriteYamlFile(file string, model any) error {
	data, err := yaml.Marshal(model)
	if err != nil {
		return errW(err, "write yaml file error",
			reason("yaml marshal error"),
			kv("file", file),
		)
	}
	return os.WriteFile(file, data, os.ModePerm)
}

func WriteTomlFile(file string, model any) error {
	data, err := toml.Marshal(model)
	if err != nil {
		return errW(err, "write toml file error",
			reason("toml marshal error"),
			kv("file", file),
		)
	}
	return os.WriteFile(file, data, os.ModePerm)
}

func WriteJsonFile(file string, model any, indent bool) error {
	var data []byte
	var err error
	if indent {
		data, err = json.MarshalIndent(model, "", "    ")
	} else {
		data, err = json.Marshal(model)
	}
	if err != nil {
		return errW(err, "write json file error",
			reason("json marshal error"),
			kv("file", file),
		)
	}
	return os.WriteFile(file, data, os.ModePerm)
}

func IsYamlFile(file string) bool {
	return strings.HasSuffix(file, ".yml") || strings.HasSuffix(file, ".yaml")
}

func IsTomlFile(file string) bool {
	return strings.HasSuffix(file, ".toml")
}

func IsJsonFile(file string) bool {
	return strings.HasSuffix(file, ".json")
}

func IsConfigYamlFile(file string) bool {
	return strings.HasSuffix(file, ".dcfg.yml") || strings.HasSuffix(file, ".dcfg.yaml")
}

func IsConfigTomlFile(file string) bool {
	return strings.HasSuffix(file, ".dcfg.toml")
}

func IsConfigJsonFile(file string) bool {
	return strings.HasSuffix(file, ".dcfg.json")
}

func IsTemplateFile(file string) bool {
	return strings.HasSuffix(file, ".dtpl")
}

func IsTemplateLibFile(file string) bool {
	return strings.HasSuffix(file, ".dtpl.lib")
}

func GetFileType(file string, types []FileType) FileType {
	includePlain := false
	for i := 0; i < len(types); i++ {
		switch types[i] {
		case FileTypeConfigYaml:
			if IsConfigYamlFile(file) {
				return FileTypeConfigYaml
			}
		case FileTypeConfigToml:
			if IsConfigTomlFile(file) {
				return FileTypeConfigToml
			}
		case FileTypeConfigJson:
			if IsConfigJsonFile(file) {
				return FileTypeConfigJson
			}
		case FileTypeTemplate:
			if IsTemplateFile(file) {
				return FileTypeTemplate
			}
		case FileTypeTemplateLib:
			if IsTemplateLibFile(file) {
				return FileTypeTemplateLib
			}
		case FileTypeYaml:
			if IsYamlFile(file) {
				return FileTypeYaml
			}
		case FileTypeToml:
			if IsTomlFile(file) {
				return FileTypeToml
			}
		case FileTypeJson:
			if IsJsonFile(file) {
				return FileTypeJson
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

func GetFileNames(globs []string, types []FileType) []string {
	var fileNames []string
	for i := 0; i < len(globs); i++ {
		fileName := globs[i]
		for j := 0; j < len(types); j++ {
			switch types[j] {
			case FileTypeConfigYaml:
				fileNames = append(fileNames, fileName+".dcfg.yml")
				fileNames = append(fileNames, fileName+".dcfg.yaml")
			case FileTypeConfigToml:
				fileNames = append(fileNames, fileName+".dcfg.toml")
			case FileTypeConfigJson:
				fileNames = append(fileNames, fileName+".dcfg.json")
			case FileTypeTemplate:
				fileNames = append(fileNames, fileName+".dtpl")
			case FileTypeTemplateLib:
				fileNames = append(fileNames, fileName+".dtpl.lib")
			case FileTypeYaml:
				fileNames = append(fileNames, fileName+".yml")
				fileNames = append(fileNames, fileName+".yaml")
			case FileTypeToml:
				fileNames = append(fileNames, fileName+".toml")
			case FileTypeJson:
				fileNames = append(fileNames, fileName+".json")
			case FileTypePlain:
				fileNames = append(fileNames, fileName)
			default:
				impossible()
			}
		}
	}
	return fileNames
}

func RemoveFileExt(file string) string {
	ext := filepath.Ext(file)
	if ext == "" {
		return file
	}
	return file[:len(file)-len(ext)]
}

type File struct {
	Path    string
	RelPath string
	Type    FileType
}

func FindFile(dir string, fileNames []string, fileTypes []FileType) *File {
	for i := 0; i < len(fileNames); i++ {
		filePath := filepath.Join(dir, fileNames[i])
		if IsFileExists(filePath) {
			return &File{Path: filePath, RelPath: fileNames[i], Type: GetFileType(filePath, fileTypes)}
		}
	}
	return nil
}

func ScanFiles(dir string, includes []string, excludes []string, types []FileType) (files []*File, err error) {
	var includeFiles map[string]bool
	if len(includes) > 0 {
		includeFiles = map[string]bool{}
		for i := 0; i < len(includes); i++ {
			includeFiles[filepath.Join(dir, includes[i])] = true
		}
	}
	var excludeFiles map[string]bool
	if len(excludes) > 0 {
		excludeFiles = map[string]bool{}
		for i := 0; i < len(excludes); i++ {
			excludeFiles[filepath.Join(dir, excludes[i])] = true
		}
	}
	err = filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return errW(err, "scan files error",
				reason("walk dir error"),
				kv("dir", dir),
				kv("path", path),
			)
		}
		if !entry.IsDir() {
			if includeFiles != nil {
				if _, exist := includeFiles[path]; !exist {
					return nil
				}
			}
			if excludeFiles != nil {
				if _, exist := excludeFiles[path]; exist {
					return nil
				}
			}
			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				return errW(err, "scan files error",
					reason("get rel-path error"),
					kv("dir", dir),
					kv("path", path),
				)
			}
			fileType := GetFileType(relPath, types)
			if fileType != "" {
				files = append(files, &File{Path: path, RelPath: relPath, Type: fileType})
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func ListChildDirs(dir string) (names []string, err error) {
	children, err := os.ReadDir(dir)
	if err != nil {
		return nil, errW(err, "list child dirs error",
			reason("read dir error"),
			kv("dir", dir),
		)
	}
	for i := 0; i < len(children); i++ {
		child := children[i]
		if child.IsDir() {
			names = append(names, child.Name())
		}
	}
	return names, nil
}
