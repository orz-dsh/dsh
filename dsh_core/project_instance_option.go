package dsh_core

import (
	"dsh/dsh_utils"
	"github.com/expr-lang/expr/vm"
	"slices"
)

type projectInstanceOption struct {
	context   *projectContext
	manifest  *projectManifest
	values    map[string]string
	items     map[string]any
	initiated bool
}

func newProjectInstanceOption(context *projectContext, manifest *projectManifest, values map[string]string) (*projectInstanceOption, error) {
	if values == nil {
		values = make(map[string]string)
	}
	option := &projectInstanceOption{
		context:  context,
		manifest: manifest,
		values:   values,
		items:    make(map[string]any),
	}
	err := option.init()
	if err != nil {
		return nil, err
	}
	return option, nil
}

func (option *projectInstanceOption) init() (err error) {
	if option.initiated {
		return nil
	}
	manifest := option.manifest
	for i := 0; i < len(manifest.Option.Items); i++ {
		if err = option.addItem(manifest.Option.Items[i]); err != nil {
			return err
		}
	}
	err = option.verify()
	if err != nil {
		return err
	}
	option.initiated = true
	return nil
}

func (option *projectInstanceOption) addItem(item *projectManifestOptionItem) error {
	manifest := option.manifest
	if _, exist := option.items[item.Name]; exist {
		return dsh_utils.NewError("duplicate option", map[string]any{
			"projectName": manifest.Name,
			"projectPath": manifest.projectPath,
			"optionName":  item.Name,
		})
	}
	var value any = nil
	originalValue := ""
	if v, exist := option.values[item.Name]; exist {
		originalValue = v
		if len(item.Choices) > 0 && !slices.Contains(item.Choices, originalValue) {
			return dsh_utils.NewError("option value invalid", map[string]any{
				"projectName":   manifest.Name,
				"projectPath":   manifest.projectPath,
				"optionName":    item.Name,
				"optionValue":   originalValue,
				"optionChoices": item.Choices,
			})
		}
		switch item.Type {
		case projectManifestOptionItemTypeString:
			value = v
		case projectManifestOptionItemTypeBool:
			value = v == "true"
		case projectManifestOptionItemTypeInteger:
			integer, err := dsh_utils.ParseInteger(v)
			if err != nil {
				return dsh_utils.WrapError(err, "option integer value invalid", map[string]any{
					"projectName": manifest.Name,
					"projectPath": manifest.projectPath,
					"optionName":  item.Name,
					"optionValue": v,
				})
			}
			value = integer
		case projectManifestOptionItemTypeDecimal:
			decimal, err := dsh_utils.ParseDecimal(v)
			if err != nil {
				return dsh_utils.WrapError(err, "option decimal value invalid", map[string]any{
					"projectName": manifest.Name,
					"projectPath": manifest.projectPath,
					"optionName":  item.Name,
					"optionValue": v,
				})
			}
			value = decimal
		}
	} else if linkValue, exist, err := option.context.getOptionLinkValue(manifest.Name, item.Name); exist {
		if linkValue != nil {
			originalValue = *linkValue
			if len(item.Choices) > 0 && !slices.Contains(item.Choices, originalValue) {
				return dsh_utils.NewError("option value invalid", map[string]any{
					"projectName":   manifest.Name,
					"projectPath":   manifest.projectPath,
					"optionName":    item.Name,
					"optionValue":   originalValue,
					"optionChoices": item.Choices,
				})
			}
			switch item.Type {
			case projectManifestOptionItemTypeString:
				value = originalValue
			case projectManifestOptionItemTypeBool:
				value = originalValue == "true"
			case projectManifestOptionItemTypeInteger:
				integer, err := dsh_utils.ParseInteger(originalValue)
				if err != nil {
					return dsh_utils.WrapError(err, "option integer value invalid", map[string]any{
						"projectName": manifest.Name,
						"projectPath": manifest.projectPath,
						"optionName":  item.Name,
						"optionValue": originalValue,
					})
				}
				value = integer
			case projectManifestOptionItemTypeDecimal:
				decimal, err := dsh_utils.ParseDecimal(originalValue)
				if err != nil {
					return dsh_utils.WrapError(err, "option decimal value invalid", map[string]any{
						"projectName": manifest.Name,
						"projectPath": manifest.projectPath,
						"optionName":  item.Name,
						"optionValue": originalValue,
					})
				}
				value = decimal
			}
		}
	} else if err != nil {
		return err
	} else if item.defaultValue != nil {
		value = item.defaultValue
		originalValue = *item.Default
	}
	if value == nil && !item.Optional {
		return dsh_utils.NewError("option required", map[string]any{
			"projectName": manifest.Name,
			"projectPath": manifest.projectPath,
			"optionName":  item.Name,
		})
	}
	if value != nil {
		if err := option.context.addOptionValue(manifest.Name, item.Name, originalValue); err != nil {
			return err
		}
		option.items[item.Name] = value
	}
	for i := 0; i < len(item.Links); i++ {
		link := item.Links[i]
		if err := option.context.addOptionLink(link.Project, link.Option, manifest.Name, item.Name, link.mapping); err != nil {
			return err
		}
	}
	return nil
}

func (option *projectInstanceOption) verify() error {
	manifest := option.manifest
	verifies := manifest.Option.verifies
	for i := 0; i < len(verifies); i++ {
		result, err := dsh_utils.EvalExprReturnBool(verifies[i], option.items)
		if err != nil {
			return err
		}
		if !result {
			return dsh_utils.NewError("option verify failed", map[string]any{
				"projectName": manifest.Name,
				"projectPath": manifest.projectPath,
				"verify":      verifies[i].Source().Content(),
			})
		}
	}
	return nil
}

func (option *projectInstanceOption) match(matchExpr *vm.Program) (bool, error) {
	return dsh_utils.EvalExprReturnBool(matchExpr, option.items)
}
