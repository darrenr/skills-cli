package skill

import (
	"fmt"
	"regexp"
	"strings"
)

// nameRe validates name character constraints: lowercase alnum and hyphens,
// no leading or trailing hyphen. Length is checked separately.
var nameRe = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// ValidationError holds one or more constraint violations found during validation.
type ValidationError struct {
	Errs []string
}

func (e *ValidationError) Error() string {
	return strings.Join(e.Errs, "; ")
}

// Validate checks s against the Agent Skills specification constraints.
// All violations are collected and returned as a single *ValidationError,
// or nil if the skill is valid.
func Validate(s *Skill) error {
	var errs []string

	// name: required, 1-64 chars, lowercase alnum + hyphens, no leading/trailing hyphen.
	switch {
	case s.Name == "":
		errs = append(errs, "name is required")
	case len(s.Name) > 64:
		errs = append(errs, fmt.Sprintf("name must be 64 characters or fewer (got %d)", len(s.Name)))
	case !nameRe.MatchString(s.Name):
		errs = append(errs, "name must contain only lowercase letters, digits, and hyphens, "+
			"and must not start or end with a hyphen")
	}

	// description: required, max 1024 chars.
	switch {
	case s.Description == "":
		errs = append(errs, "description is required")
	case len(s.Description) > 1024:
		errs = append(errs, fmt.Sprintf("description must be 1024 characters or fewer (got %d)", len(s.Description)))
	}

	// compatibility: optional, max 500 chars.
	if len(s.Compatibility) > 500 {
		errs = append(errs, fmt.Sprintf("compatibility must be 500 characters or fewer (got %d)", len(s.Compatibility)))
	}

	if len(errs) > 0 {
		return &ValidationError{Errs: errs}
	}
	return nil
}

// ValidateNameMatchesDir returns an error if the skill's name field does not
// match the directory name it was loaded from. Per spec, these must be equal.
func ValidateNameMatchesDir(s *Skill, dirName string) error {
	if s.Name != dirName {
		return fmt.Errorf("skill name %q does not match directory name %q", s.Name, dirName)
	}
	return nil
}
