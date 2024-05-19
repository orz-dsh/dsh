package dsh_core

import "dsh/dsh_utils"

type manifestMetadata struct {
	manifestPath string
	manifestType manifestMetadataType
}

type manifestMetadataType string

const (
	manifestMetadataTypeYaml manifestMetadataType = "yaml"
	manifestMetadataTypeToml manifestMetadataType = "toml"
	manifestMetadataTypeJson manifestMetadataType = "json"
)

func loadManifest(dir string, fileNames []string, manifestEntity any, required bool) (metadata *manifestMetadata, err error) {
	var findFileNames []string
	for i := 0; i < len(fileNames); i++ {
		fileName := fileNames[i]
		findFileNames = append(findFileNames, fileName+".yml")
		findFileNames = append(findFileNames, fileName+".yaml")
		findFileNames = append(findFileNames, fileName+".toml")
		findFileNames = append(findFileNames, fileName+".json")
	}

	manifestPath, manifestFileType := dsh_utils.FindFile(dir, findFileNames, []dsh_utils.FileType{
		dsh_utils.FileTypeYaml,
		dsh_utils.FileTypeToml,
		dsh_utils.FileTypeJson,
	})
	if manifestPath == "" {
		if required {
			return nil, errN("load manifest error",
				reason("manifest file not found"),
				kv("dir", dir),
				kv("fileNames", findFileNames),
			)
		} else {
			return nil, nil
		}
	}

	var manifestType manifestMetadataType
	switch manifestFileType {
	case dsh_utils.FileTypeYaml:
		manifestType = manifestMetadataTypeYaml
		err = dsh_utils.ReadYamlFile(manifestPath, manifestEntity)
	case dsh_utils.FileTypeToml:
		manifestType = manifestMetadataTypeToml
		err = dsh_utils.ReadTomlFile(manifestPath, manifestEntity)
	case dsh_utils.FileTypeJson:
		manifestType = manifestMetadataTypeJson
		err = dsh_utils.ReadJsonFile(manifestPath, manifestEntity)
	default:
		impossible()
	}
	if err != nil {
		return nil, errW(err, "load manifest error",
			reason("read manifest file error"),
			kv("manifestPath", manifestPath),
			kv("manifestType", manifestType),
		)
	}

	metadata = &manifestMetadata{
		manifestPath: manifestPath,
		manifestType: manifestType,
	}
	return metadata, nil
}
