package models

type Language struct {
	Model
	IsoName    string `json:"iso_name"`
	NativeName string `json:"native_name"`
	Code       string `json:"code"`
}
