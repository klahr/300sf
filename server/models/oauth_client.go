// SPDX-License-Identifier: AGPL-3.0-or-later
package models

// OauthClient is an app that registered itself to let users connect it to
// their account with OAuth, such as Claude. Each user who connects it gets
// an ApiKey of their own pointing back here.
type OauthClient struct {
	Model
	ClientId string `json:"client_id"`
	Name     string `json:"name"`
	// RedirectUris is a JSON array of the addresses the app may be sent back to.
	RedirectUris string `json:"-"`
}
