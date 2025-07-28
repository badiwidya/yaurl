package types

import (
	"fmt"
	"strings"
)

type ValidationErrors map[string]string

func (v ValidationErrors) Error() string {
	var errs []string
	for key, err := range v {
		errs = append(errs, fmt.Sprintf("%s: %s", key, err))
	}
	return strings.Join(errs, ", ")
}
