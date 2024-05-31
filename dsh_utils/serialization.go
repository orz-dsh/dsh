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

var deserializeFindFileTypes = []FileType{
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

func DeserializeByDir(dir string, fileNames []string, manifestEntity any, required bool) (metadata *SerializationMetadata, err error) {
	var findFileNames []string
	for i := 0; i < len(fileNames); i++ {
		fileName := fileNames[i]
		findFileNames = append(findFileNames, fileName+".yml")
		findFileNames = append(findFileNames, fileName+".yaml")
		findFileNames = append(findFileNames, fileName+".toml")
		findFileNames = append(findFileNames, fileName+".json")
	}

	filePath, fileType := FindFile(dir, findFileNames, deserializeFindFileTypes)
	if filePath == "" {
		if required {
			return nil, errN("deserialize error",
				reason("file not found"),
				kv("dir", dir),
				kv("findFileNames", findFileNames),
			)
		} else {
			return nil, nil
		}
	}

	return DeserializeByFile(filePath, GetSerializationFormat(fileType), manifestEntity)
}

func DeserializeByFile(path string, format SerializationFormat, model any) (metadata *SerializationMetadata, err error) {
	if format == "" {
		if !IsFileExists(path) {
			return nil, errN("deserialize error",
				reason("file not found"),
				kv("path", path),
			)
		}
		fileType := GetFileType(path, deserializeFindFileTypes)
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
