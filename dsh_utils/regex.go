package dsh_utils

import (
	"fmt"
	"regexp"
)

func RegexMatch(regex *regexp.Regexp, str string) (matched bool, values map[string]string) {
	match := regex.FindStringSubmatch(str)
	if len(match) == 0 {
		return false, nil
	}
	groups := regex.SubexpNames()
	values = map[string]string{}
	for i := 1; i < len(match); i++ {
		group := groups[i]
		if group != "" {
			values[groups[i]] = match[i]
		}
		values[fmt.Sprintf("$%d", i)] = match[i]
	}
	return true, values
}
