package handlers

import (
	"BitGainz-Bittrex/api"
	"BitGainz-Bittrex/lib"
	bittrex "BitGainz-Bittrex/lib/go-bittrex"
	"BitGainz-Bittrex/lib/utils"
	"BitGainz-Bittrex/models"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
)

type buySessionInfo struct {
	APIKey      string  `json:"apikey"`
	APISecret   string  `json:"secret"`
	Market      string  `json:"markets"`
	Duration    int     `json:"duration"`
	Email       string  `json:"email"`
	TargetPrice float64 `json:"target_price"`
	QuantityBtc float64 `json:"quantity_btc"`
	ProfitRange float64 `json:"profit_range"`
	LowerBound  float64 `json:"lower_bound"`
}

// StartMonitorBuySession : start a monitoring session to buy
func StartMonitorBuySession(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	errMsg := ""

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		utils.HandleErrorCritical(err)
		w.Write([]byte(`{"success": "false"}`))
	}

	buySession := buySessionInfo{}
	err = json.Unmarshal(body, &buySession)
	if err != nil {
		utils.HandleErrorCritical(err)
		w.Write([]byte(`{"success": "false"}`))
	}

	db := models.ConnectSQL(buySession.Duration * 2)
	b := api.Bittrex{Inst: bittrex.New(buySession.APIKey, buySession.APISecret)}

	var markets []string
	markets = append(markets, buySession.Market)
	sellSession := sellSessionInfo{APIKey: buySession.APIKey, APISecret: buySession.APISecret, Markets: markets, Duration: buySession.Duration, Email: buySession.Email, ProfitRange: buySession.ProfitRange, LowerBound: buySession.LowerBound}

	if models.CheckBuyTaskOngoing(buySession.APIKey, buySession.Market, db) != 0 {
		errMsg = errMsg + "Monitor Buy Task ongoing for user in market: " + buySession.Market + "\n"
	} else {
		go buySession.subscribeToMarket(sellSession, b, db)
	}

	w.Write([]byte(`{"success": "true", "message": "` + errMsg + `"}`))
}

// EndBuySession : terminate a currently ongoing session
func EndBuySession(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	errMsg := ""

	body, err := ioutil.ReadAll(req.Body)
	utils.HandleError(err)
	if err != nil {
		w.Write([]byte(`{"success": "false", "message": "invalid request body"}`))
	}

	// Hacky here should be able to use array
	session := sellSessionInfo{}
	err = json.Unmarshal(body, &session)
	utils.HandleError(err)
	if err != nil {
		w.Write([]byte(`{"success": "false", "message": "invalid request body"}`))
	}

	db := models.ConnectSQL(15)

	for _, market := range session.Markets {
		taskID := models.CheckBuyTaskOngoing(session.APIKey, market, db)

		if taskID == 0 {
			errMsg = errMsg + "There is no ongoing buy tasks for market " + market + "\n"
		} else {
			models.UpdateBuyTaskState(int64(taskID), 4, db)
		}
	}

	db.Close()

	w.Write([]byte(`{"success": "true", "message": "` + errMsg + `"}`))
}

func (s buySessionInfo) subscribeToMarket(sellSession sellSessionInfo, b api.Bittrex, db *sqlx.DB) {
	ticker, err := b.Inst.GetTicker(s.Market)
	utils.HandleError(err)
	if err != nil {
		return
	}
	currentLast := ticker.Last

	buyTask := models.BuyTask{APIKey: s.APIKey, Market: s.Market, TargetPrice: s.TargetPrice, StartingPrice: currentLast, QuantityBtc: s.QuantityBtc, State: 1, Duration: s.Duration}
	buyTaskID := buyTask.CreateTask(db)

	last := b.GetCurrentLast(s.Market, buyTaskID, db)

	if last == -1 {
		models.UpdateBuyTaskState(buyTaskID, 5, db)
		return
	}

	for i := 0; i < s.Duration/10; i++ {
		log.Println(s.Market, "Target Last:", s.TargetPrice, "Current Last:", currentLast)

		time.Sleep(10 * time.Second)

		if models.CheckBuyTaskOngoing(s.APIKey, s.Market, db) == 0 {
			return
		}

		currentLast = b.GetCurrentLast(s.Market, buyTaskID, db)
		if currentLast != -1 {
			last = currentLast
		}

		if last < s.TargetPrice {
			// place buy order
			log.Println("Placing buy order at price ", last)
			availableBtc := b.GetBalance("BTC")
			if availableBtc < s.QuantityBtc {
				s.QuantityBtc = availableBtc
			}
			quantity := s.QuantityBtc / last
			uuid := b.PlaceBuyOrder(s.Market, quantity, last, buyTaskID, db, s.Email)
			if uuid != "fail" {
				buyOrder := models.BuyOrder{APIKey: s.APIKey, Market: s.Market, BuyOrderPrice: last, QuantityBtc: s.QuantityBtc}
				buyOrder.CreateBuyOrder(db)
				models.UpdateBuyTaskState(buyTaskID, 3, db)
				lib.BuyOrderEmailNotification(s.Email, s.Market, last, buyTaskID)

				// wait for the order to go through
				time.Sleep(60 * time.Second)
				b.CancelOrder(uuid)

				// start sell session
				sellSession.subscribeToMarket(s.Market, b, db)
			}

			return
		}
	}
}
