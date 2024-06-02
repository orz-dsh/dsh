package dsh_core

// region import

type projectSchemaImport struct {
	Link  string
	Match string
	link  *projectLink
	match *EvalExpr
}

type projectSchemaImportSet []*projectSchemaImport

func newProjectSchemaImport(link string, match string, linkObj *projectLink, matchObj *EvalExpr) *projectSchemaImport {
	return &projectSchemaImport{
		Link:  link,
		Match: match,
		link:  linkObj,
		match: matchObj,
	}
}

// endregion
