package models

type Resource struct {
	Model
	Type     string `json:"type"`
	Mimetype string `json:"mimetype"`
	Filename string `json:"filename"`
	Url      string `json:"url"`
}
