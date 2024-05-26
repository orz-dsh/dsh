package dsh_core

type projectImportEvaluator struct {
	data *appProfileEvalData
}

func newProjectImportEvaluator(data *appProfileEvalData) *projectImportEvaluator {
	return &projectImportEvaluator{
		data: data,
	}
}

//func (e *projectImportEvaluator) evalRegistry(resource string, path string, ref string) (string, error) {
//	return dsh_utils.EvalStringTemplate(resource, e.data.mergeMap(
//		map[string]any{
//			"path": path,
//			"ref":  ref,
//		},
//	), nil)
//}
//
//func (e *projectImportEvaluator) evalRedirect(resource string, path string, original *projectImport) (string, error) {
//	originalData := make(map[string]any)
//	if original.Local != nil {
//		originalData["mode"] = "local"
//		originalData["dir"] = original.Local.RawDir
//	} else if original.Git != nil {
//		originalData["mode"] = "git"
//		originalData["url"] = original.Git.RawUrl
//		originalData["ref"] = original.Git.RawRef
//	} else {
//		impossible()
//	}
//	return dsh_utils.EvalStringTemplate(resource, e.data.mergeMap(
//		map[string]any{
//			"path":     path,
//			"original": originalData,
//		},
//	), nil)
//}
