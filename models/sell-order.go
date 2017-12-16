package models

import (
	"BitGainz-Bittrex/lib/utils"

	"github.com/jmoiron/sqlx"
)

// SellOrder : sell_orders schema
type SellOrder struct {
	APIKey         string
	Market         string
	BoughtInPrice  float64
	SellOrderPrice float64
	Quantity       float64
}

// CreateSellOrder : new sell order insert to db
func (s SellOrder) CreateSellOrder(db *sqlx.DB) {

	stmt, err := db.Prepare("INSERT sell_orders SET apikey=?,market=?,bought_in_price=?, sell_order_price=?, quantity=?")
	utils.HandleError(err)

	defer stmt.Close()

	res, err := stmt.Exec(s.APIKey, s.Market, s.BoughtInPrice, s.SellOrderPrice, s.Quantity)
	utils.HandleError(err)

	_, err = res.LastInsertId()
	utils.HandleError(err)

}
