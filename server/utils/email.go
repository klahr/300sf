// SPDX-License-Identifier: AGPL-3.0-or-later
package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

// emailDir holds layout.html and, per language, common.yml and one file per
// email with its text.
const emailDir = "email-templates"

// EmailTemplate is a rendered email, ready to store and send.
type EmailTemplate struct {
	Subject string
	Email   string // sender address
	Name    string // sender name
	ReplyTo string
	Txt     string
	Html    string
}

type emailCommon struct {
	AppName        string `yaml:"app_name"`
	ButtonFallback string `yaml:"button_fallback"`
	Privacy        string `yaml:"privacy"`
	Support        string `yaml:"support"`
	Company        string `yaml:"company"`
}

type emailContent struct {
	Subject   string   `yaml:"subject"`
	Email     string   `yaml:"email"`
	Name      string   `yaml:"name"`
	Preheader string   `yaml:"preheader"`
	Heading   string   `yaml:"heading"`
	Body      []string `yaml:"body"`
	Button    string   `yaml:"button"`
	Link      string   `yaml:"link"`
	After     []string `yaml:"after"`
	// ShowLink prints the link under the button too, for links that only work
	// once and must not be lost if the button can't be clicked.
	ShowLink  bool   `yaml:"show_link"`
	Signature string `yaml:"signature"`
	Reason    string `yaml:"reason"`
}

func readYaml(path string, v any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(raw, v)
}

// GetEmailTemplate renders email name in lang (English if lang has no such
// email). Every ${KEY} in its text is replaced by vars[KEY]; ${DOMAIN} is the
// site's base URL. Values are escaped in the HTML, so they can't add markup.
func GetEmailTemplate(name, lang string, domain string, vars map[string]string) (*EmailTemplate, error) {
	if lang == "" || strings.ContainsAny(lang, `/\.`) {
		lang = "en"
	}
	if _, err := os.Stat(filepath.Join(emailDir, lang, name+".yml")); err != nil {
		lang = "en"
	}

	var common emailCommon
	if err := readYaml(filepath.Join(emailDir, lang, "common.yml"), &common); err != nil {
		return nil, err
	}
	var c emailContent
	if err := readYaml(filepath.Join(emailDir, lang, name+".yml"), &c); err != nil {
		return nil, err
	}

	pairs := []string{"${DOMAIN}", domain}
	for k, v := range vars {
		pairs = append(pairs, "${"+k+"}", v)
	}
	r := strings.NewReplacer(pairs...)
	expand := func(list []string) []string {
		out := make([]string, len(list))
		for i, s := range list {
			out[i] = r.Replace(s)
		}
		return out
	}
	c.Subject, c.Preheader, c.Heading = r.Replace(c.Subject), r.Replace(c.Preheader), r.Replace(c.Heading)
	c.Body, c.After = expand(c.Body), expand(c.After)
	c.Link, c.Signature, c.Reason = r.Replace(c.Link), r.Replace(c.Signature), r.Replace(c.Reason)

	layout, err := template.ParseFiles(filepath.Join(emailDir, "layout.html"))
	if err != nil {
		return nil, err
	}
	var html bytes.Buffer
	if err := layout.Execute(&html, map[string]any{
		"Lang": lang, "Domain": domain, "AppName": common.AppName, "Common": common,
		"Subject": c.Subject, "Preheader": c.Preheader, "Heading": c.Heading, "Body": c.Body,
		"Button": c.Button, "Link": c.Link, "After": c.After,
		"Signature": c.Signature, "Reason": c.Reason, "ShowLink": c.ShowLink,
	}); err != nil {
		return nil, err
	}

	return &EmailTemplate{
		Subject: c.Subject,
		Email:   c.Email,
		Name:    c.Name,
		ReplyTo: common.Support,
		Txt:     emailText(common, c, domain),
		Html:    html.String(),
	}, nil
}

// emailText is the plain-text part, for clients that don't show HTML.
func emailText(common emailCommon, c emailContent, domain string) string {
	var b strings.Builder
	b.WriteString(c.Heading + "\n\n")
	for _, p := range c.Body {
		b.WriteString(p + "\n\n")
	}
	if c.Button != "" {
		fmt.Fprintf(&b, "%s:\n%s\n\n", c.Button, c.Link)
	}
	for _, p := range c.After {
		b.WriteString(p + "\n\n")
	}
	if c.Signature != "" {
		b.WriteString(c.Signature + "\n\n")
	}
	fmt.Fprintf(&b, "-- \n%s\n%s %s/\n%s\n", c.Reason, common.AppName, domain, common.Company)
	return b.String()
}
