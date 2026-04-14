package helpers

import (
	"strconv"

	"github.com/largeoliu/redmine-cli/internal/errors"
)

// ParseID parses and validates a resource ID from a string argument.
func ParseID(arg string, resourceName string) (int, error) {
	id, err := strconv.Atoi(arg)
	if err != nil {
		return 0, errors.NewValidation("invalid " + resourceName + " ID")
	}
	return id, nil
}
