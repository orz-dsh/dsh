package dsh_core

import (
	"github.com/expr-lang/expr/vm"
)

// region project option

type projectOptionDefinition struct {
	Name  string
	Value string
	Match string
	match *vm.Program
}

type projectOptionDefinitions []*projectOptionDefinition

func newProjectOptionDefinition(name string, value string, match string, matchExpr *vm.Program) *projectOptionDefinition {
	return &projectOptionDefinition{
		Name:  name,
		Value: value,
		Match: match,
		match: matchExpr,
	}
}

func (ds projectOptionDefinitions) getItems(evaluator *Evaluator) (map[string]string, error) {
	items := map[string]string{}
	for i := 0; i < len(ds); i++ {
		definition := ds[i]
		if _, exist := items[definition.Name]; exist {
			// The priority of the previous one is higher than that of the later one.
			continue
		}
		matched, err := evaluator.EvalBoolExpr(definition.match)
		if err != nil {
			return nil, err
		}
		if matched {
			items[definition.Name] = definition.Value
		}
	}
	return items, nil
}

// endregion

// region project source

type projectSourceDefinition struct {
	Dir   string
	Files []string
	Match string
	match *vm.Program
}

func newProjectSourceDefinition(dir string, files []string, match string, matchExpr *vm.Program) *projectSourceDefinition {
	return &projectSourceDefinition{
		Dir:   dir,
		Files: files,
		Match: match,
		match: matchExpr,
	}
}

// endregion

// region project import

type projectImportDefinition struct {
	Link  *ProjectLink
	Match string
	match *vm.Program
}

func newProjectImportDefinition(link *ProjectLink, match string, matchExpr *vm.Program) *projectImportDefinition {
	return &projectImportDefinition{
		Link:  link,
		Match: match,
		match: matchExpr,
	}
}

// endregion
