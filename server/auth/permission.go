// SPDX-License-Identifier: AGPL-3.0-or-later
package auth

const (
	AuthScopeNone  = ""
	AuthScopeUser  = "user"
	AuthScopeAdmin = "admin"
	// AuthScopeRead allows only the normal routes that read. Everyone with
	// the user scope has it too.
	AuthScopeRead = "read"
)
