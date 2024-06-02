package dsh_core

// region source

type projectSchemaSource struct {
	Dir   string
	Files []string
	Match string
	match *EvalExpr
}

type projectSchemaSourceSet []*projectSchemaSource

func newProjectSchemaSource(dir string, files []string, match string, matchObj *EvalExpr) *projectSchemaSource {
	return &projectSchemaSource{
		Dir:   dir,
		Files: files,
		Match: match,
		match: matchObj,
	}
}

// endregion
