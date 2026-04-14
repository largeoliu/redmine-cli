package helpers

import (
	"fmt"
)

// DryRunCreate prints a dry-run message for a create operation and returns true
// (meaning the operation was a dry-run and should be skipped).
func DryRunCreate(resourceName string, req any) bool {
	fmt.Printf("[dry-run] Would create %s: %+v\n", resourceName, req)
	return true
}

// DryRunUpdate prints a dry-run message for an update operation and returns true.
func DryRunUpdate(resourceName string, id int, req any) bool {
	fmt.Printf("[dry-run] Would update %s #%d: %+v\n", resourceName, id, req)
	return true
}

// DryRunDelete prints a dry-run message for a delete operation and returns true.
func DryRunDelete(resourceName string, id int) bool {
	fmt.Printf("[dry-run] Would delete %s #%d\n", resourceName, id)
	return true
}
