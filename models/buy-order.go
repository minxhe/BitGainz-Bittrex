package models

import (
	"BitGainz-Bittrex/lib/utils"

	"github.com/jmoiron/sqlx"
)

// BuyOrder : sell_orders schema
type BuyOrder struct {
	APIKey        string
	Market        string
	BuyOrderPrice float64
	QuantityBtc   float64
}

// CreateBuyOrder : new sell order insert to db
func (b BuyOrder) CreateBuyOrder(db *sqlx.DB) {

	stmt, err := db.Prepare("INSERT buy_orders SET apikey=?,market=?,buy_order_price=?, quantity_btc=?")
	utils.HandleError(err)

	defer stmt.Close()

	res, err := stmt.Exec(b.APIKey, b.Market, b.BuyOrderPrice, b.QuantityBtc)
	utils.HandleError(err)

	_, err = res.LastInsertId()
	utils.HandleError(err)

}
