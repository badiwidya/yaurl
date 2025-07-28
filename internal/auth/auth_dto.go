package auth

import (
	"github.com/badiwidya/yaurl/internal/pkg/types"
)

type RegisterUserRequest struct {
	Name     string `json:"name,omitempty"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r RegisterUserRequest) Validate() error {
	errs := make(types.ValidationErrors)

	if len(r.Name) < 3 {
		errs["name"] = "must be at least 3 characters"
	}

	if len(r.Username) < 3 {
		errs["username"] = "must be at least 3 characters"
	}

	if len(r.Password) < 8 {
		errs["password"] = "must be at least 3 characters"
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (l LoginUserRequest) Validate() error {
	errs := make(types.ValidationErrors)

	if l.Username == "" {
		errs["username"] = "field required"
	}

	if l.Password == "" {
		errs["password"] = "field required"
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
