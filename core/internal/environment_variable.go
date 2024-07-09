package internal

import (
	"encoding/json"
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/utils"
	"slices"
	"strings"
)

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
			items = append(items, NewEnvironmentVariableItem(key, value, EnvironmentVariableItemSourceAssign))
			itemKeysDict[key] = true
		}
	}
	for rawKey, value := range system.Variables {
		if key, matched := strings.CutPrefix(rawKey, "DSH_"); matched {
			key = strings.ToLower(key)
			if !itemKeysDict[key] {
				items = append(items, NewEnvironmentVariableItem(key, value, EnvironmentVariableItemSourceSystem))
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
	Source EnvironmentVariableItemSource
	Kind   EnvironmentVariableItemKind
}

type EnvironmentVariableItemSource string

const (
	EnvironmentVariableItemSourceAssign EnvironmentVariableItemSource = "assign"
	EnvironmentVariableItemSourceSystem EnvironmentVariableItemSource = "system"
)

type EnvironmentVariableItemKind string

const (
	EnvironmentVariableItemKindArgumentItem      EnvironmentVariableItemKind = "argument_item"
	EnvironmentVariableItemKindWorkspaceDir      EnvironmentVariableItemKind = "workspace_dir"
	EnvironmentVariableItemKindWorkspaceClean    EnvironmentVariableItemKind = "workspace_clean"
	EnvironmentVariableItemKindWorkspaceProfile  EnvironmentVariableItemKind = "workspace_profile_item"
	EnvironmentVariableItemKindWorkspaceExecutor EnvironmentVariableItemKind = "workspace_executor_item"
	EnvironmentVariableItemKindWorkspaceRegistry EnvironmentVariableItemKind = "workspace_registry_item"
	EnvironmentVariableItemKindWorkspaceRedirect EnvironmentVariableItemKind = "workspace_redirect_item"
	EnvironmentVariableItemKindUnknown           EnvironmentVariableItemKind = "unknown"
)

func NewEnvironmentVariableItem(key, value string, source EnvironmentVariableItemSource) *EnvironmentVariableItem {
	class := EnvironmentVariableItemKindUnknown
	name := key
	if key == "workspace_dir" {
		class = EnvironmentVariableItemKindWorkspaceDir
	} else if key == "workspace_clean" {
		class = EnvironmentVariableItemKindWorkspaceClean
	} else if str, matched := strings.CutPrefix(key, "argument_item_"); matched {
		name = str
		class = EnvironmentVariableItemKindArgumentItem
	} else if str, matched = strings.CutPrefix(key, "workspace_profile_item_"); matched {
		name = str
		class = EnvironmentVariableItemKindWorkspaceProfile
	} else if str, matched = strings.CutPrefix(key, "workspace_executor_item_"); matched {
		name = str
		class = EnvironmentVariableItemKindWorkspaceExecutor
	} else if str, matched = strings.CutPrefix(key, "workspace_registry_item_"); matched {
		name = str
		class = EnvironmentVariableItemKindWorkspaceRegistry
	} else if str, matched = strings.CutPrefix(key, "workspace_redirect_item_"); matched {
		name = str
		class = EnvironmentVariableItemKindWorkspaceRedirect
	}
	return &EnvironmentVariableItem{
		Key:    key,
		Name:   name,
		Value:  value,
		Source: source,
		Kind:   class,
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

// region EnvironmentVariableParsedItemSet

type EnvironmentVariableParsedItemSet[T any] []*EnvironmentVariableParsedItem[T]

func (s EnvironmentVariableParsedItemSet[T]) Sort() EnvironmentVariableParsedItemSet[T] {
	slices.SortStableFunc(s, func(l, r *EnvironmentVariableParsedItem[T]) int {
		return strings.Compare(l.Name, r.Name)
	})
	return s
}

func (s EnvironmentVariableParsedItemSet[T]) GetValues() []T {
	result := make([]T, 0, len(s))
	for i := 0; i < len(s); i++ {
		result = append(result, s[i].Value)
	}
	return result
}

// endregion
