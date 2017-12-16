package handlers

import (
	"BitGainz-Bittrex/api"
	"BitGainz-Bittrex/lib"
	"BitGainz-Bittrex/lib/utils"
	"BitGainz-Bittrex/models"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	bittrex "BitGainz-Bittrex/lib/go-bittrex"

	"github.com/jmoiron/sqlx"
)

type sellSessionInfo struct {
	APIKey      string   `json:"apikey"`
	APISecret   string   `json:"secret"`
	Markets     []string `json:"markets"`
	Duration    int      `json:"duration"`
	Email       string   `json:"email"`
	ProfitRange float64  `json:"profit_range"`
	LowerBound  float64  `json:"lower_bound"`
}

// StartMonitorSellSession : start a monitoring session to sell
func StartMonitorSellSession(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	errMsg := ""

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		utils.HandleErrorCritical(err)
		w.Write([]byte(`{"success": "false"}`))
	}

	session := sellSessionInfo{}
	err = json.Unmarshal(body, &session)
	if err != nil {
		utils.HandleErrorCritical(err)
		w.Write([]byte(`{"success": "false"}`))
	}

	db := models.ConnectSQL(session.Duration)

	b := api.Bittrex{Inst: bittrex.New(session.APIKey, session.APISecret)}

	for _, market := range session.Markets {
		if models.CheckSellTaskOngoing(session.APIKey, market, db) != 0 {
			errMsg = errMsg + "Monitor Sell Task ongoing for user in market: " + market + "\n"
		} else {
			go session.subscribeToMarket(market, b, db)
		}
	}

	w.Write([]byte(`{"success": "true", "message": "` + errMsg + `"}`))
}

// EndSellSession : terminate a currently ongoing session
func EndSellSession(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	errMsg := ""

	body, err := ioutil.ReadAll(req.Body)
	utils.HandleError(err)
	if err != nil {
		w.Write([]byte(`{"success": "false", "message": "invalid request body"}`))
	}

	session := sellSessionInfo{}
	err = json.Unmarshal(body, &session)
	utils.HandleError(err)
	if err != nil {
		w.Write([]byte(`{"success": "false", "message": "invalid request body"}`))
	}

	db := models.ConnectSQL(15)

	for _, market := range session.Markets {
		taskID := models.CheckSellTaskOngoing(session.APIKey, market, db)

		if taskID == 0 {
			errMsg = errMsg + "There is no ongoing sell tasks for market " + market + "\n"
		} else {
			models.UpdateSellTaskState(int64(taskID), 4, db)
		}
	}

	db.Close()

	w.Write([]byte(`{"success": "true", "message": "` + errMsg + `"}`))
}

func (s sellSessionInfo) subscribeToMarket(market string, b api.Bittrex, db *sqlx.DB) {
	ticker, err := b.Inst.GetTicker(market)
	utils.HandleError(err)
	if err != nil {
		return
	}
	currentBid := ticker.Bid
	orderInfo := b.GetBoughtInPriceAndQuantity(market)
	boughtInPrice := orderInfo[0]
	boughtInQuantity := orderInfo[1]

	if boughtInPrice == 0 && boughtInQuantity == 0 {
		log.Println(market, "doest not have any recent transactions to monitor on")
		return
	}

	lowerBound := lib.CalculateLowerBound(boughtInPrice, currentBid, currentBid*0.95, s.ProfitRange, s.LowerBound)

	task := models.SellTask{APIKey: s.APIKey, Market: market, BoughtInPrice: boughtInPrice, StartingPrice: currentBid, Quantity: boughtInQuantity, ProfitRange: s.ProfitRange, LowerBound: s.LowerBound, State: 1, Duration: s.Duration}
	taskID := task.CreateTask(db)

	bid := b.GetCurrentBid(market, taskID, db, s.Email)

	if bid == -1 {
		models.UpdateSellTaskState(taskID, 5, db)
		return
	}

	for i := 0; i < s.Duration/10; i++ {
		log.Println(market, "Bought In Price:", boughtInPrice, "Current Bid:", currentBid, "Current Lower Bound", lowerBound)

		time.Sleep(10 * time.Second)

		if models.CheckSellTaskOngoing(s.APIKey, market, db) == 0 {
			return
		}

		currentBid = b.GetCurrentBid(market, taskID, db, s.Email)
		if currentBid != -1 {
			bid = currentBid
		}

		if bid <= lowerBound {
			//place sell order
			log.Println("Placing sell order at price ", bid)
			success := b.PlaceSellOrder(market, boughtInQuantity, bid, taskID, db, s.Email)
			if success {
				sellOrder := models.SellOrder{APIKey: s.APIKey, Market: market, BoughtInPrice: boughtInPrice, SellOrderPrice: bid, Quantity: boughtInQuantity}
				sellOrder.CreateSellOrder(db)
				lib.SellOrderEmailNotification(s.Email, market, boughtInPrice, bid, taskID)
				models.UpdateSellTaskState(taskID, 3, db)
			}
			return
		}

		if currentBid <= bid {
			currentBid = bid
			lowerBound = lib.CalculateLowerBound(boughtInPrice, currentBid, lowerBound, s.ProfitRange, s.LowerBound)
		} else if currentBid > bid {
			currentBid = bid
		}

	}

	models.UpdateSellTaskState(taskID, 2, db)
}
