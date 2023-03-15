package utils

import "git.sr.ht/~klahr/300sf/models"

func Translation2Value(t models.Translation, lang string) string {
	switch lang {
	case "sv":
		return t.Sv.String
	case "en":
		return t.En.String
	case "la":
		return t.La.String
	}

	return ""
}
