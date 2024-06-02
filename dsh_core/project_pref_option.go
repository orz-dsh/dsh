package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
)

// region option

type projectPrefOption struct {
	Items    []*projectPrefOptionItem
	Verifies []string
}

type projectPrefOptionItem struct {
	Name     string
	Type     projectOptionValueType
	Choices  []string
	Default  *string
	Optional bool
	Assigns  []*projectPrefOptionItemAssign
}

type projectPrefOptionItemAssign struct {
	Project string
	Option  string
	Mapping string
}

func (o *projectPrefOption) init(manifest *projectManifest) (projectSchemaOptionSet, projectSchemaOptionVerifySet, error) {
	declares := projectSchemaOptionSet{}
	optionNamesDict := map[string]bool{}
	assignTargetsDict := map[string]bool{}
	for i := 0; i < len(o.Items); i++ {
		if declareEntity, err := o.Items[i].init(manifest, optionNamesDict, assignTargetsDict, i); err != nil {
			return nil, nil, err
		} else {
			declares = append(declares, declareEntity)
		}
	}

	verifies := projectSchemaOptionVerifySet{}
	for i := 0; i < len(o.Verifies); i++ {
		expr := o.Verifies[i]
		if expr == "" {
			return nil, nil, errN("project manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("option.verifies[%d]", i)),
			)
		}
		exprObj, err := dsh_utils.CompileExpr(expr)
		if err != nil {
			return nil, nil, errW(err, "project manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("option.verifies[%d]", i)),
				kv("value", expr),
			)
		}
		verifies = append(verifies, newProjectSchemaOptionVerify(expr, exprObj))
	}

	return declares, verifies, nil
}

func (i *projectPrefOptionItem) init(manifest *projectManifest, itemNamesDict, assignTargetsDict map[string]bool, itemIndex int) (entity *projectSchemaOption, err error) {
	if i.Name == "" {
		return nil, errN("project manifest invalid",
			reason("name empty"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("option.items[%d].name", itemIndex)),
		)
	}
	if checked := projectOptionNameCheckRegex.MatchString(i.Name); !checked {
		return nil, errN("project manifest invalid",
			reason("value invalid"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("option.items[%d].name", itemIndex)),
			kv("value", i.Name),
		)
	}
	if _, exist := itemNamesDict[i.Name]; exist {
		return nil, errN("project manifest invalid",
			reason("name duplicated"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("option.items[%d].name", itemIndex)),
			kv("value", i.Name),
		)
	}
	valueType := i.Type
	if valueType == "" {
		valueType = projectOptionValueTypeString
	}
	switch valueType {
	case projectOptionValueTypeString:
	case projectOptionValueTypeBool:
	case projectOptionValueTypeInteger:
	case projectOptionValueTypeDecimal:
	default:
		return nil, errN("project manifest invalid",
			reason("value invalid"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("option.items[%d].type", itemIndex)),
			kv("value", i.Type),
		)
	}
	entity = newProjectSchemaOption(i.Name, valueType, i.Choices, i.Optional)
	if err = entity.setDefaultValue(i.Default); err != nil {
		return nil, errW(err, "project manifest invalid",
			reason("value invalid"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("option.items[%d].default", itemIndex)),
			kv("value", *i.Default),
		)
	}

	for assignIndex := 0; assignIndex < len(i.Assigns); assignIndex++ {
		assign := i.Assigns[assignIndex]
		if assignEntity, err := assign.init(manifest, assignTargetsDict, itemIndex, assignIndex); err != nil {
			return nil, err
		} else {
			entity.addAssign(assignEntity)
		}
	}

	itemNamesDict[i.Name] = true
	return entity, nil
}

func (a *projectPrefOptionItemAssign) init(manifest *projectManifest, targetsDict map[string]bool, itemIndex int, assignIndex int) (entity *projectSchemaOptionAssign, err error) {
	if a.Project == "" {
		return nil, errN("project manifest invalid",
			reason("value empty"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("option.items[%d].assigns[%d].project", itemIndex, assignIndex)),
		)
	}
	if a.Project == manifest.Name {
		return nil, errN("project manifest invalid",
			reason("can not assign to self project option"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("option.items[%d].assigns[%d].project", itemIndex, assignIndex)),
		)
	}
	if a.Option == "" {
		return nil, errN("project manifest invalid",
			reason("value empty"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("option.items[%d].assigns[%d].option", itemIndex, assignIndex)),
		)
	}
	assignTarget := a.Project + "." + a.Option
	if _, exists := targetsDict[assignTarget]; exists {
		return nil, errN("project manifest invalid",
			reason("option assign target duplicated"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("option.items[%d].assigns[%d]", itemIndex, assignIndex)),
			kv("target", assignTarget),
		)
	}
	var mappingObj *EvalExpr
	if a.Mapping != "" {
		mappingObj, err = dsh_utils.CompileExpr(a.Mapping)
		if err != nil {
			return nil, errW(err, "project manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("option.items[%d].assigns[%d].mapping", itemIndex, assignIndex)),
				kv("value", a.Mapping),
			)
		}
	}

	targetsDict[assignTarget] = true
	return newProjectSchemaOptionAssign(a.Project, a.Option, a.Mapping, mappingObj), nil
}

// endregion
