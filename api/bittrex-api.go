package api

import (
	"BitGainz-Bittrex/lib"
	"BitGainz-Bittrex/lib/utils"
	"BitGainz-Bittrex/models"

	bittrex "BitGainz-Bittrex/lib/go-bittrex"

	"github.com/jmoiron/sqlx"
)

// Bittrex instance
type Bittrex struct {
	Inst *bittrex.Bittrex
}

// GetCurrentBid : returns the current bid value for a specific market
func (b Bittrex) GetCurrentBid(market string, taskID int64, db *sqlx.DB, email string) float64 {

	ticker, err := b.Inst.GetTicker(market)
	if err != nil {
		utils.HandleError(err)
		return -1
	}

	return ticker.Bid
}

// GetCurrentLast : returns the current last value for a specific market
func (b Bittrex) GetCurrentLast(market string, taskID int64, db *sqlx.DB) float64 {
	ticker, err := b.Inst.GetTicker(market)
	if err != nil {
		utils.HandleError(err)
		return -1
	}

	return ticker.Last
}

// GetBalance : get user's balance for a specific currency
func (b Bittrex) GetBalance(currency string) float64 {
	balance, err := b.Inst.GetBalance(currency)
	utils.HandleErrorCritical(err)

	return balance.Available
}

// GetBoughtInPriceAndQuantity : get the bought in unit price for the lastest purchase for a single market
// It on fetches the most recent LIMIT_BUY order
func (b Bittrex) GetBoughtInPriceAndQuantity(market string) [2]float64 {
	orderHistory, err := b.Inst.GetOrderHistory(market)
	utils.HandleError(err)

	for _, singleOrder := range orderHistory {
		if singleOrder.OrderType == "LIMIT_SELL" {
			return [2]float64{0, 0}
		} else if singleOrder.OrderType == "LIMIT_BUY" {

			uuid := singleOrder.OrderUuid
			order, err := b.Inst.GetOrder(uuid)
			utils.HandleError(err)

			if order.IsOpen == false {
				return [2]float64{order.PricePerUnit, order.Quantity}
			}
		}

	}
	return [2]float64{0, 0}
}

// PlaceSellOrder : place sell order xp boom
func (b Bittrex) PlaceSellOrder(market string, quantity float64, unitPrice float64, taskID int64, db *sqlx.DB, email string) bool {
	_, err := b.Inst.SellLimit(market, quantity, unitPrice)
	if err != nil {
		models.UpdateSellTaskState(taskID, 5, db)
		lib.BittrexErrorEmailNotification(email, market, taskID, err.Error())
		utils.HandleError(err)
		return false
	}
	return true
}

// PlaceBuyOrder : placing buy order
func (b Bittrex) PlaceBuyOrder(market string, quantity float64, unitPrice float64, taskID int64, db *sqlx.DB, email string) string {
	uuid, err := b.Inst.BuyLimit(market, quantity, unitPrice)
	if err != nil {
		models.UpdateBuyTaskState(taskID, 5, db)
		lib.BittrexErrorEmailNotification(email, market, taskID, err.Error())
		utils.HandleError(err)
		return "fail"
	}
	return uuid
}

// CancelOrder : cancel a buy or sell order
func (b Bittrex) CancelOrder(uuid string) {
	err := b.Inst.CancelOrder(uuid)
	utils.HandleError(err)
}
