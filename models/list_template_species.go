package models

import "time"

type ListTemplateSpecies struct {
	Timestamp      time.Time `json:"timestamp"`
	ListTemplateId int64     `json:"list_template_id"`
	SpeciesId      int64     `json:"species_id"`
}
