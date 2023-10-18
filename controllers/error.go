// SPDX-License-Identifier: GPL-2.0-or-later
package controllers

const (
	ErrorInternalServerError = 1000
	ErrorRecordNotFound      = 1001
	ErrorInsert              = 1002
	ErrorDelete              = 1003

	ErrorParse       = 2001
	ErrorLinkExpired = 2002

	ErrorLoggedOut         = 3001
	ErrorMissingApiKey     = 3002
	ErrorInvalidApiKey     = 3003
	ErrorTooManyApiCalls   = 3004
	ErrorMissingPermission = 3005

	ErrorSignUpInvalidDisplayName = 4001
	ErrorSignUpInvalidEmail       = 4002
	ErrorSignUpInvalidPassword    = 4003
	ErrorSignUpTermsNotAccepted   = 4004

	ErrorMigrateCompleted          = 5001
	ErrorMigrateInvalidCredentials = 5002
)
