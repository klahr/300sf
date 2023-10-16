package utils

import (
	"fmt"
	"os"
)

type EmailTemplate struct {
	Html string
	Txt  string
}

func GetEmailTemplate(name, lang string) (*EmailTemplate, error) {
	txt, err := os.ReadFile(fmt.Sprintf("email-templates/%s.%s.txt", name, lang))
	if err != nil {
		return nil, err
	}
	html, err := os.ReadFile(fmt.Sprintf("email-templates/%s.%s.html", name, lang))
	if err != nil {
		return nil, err
	}
	return &EmailTemplate{
		Html: string(html),
		Txt:  string(txt),
	}, nil
}
