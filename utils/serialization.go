package utils

import (
	"encoding/json"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
	"os"
)

type SerializationMetadata struct {
	File   string
	Format SerializationFormat
}

type SerializationFormat string

const (
	SerializationFormatYaml SerializationFormat = "yaml"
	SerializationFormatToml SerializationFormat = "toml"
	SerializationFormatJson SerializationFormat = "json"
)

var serializationSupportedFileTypes = []FileType{
	FileTypeYaml,
	FileTypeToml,
	FileTypeJson,
}

func GetSerializationFormat(fileType FileType) SerializationFormat {
	switch fileType {
	case FileTypeConfigYaml, FileTypeYaml:
		return SerializationFormatYaml
	case FileTypeConfigToml, FileTypeToml:
		return SerializationFormatToml
	case FileTypeConfigJson, FileTypeJson:
		return SerializationFormatJson
	default:
		Impossible()
	}
	return ""
}

func DeserializeDir(dir string, globs []string, model any, required bool) (metadata *SerializationMetadata, err error) {
	names := GetFileNames(globs, serializationSupportedFileTypes)
	file := FindFile(dir, names, serializationSupportedFileTypes)
	if file == nil {
		if required {
			return nil, ErrN("deserialize error",
				Reason("file not found"),
				KV("dir", dir),
				KV("names", names),
			)
		} else {
			return nil, nil
		}
	}

	return DeserializeFile(file.Path, GetSerializationFormat(file.Type), model)
}

func DeserializeFile(file string, format SerializationFormat, model any) (metadata *SerializationMetadata, err error) {
	if format == "" {
		if !IsFileExists(file) {
			return nil, ErrN("deserialize error",
				Reason("file not found"),
				KV("file", file),
			)
		}
		fileType := GetFileType(file, serializationSupportedFileTypes)
		if fileType == "" {
			return nil, ErrN("deserialize error",
				Reason("file type not supported"),
				KV("file", file),
			)
		}
		format = GetSerializationFormat(fileType)
	}

	switch format {
	case SerializationFormatYaml:
		err = ReadYamlFile(file, model)
	case SerializationFormatToml:
		err = ReadTomlFile(file, model)
	case SerializationFormatJson:
		err = ReadJsonFile(file, model)
	default:
		Impossible()
	}

	if err != nil {
		return nil, ErrW(err, "deserialize error",
			Reason("read file error"),
			KV("file", file),
			KV("format", format),
		)
	}

	metadata = &SerializationMetadata{
		File:   file,
		Format: format,
	}
	return metadata, nil
}

// region Serializer

type Serializer interface {
	GetFormat() SerializationFormat

	GetFileExt() string

	SerializeFile(file string, model any) error
}

// endregion

// region YamlSerializer

type YamlSerializer struct {
	Indent int
}

var YamlSerializerDefault = NewYamlSerializer(2)

func NewYamlSerializer(indent int) *YamlSerializer {
	return &YamlSerializer{
		Indent: indent,
	}
}

func (s *YamlSerializer) SetIndent(indent int) *YamlSerializer {
	return NewYamlSerializer(indent)
}

func (s *YamlSerializer) GetFormat() SerializationFormat {
	return SerializationFormatYaml
}

func (s *YamlSerializer) GetFileExt() string {
	return ".yml"
}

func (s *YamlSerializer) SerializeFile(file string, model any) error {
	writer, err := os.Create(file)
	if err != nil {
		return ErrW(err, "serialize error",
			Reason("create writer error"),
			KV("file", file),
		)
	}
	defer writer.Close()

	encoder := yaml.NewEncoder(writer)
	encoder.SetIndent(s.Indent)
	defer encoder.Close()

	if err = encoder.Encode(model); err != nil {
		return ErrW(err, "serialize error",
			Reason("encode yaml error"),
			KV("file", file),
		)
	}
	return nil
}

// endregion

// region TomlSerializer

