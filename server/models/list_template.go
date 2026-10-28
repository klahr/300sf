// SPDX-License-Identifier: AGPL-3.0-or-later
package models

import "gopkg.in/guregu/null.v4/zero"

type ListTemplate struct {
	Model
	NameId        int64       `json:"-"`
	Name          Translation `json:"name"`
	UserId        int64       `json:"-"`
	Species       []Species   `json:"species" gorm:"many2many:list_template_species"`
	Visibility    int64       `json:"visibility"`
	Archived      zero.Time   `json:"-"`
	DescriptionId int64       `json:"-"`
	Description   Translation `json:"description"`
}
