// SPDX-License-Identifier: GPL-2.0-or-later
package models

import "gopkg.in/guregu/null.v4/zero"

type Translation struct {
	Model
	Sv zero.String `json:"sv"`
	En zero.String `json:"en"`
	La zero.String `json:"la"`
}
