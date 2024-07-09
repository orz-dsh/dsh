package internal

import (
	"encoding/json"
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/utils"
	"slices"
	"strings"
)

// region base

type EnvironmentVariableSource string

const (
	EnvironmentVariableSourceAssign EnvironmentVariableSource = "assign"
	EnvironmentVariableSourceSystem EnvironmentVariableSource = "system"
)

type EnvironmentVariableKind string

const (
	EnvironmentVariableKindArgumentItem      EnvironmentVariableKind = "argument_item"
	EnvironmentVariableKindWorkspaceDir      EnvironmentVariableKind = "workspace_dir"
	EnvironmentVariableKindWorkspaceClean    EnvironmentVariableKind = "workspace_clean"
	EnvironmentVariableKindWorkspaceProfile  EnvironmentVariableKind = "workspace_profile_item"
	EnvironmentVariableKindWorkspaceExecutor EnvironmentVariableKind = "workspace_executor_item"
	EnvironmentVariableKindWorkspaceRegistry EnvironmentVariableKind = "workspace_registry_item"
	EnvironmentVariableKindWorkspaceRedirect EnvironmentVariableKind = "workspace_redirect_item"
	EnvironmentVariableKindUnknown           EnvironmentVariableKind = "unknown"
)

// endregion

// region EnvironmentVariable

type EnvironmentVariable struct {
	Items []*EnvironmentVariableItem
}

func NewEnvironmentVariable(system *System, assigns map[string]string) *EnvironmentVariable {
	var items []*EnvironmentVariableItem
	itemKeysDict := map[string]bool{}
	for rawKey, value := range assigns {
		key := strings.ToLower(rawKey)
		if !itemKeysDict[key] {
			items = append(items, NewEnvironmentVariableItem(key, value, EnvironmentVariableSourceAssign))
			itemKeysDict[key] = true
		}
	}
	for rawKey, value := range system.Variables {
		if key, matched := strings.CutPrefix(rawKey, "DSH_"); matched {
			key = strings.ToLower(key)
			if !itemKeysDict[key] {
				items = append(items, NewEnvironmentVariableItem(key, value, EnvironmentVariableSourceSystem))
				itemKeysDict[key] = true
			}
		}
	}
	return &EnvironmentVariable{
		Items: items,
	}
}

func (v *EnvironmentVariable) Inspect() *EnvironmentVariableInspection {
	items := make([]*EnvironmentVariableItemInspection, 0, len(v.Items))
	for i := 0; i < len(v.Items); i++ {
		items = append(items, v.Items[i].Inspect())
	}
	return NewEnvironmentVariableInspection(items)
}

// endregion

// region EnvironmentVariableItem

type EnvironmentVariableItem struct {
	Key    string
	Name   string
	Value  string
	Source EnvironmentVariableSource
	Kind   EnvironmentVariableKind
}

func NewEnvironmentVariableItem(key, value string, source EnvironmentVariableSource) *EnvironmentVariableItem {
	name := key
	kind := EnvironmentVariableKindUnknown
	if key == "workspace_dir" {
		kind = EnvironmentVariableKindWorkspaceDir
	} else if key == "workspace_clean" {
		kind = EnvironmentVariableKindWorkspaceClean
	} else if str, matched := strings.CutPrefix(key, "argument_item_"); matched {
		name = str
		kind = EnvironmentVariableKindArgumentItem
	} else if str, matched = strings.CutPrefix(key, "workspace_profile_item_"); matched {
		name = str
		kind = EnvironmentVariableKindWorkspaceProfile
	} else if str, matched = strings.CutPrefix(key, "workspace_executor_item_"); matched {
		name = str
		kind = EnvironmentVariableKindWorkspaceExecutor
	} else if str, matched = strings.CutPrefix(key, "workspace_registry_item_"); matched {
		name = str
		kind = EnvironmentVariableKindWorkspaceRegistry
	} else if str, matched = strings.CutPrefix(key, "workspace_redirect_item_"); matched {
		name = str
		kind = EnvironmentVariableKindWorkspaceRedirect
	}
	return &EnvironmentVariableItem{
		Key:    key,
		Name:   name,
		Value:  value,
		Source: source,
		Kind:   kind,
	}
}

func (i *EnvironmentVariableItem) Inspect() *EnvironmentVariableItemInspection {
	return NewEnvironmentVariableItemInspection(i.Key, i.Name, i.Value, string(i.Source), string(i.Kind))
}

// endregion

// region EnvironmentVariableParsedItem

type EnvironmentVariableParsedItem[T any] struct {
	Name  string
	Value T
}

func NewEnvironmentVariableParsedItem[T any](item *EnvironmentVariableItem, value T) (*EnvironmentVariableParsedItem[T], error) {
	if err := json.Unmarshal([]byte(item.Value), value); err != nil {
		return nil, ErrW(err, "parse environment variable error",
			Reason("unmarshal json error"),
			KV("item", item),
		)
	}
	parsed := &EnvironmentVariableParsedItem[T]{
		Name:  item.Name,
		Value: value,
	}
	return parsed, nil
}

// endregion

// region EnvironmentVariableParsedItemSlice

type EnvironmentVariableParsedItemSlice[T any] []*EnvironmentVariableParsedItem[T]

func (s EnvironmentVariableParsedItemSlice[T]) Sort() EnvironmentVariableParsedItemSlice[T] {
	slices.SortStableFunc(s, func(l, r *EnvironmentVariableParsedItem[T]) int {
		return strings.Compare(l.Name, r.Name)
	})
	return s
}

func (s EnvironmentVariableParsedItemSlice[T]) GetValues() []T {
	result := make([]T, 0, len(s))
	for i := 0; i < len(s); i++ {
		result = append(result, s[i].Value)
	}
	return result
}

// endregion
