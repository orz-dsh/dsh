package dsh_utils

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
		impossible()
	}
	return ""
}

func DeserializeFromDir(dir string, globs []string, model any, required bool) (metadata *SerializationMetadata, err error) {
	names := GetFileNames(globs, serializationSupportedFileTypes)
	file := FindFile(dir, names, serializationSupportedFileTypes)
	if file == nil {
		if required {
			return nil, errN("deserialize error",
				reason("file not found"),
				kv("dir", dir),
				kv("names", names),
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
			return nil, errN("deserialize error",
				reason("file not found"),
				kv("file", file),
			)
		}
		fileType := GetFileType(file, serializationSupportedFileTypes)
		if fileType == "" {
			return nil, errN("deserialize error",
				reason("file type not supported"),
				kv("file", file),
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
		impossible()
	}

	if err != nil {
		return nil, errW(err, "deserialize error",
			reason("read file error"),
			kv("file", file),
			kv("format", format),
		)
	}

	metadata = &SerializationMetadata{
		File:   file,
		Format: format,
	}
	return metadata, nil
}
