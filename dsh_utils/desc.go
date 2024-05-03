package dsh_utils

import (
	"fmt"
	"strings"
)

type Desc struct {
	Title string
	Body  DescBody
}

type DescBody []string

type DescKeyValue struct {
	Key   string
	Value any
}

type DescList []Desc

func NewDesc(title string, kvs []DescKeyValue) Desc {
	return Desc{
		Title: title,
		Body:  NewDescBody(kvs),
	}
}

func (d Desc) ToString(titleIdent string, bodyIdent string) string {
	if len(d.Body) == 0 {
		return fmt.Sprintf("%s%s\n", titleIdent, d.Title)
	}
	return fmt.Sprintf("%s%s\n%s", titleIdent, d.Title, d.Body.ToString(bodyIdent))
}

func (d Desc) String() string {
	return d.ToString("", "\t")
}

func NewDescBody(kvs []DescKeyValue) DescBody {
	var body []string
	for i := 0; i < len(kvs); i++ {
		body = append(body, kvs[i].String())
	}
	return body
}

func (body DescBody) ToString(ident string) string {
	var builder strings.Builder
	for i := 0; i < len(body); i++ {
		builder.WriteString(ident + body[i] + "\n")
	}
	return builder.String()
}

func (body DescBody) String() string {
	return body.ToString("")
}

func NewDescKeyValue(key string, value any) DescKeyValue {
	return DescKeyValue{
		Key:   key,
		Value: value,
	}
}

func (kv DescKeyValue) String() string {
	k := kv.Key
	v := kv.Value
	vStr := fmt.Sprintf("%+v", v)
	vStr = strings.ReplaceAll(vStr, "\n", "\\n")
	vStr = strings.ReplaceAll(vStr, "\r", "\\r")
	return k + " = `" + vStr + "`"
}

func (list DescList) ToString(titleIdent string, bodyIdent string) string {
	var builder strings.Builder
	for i := 0; i < len(list); i++ {
		builder.WriteString(list[i].ToString(titleIdent, bodyIdent))
	}
	return builder.String()
}

func (list DescList) String() string {
	return list.ToString("", "\t")
}
