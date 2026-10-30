// SPDX-License-Identifier: AGPL-3.0-or-later
package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/klahr/300sf/server/auth"
	"github.com/klahr/300sf/server/config"
	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/utils"
)

// bootstrap sets up a new installation, where nobody can sign in yet: the
// website signs in with an API key, and API keys are made by signed-in users.
// It makes the user with this email (who signed up on the website) an admin
// and, unless site.client-id is set, makes the website's API key, owned by
// that user. The key is printed once, for the config.
func bootstrap(email string) error {
	var user models.User
	if err := models.DB.Where("email=?", email).Where("deleted_at IS NULL").First(&user).Error; err != nil {
		return fmt.Errorf("no account for %s: sign up on the website first", email)
	}
	if !user.HasScope(auth.AuthScopeAdmin) {
		scopes := strings.TrimSpace(user.Scopes + " " + auth.AuthScopeAdmin)
		if err := models.DB.Model(&user).Update("scopes", scopes).Error; err != nil {
			return err
		}
	}
	fmt.Printf("%s is an admin.\n", email)

	if config.AppConfig.Site.ClientId != "" {
		fmt.Println("site.client-id is already set; no new website key made.")
		return nil
	}

	secret := utils.RandomSha256()
	prefix := utils.GenerateUuidWithFixedLength(7)
	key := models.ApiKey{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateShortUuid(),
		},
		Name:     "Website",
		ClientId: utils.GenerateUuidWithFixedLength(32),
		Prefix:   prefix,
		Hash:     utils.ComputeHmac256(secret, config.AppConfig.HashingSecret),
		UserId:   user.Id,
		// The website signs in admins too; a session gets the scopes both
		// the user and the key have.
		Scopes:    auth.AuthScopeUser + " " + auth.AuthScopeAdmin,
		SoftLimit: config.AppConfig.DefaultApiSoftLimit,
		HardLimit: config.AppConfig.DefaultApiHardLimit,
	}
	if err := models.DB.Create(&key).Error; err != nil {
		return err
	}
	fmt.Printf("\nAdd to the config's site: section, then restart:\n\n  client-id: %q\n  client-secret: %q\n", key.ClientId, prefix+"."+secret)
	return nil
}