type TomlSerializer struct {
	TablesInline    bool
	IndentTables    bool
	ArraysMultiline bool
	IndentSymbol    string
}

var TomlSerializerDefault = NewTomlSerializer(false, true, true, "  ")

func NewTomlSerializer(tableInline, indentTables, arraysMultiline bool, indentSymbol string) *TomlSerializer {
	return &TomlSerializer{
		TablesInline:    tableInline,
		IndentTables:    indentTables,
		ArraysMultiline: arraysMultiline,
		IndentSymbol:    indentSymbol,
	}
}

func (s *TomlSerializer) SetTablesInline(inline bool) *TomlSerializer {
	return NewTomlSerializer(inline, s.IndentTables, s.ArraysMultiline, s.IndentSymbol)
}

func (s *TomlSerializer) SetIndentTables(indent bool) *TomlSerializer {
	return NewTomlSerializer(s.TablesInline, indent, s.ArraysMultiline, s.IndentSymbol)
}

func (s *TomlSerializer) SetArraysMultiline(multiline bool) *TomlSerializer {
	return NewTomlSerializer(s.TablesInline, s.IndentTables, multiline, s.IndentSymbol)
}

func (s *TomlSerializer) SetIndentSymbol(symbol string) *TomlSerializer {
	return NewTomlSerializer(s.TablesInline, s.IndentTables, s.ArraysMultiline, symbol)
}

func (s *TomlSerializer) GetFormat() SerializationFormat {
	return SerializationFormatToml
}

func (s *TomlSerializer) GetFileExt() string {
	return ".toml"
}

func (s *TomlSerializer) SerializeFile(file string, model any) error {
	writer, err := os.Create(file)
	if err != nil {
		return ErrW(err, "serialize error",
			Reason("create writer error"),
			KV("file", file),
		)
	}
	defer writer.Close()

	encoder := toml.NewEncoder(writer)
	encoder.SetTablesInline(s.TablesInline)
	encoder.SetIndentTables(s.IndentTables)
	encoder.SetArraysMultiline(s.ArraysMultiline)
	encoder.SetIndentSymbol(s.IndentSymbol)
	if err = encoder.Encode(model); err != nil {
		return ErrW(err, "serialize error",
			Reason("encode toml error"),
			KV("file", file),
		)
	}
	return nil
}

// endregion

// region JsonSerializer

type JsonSerializer struct {
	PrefixSymbol string
	IndentSymbol string
}

var JsonSerializerDefault = NewJsonSerializer("", "  ")

func NewJsonSerializer(prefixSymbol, indentSymbol string) *JsonSerializer {
	return &JsonSerializer{
		PrefixSymbol: prefixSymbol,
		IndentSymbol: indentSymbol,
	}
}

func (s *JsonSerializer) SetPrefixSymbol(symbol string) *JsonSerializer {
	return NewJsonSerializer(symbol, s.IndentSymbol)
}

func (s *JsonSerializer) SetIndentSymbol(symbol string) *JsonSerializer {
	return NewJsonSerializer(s.PrefixSymbol, symbol)
}

func (s *JsonSerializer) GetFormat() SerializationFormat {
	return SerializationFormatJson
}

func (s *JsonSerializer) GetFileExt() string {
	return ".json"
}

func (s *JsonSerializer) SerializeFile(file string, model any) error {
	var data []byte
	var err error
	if s.PrefixSymbol != "" || s.IndentSymbol != "" {
		data, err = json.MarshalIndent(model, s.PrefixSymbol, s.IndentSymbol)
	} else {
		data, err = json.Marshal(model)
	}
	if err != nil {
		return ErrW(err, "serialize error",
			Reason("marshal json error"),
			KV("file", file),
		)
	}

	if err = os.WriteFile(file, data, os.ModePerm); err != nil {
		return ErrW(err, "serialize error",
			Reason("write file error"),
			KV("file", file),
		)
	}
	return nil
}

// endregion
