package utils

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type EmailTemplate struct {
	Subject string `yaml:"subject"`
	Email   string `yaml:"email"`
	Name    string `yaml:"name"`
	Txt     string `yaml:"txt"`
	Html    string `yaml:"html"`
}

func GetEmailTemplate(name, lang string) (*EmailTemplate, error) {
	filename := fmt.Sprintf("email-templates/%s.%s.yml", name, lang)
	raw, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var e EmailTemplate
	if err := yaml.Unmarshal(raw, &e); err != nil {
		return nil, err
	}
	return &e, nil
}
