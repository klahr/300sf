package models

type SubscriptionSpecies struct {
	Id             int64 `json:"-"`
	SubscriptionId int64 `json:"-"`
	SpeciesId      int64 `json:"-"`
}
