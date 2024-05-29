package dsh_core

import "github.com/expr-lang/expr/vm"

// region option specify

type projectOptionSpecifyEntity struct {
	Name  string
	Value string
	Match string
	match *vm.Program
}

type projectOptionSpecifyEntitySet []*projectOptionSpecifyEntity

func newProjectOptionSpecifyEntity(name string, value string, match string, matchObj *vm.Program) *projectOptionSpecifyEntity {
	return &projectOptionSpecifyEntity{
		Name:  name,
		Value: value,
		Match: match,
		match: matchObj,
	}
}

func (s projectOptionSpecifyEntitySet) getItems(evaluator *Evaluator) (map[string]string, error) {
	items := map[string]string{}
	for i := 0; i < len(s); i++ {
		entity := s[i]
		if _, exist := items[entity.Name]; exist {
			continue
		}
		matched, err := evaluator.EvalBoolExpr(entity.match)
		if err != nil {
			return nil, errW(err, "get project option specify items error",
				reason("eval expr error"),
				kv("entity", entity),
			)
		}
		if matched {
			items[entity.Name] = entity.Value
		}
	}
	return items, nil
}

// endregion

// region source

type projectSourceEntity struct {
	Dir   string
	Files []string
	Match string
	match *vm.Program
}

type projectSourceEntitySet []*projectSourceEntity

func newProjectSourceEntity(dir string, files []string, match string, matchObj *vm.Program) *projectSourceEntity {
	return &projectSourceEntity{
		Dir:   dir,
		Files: files,
		Match: match,
		match: matchObj,
	}
}

// endregion

// region import

type projectImportEntity struct {
	Link  string
	Match string
	link  *ProjectLink
	match *vm.Program
}

type projectImportEntitySet []*projectImportEntity

func newProjectImportEntity(link string, match string, linkObj *ProjectLink, matchObj *vm.Program) *projectImportEntity {
	return &projectImportEntity{
		Link:  link,
		Match: match,
		link:  linkObj,
		match: matchObj,
	}
}

// endregion
