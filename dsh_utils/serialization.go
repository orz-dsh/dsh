package dsh_utils

type SerializationMetadata struct {
	Path   string
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
	case FileTypeYaml:
		return SerializationFormatYaml
	case FileTypeToml:
		return SerializationFormatToml
	case FileTypeJson:
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

func DeserializeFromFile(path string, format SerializationFormat, model any) (metadata *SerializationMetadata, err error) {
	if format == "" {
		if !IsFileExists(path) {
			return nil, errN("deserialize error",
				reason("file not found"),
				kv("path", path),
			)
		}
		fileType := GetFileType(path, serializationSupportedFileTypes)
		if fileType == "" {
			return nil, errN("deserialize error",
				reason("file type not supported"),
				kv("path", path),
			)
		}
		format = GetSerializationFormat(fileType)
	}

	switch format {
	case SerializationFormatYaml:
		err = ReadYamlFile(path, model)
	case SerializationFormatToml:
		err = ReadTomlFile(path, model)
	case SerializationFormatJson:
		err = ReadJsonFile(path, model)
	default:
		impossible()
	}

	if err != nil {
		return nil, errW(err, "deserialize error",
			reason("read file error"),
			kv("path", path),
			kv("format", format),
		)
	}

	metadata = &SerializationMetadata{
		Path:   path,
		Format: format,
	}
	return metadata, nil
}
