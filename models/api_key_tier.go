package models

type ApiKeyTier struct {
	Model
	Name      string `json:"name"`
	SoftLimit int64  `json:"soft_limit"`
	HardLimit int64  `json:"hard_limit"`
}
