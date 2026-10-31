// SPDX-License-Identifier: AGPL-3.0-or-later
package models

const (
	VisibilityPrivate = 0
	VisibilityHidden  = 1
	VisibilityPublic  = 2
)

func ValidVisibility(v int64) bool {
	return v >= VisibilityPrivate && v <= VisibilityPublic
}

// ListVisibility is how a checklist is stored with visibility v. Nothing lists
// public checklists, so public is the same as anyone with the link and is
// stored as that.
func ListVisibility(v int64) int64 {
	if v == VisibilityPublic {
		return VisibilityHidden
	}
	return v
}
