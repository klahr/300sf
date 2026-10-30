// SPDX-License-Identifier: AGPL-3.0-or-later

// Package seed holds starter data for a new installation: the bird orders and
// every species accepted as observed in Sweden, named in Swedish, English,
// German and Latin. generate.py builds the CSV files; see it for the sources.
package seed

import (
	"bytes"
	_ "embed"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/table"
	"github.com/klahr/300sf/server/utils"
	"gopkg.in/guregu/null.v4/zero"
	"gorm.io/gorm"
)

//go:embed orders.csv
var ordersCSV []byte

//go:embed species.csv
var speciesCSV []byte

type Bird struct {
	Order string
	La    string
	Sv    string
	En    string
	De    string
}

// Orders and Species read the embedded files; an order's Order is empty.
func Orders() ([]Bird, error) {
	return read(ordersCSV, []string{"la", "sv", "en", "de"}, func(r []string) Bird {
		return Bird{La: r[0], Sv: r[1], En: r[2], De: r[3]}
	})
}

func Species() ([]Bird, error) {
	return read(speciesCSV, []string{"order", "la", "sv", "en", "de"}, func(r []string) Bird {
		return Bird{Order: r[0], La: r[1], Sv: r[2], En: r[3], De: r[4]}
	})
}

func read(data []byte, header []string, bird func([]string) Bird) ([]Bird, error) {
	rows, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 || fmt.Sprint(rows[0]) != fmt.Sprint(header) {
		return nil, fmt.Errorf("want columns %v", header)
	}
	birds := make([]Bird, 0, len(rows)-1)
	for _, r := range rows[1:] {
		birds = append(birds, bird(r))
	}
	return birds, nil
}

// Load adds the orders and species the database doesn't have yet, matched on
// their Latin names, so it can run again after an update or on a database
// that already has some. It reports how many of each it added.
func Load(db *gorm.DB) (orders, species int, err error) {
	allOrders, err := Orders()
	if err != nil {
		return 0, 0, err
	}
	allSpecies, err := Species()
	if err != nil {
		return 0, 0, err
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		orderIds, err := existing(tx, table.Orders)
		if err != nil {
			return err
		}
		for _, o := range allOrders {
			if _, ok := orderIds[o.La]; ok {
				continue
			}
			name, err := createName(tx, o)
			if err != nil {
				return err
			}
			order := models.Order{Model: newModel(), NameId: name}
			if err := tx.Omit("Name").Create(&order).Error; err != nil {
				return err
			}
			orderIds[o.La] = order.Id
			orders++
		}

		speciesIds, err := existing(tx, table.Species)
		if err != nil {
			return err
		}
		for _, s := range allSpecies {
			if _, ok := speciesIds[s.La]; ok {
				continue
			}
			orderId, ok := orderIds[s.Order]
			if !ok {
				return fmt.Errorf("%s: no order %s", s.La, s.Order)
			}
			name, err := createName(tx, s)
			if err != nil {
				return err
			}
			row := models.Species{Model: newModel(), NameId: name, OrderId: orderId}
			if err := tx.Omit("Name", "Order").Create(&row).Error; err != nil {
				return err
			}
			species++
		}
		return nil
	})
	return orders, species, err
}

// existing maps the Latin names of a table's rows to their ids.
func existing(tx *gorm.DB, tableName string) (map[string]int64, error) {
	var rows []struct {
		Id int64
		La string
	}
	err := tx.Table(tableName).Select(tableName + ".id, " + table.Translations + ".la").
		Joins("JOIN " + table.Translations + " ON " + table.Translations + ".id=" + tableName + ".name_id").
		Scan(&rows).Error
	ids := make(map[string]int64, len(rows))
	for _, r := range rows {
		ids[r.La] = r.Id
	}
	return ids, err
}

func createName(tx *gorm.DB, b Bird) (int64, error) {
	name := models.Translation{
		Model: newModel(),
		La:    zero.StringFrom(b.La),
		Sv:    zero.StringFrom(b.Sv),
		En:    zero.StringFrom(b.En),
		De:    zero.StringFrom(b.De),
	}
	err := tx.Create(&name).Error
	return name.Id, err
}

func newModel() models.Model {
	return models.Model{Timestamp: time.Now(), Uuid: utils.GenerateShortUuid()}
}
