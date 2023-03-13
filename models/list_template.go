package models

import "gopkg.in/guregu/null.v4/zero"

type ListTemplate struct {
	Model
	NameId        int64       `json:"-"`
	Name          Translation `json:"name"`
	UserId        int64       `json:"-"`
	User          User        `json:"user"`
	Species       []Species   `json:"species" gorm:"many2many:list_template_species"`
	Visibility    int64       `json:"visibility"`
	Archived      zero.Time   `json:"archived"`
	DescriptionId int64       `json:"-"`
	Description   Translation `json:"description"`
}
