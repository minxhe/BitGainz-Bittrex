package handlers

import (
	"BitGainz-Bittrex/lib/utils"
	"BitGainz-Bittrex/models"
	"encoding/json"
	"net/http"
)

type userInfo struct {
	APIKey  string   `json:"apikey"`
	Markets []string `json:"markets"`
}

type ongoingTasks struct {
	SellTasks []models.SellTask
	BuyTasks  []models.BuyTask
}

// GetOngoingTasks : get ongoing tasks
func GetOngoingTasks(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	apikey := req.URL.Query().Get("apikey")
	if apikey == "" {
		w.Write([]byte(`{"success": "false", "message": "invalid apikey"}`))
	}
	db := models.ConnectSQL(20)

	tasks := ongoingTasks{SellTasks: models.GetOngoingSellTasks(apikey, db), BuyTasks: models.GetOngoingBuyTasks(apikey, db)}
	ongoingTasks, err := json.Marshal(tasks)
	utils.HandleError(err)

	db.Close()

	w.Write(ongoingTasks)
}

// GetCompletedTasks : get completed tasks
func GetCompletedTasks(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
}
