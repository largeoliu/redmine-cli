// Package helpers provides shared utilities for resource command implementations.
package helpers

import (
	"fmt"
)

// ConfirmDelete prompts the user to confirm a delete operation.
// Returns true if the user confirms or if yes is true (skip prompt).
func ConfirmDelete(resourceName string, id int, yes bool) bool {
	if yes {
		return true
	}
	fmt.Printf("Delete %s #%d? [y/N]: ", resourceName, id)
	var confirm string
	if _, err := fmt.Scanln(&confirm); err != nil {
		confirm = "n"
	}
	if confirm != "y" && confirm != "Y" {
		fmt.Println("Canceled")
		return false
	}
	return true
}
