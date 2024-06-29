package utils

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

func DeserializeFromDir(dir string, globs []string, model any, required bool) (metadata *SerializationMetadata, err error) {
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

	return DeserializeFromFile(file.Path, GetSerializationFormat(file.Type), model)
}

func DeserializeFromFile(file string, format SerializationFormat, model any) (metadata *SerializationMetadata, err error) {
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
