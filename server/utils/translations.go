// SPDX-License-Identifier: GPL-2.0-or-later
package utils

import "github.com/klahr/300sf/server/models"

func Translation2Value(t models.Translation, lang string) string {
	switch lang {
	case "sv":
		return t.Sv.String
	case "en":
		return t.En.String
	case "de":
		return t.De.String
	case "la":
		return t.La.String
	}

	return ""
}
