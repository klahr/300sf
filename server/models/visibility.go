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
