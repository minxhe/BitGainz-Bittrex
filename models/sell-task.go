package models

import (
	"BitGainz-Bittrex/lib"
	bittrex "BitGainz-Bittrex/lib/go-bittrex"
	"BitGainz-Bittrex/lib/utils"

	"github.com/jmoiron/sqlx"
)

// SellTask : schema
// Possible States:
//     1 : ongoing
//     2 : ended
//     3 : ended sold
//     4 : ended with interuption
//     5 : bittrex api busted lmao
type SellTask struct {
	TaskID            int     `db:"task_id"`
	APIKey            string  `db:"apikey"`
	Market            string  `db:"market"`
	BoughtInPrice     float64 `db:"bought_in_price"`
	StartingPrice     float64 `db:"starting_price"`
	CurrentBid        float64
	CurrentLowerBound float64
	Quantity          float64 `db:"quantity"`
	ProfitRange       float64 `db:"profit_range"`
	LowerBound        float64 `db:"lower_bound"`
	State             int     `db:"state"`
	Duration          int     `db:"duration"`
	CreatedAt         string  `db:"created_at"`
}

//CreateTask : register the task in database
func (t SellTask) CreateTask(db *sqlx.DB) int64 {
	newSellTask := `INSERT INTO monitor_sell_tasks (apikey, market, bought_in_price, starting_price, quantity, profit_range, lower_bound, state, duration) VALUES (?, ?, ?, ? ,? ,? ,? ,? ,?)`
	result := db.MustExec(newSellTask, t.APIKey, t.Market, t.BoughtInPrice, t.StartingPrice, t.Quantity, t.ProfitRange, t.LowerBound, t.State, t.Duration)

	id, err := result.LastInsertId()
	utils.HandleErrorCritical(err)

	return id
}

//CheckSellTaskOngoing : check if monotoring for a specific market of key is already on going
func CheckSellTaskOngoing(apikey string, market string, db *sqlx.DB) int {
	rows, err := db.Queryx("SELECT * FROM monitor_sell_tasks WHERE apikey = ? AND market = ?", apikey, market)
	utils.HandleError(err)
	defer rows.Close()

	sellTask := SellTask{}

	for rows.Next() {
		err := rows.StructScan(&sellTask)
		utils.HandleError(err)

		if sellTask.State == 1 {
			return sellTask.TaskID
		}
	}

	return 0
}

// UpdateSellTaskState : updating task state
func UpdateSellTaskState(taskID int64, state int, db *sqlx.DB) {
	stmt, err := db.Prepare("UPDATE monitor_sell_tasks SET state=? WHERE task_id=?")
	utils.HandleError(err)

	defer stmt.Close()

	_, err = stmt.Exec(state, taskID)
	utils.HandleError(err)

}

// GetOngoingSellTasks : get ongoing tasks for a user
func GetOngoingSellTasks(apikey string, db *sqlx.DB) []SellTask {
	b := bittrex.New(apikey, "nil")
	ongoingTasks := []SellTask{}
	onGoingSellTask := SellTask{}

	rows, err := db.Queryx("SELECT * FROM monitor_sell_tasks WHERE apikey=? AND state=?", apikey, 1)
	utils.HandleErrorCritical(err)
	defer rows.Close()

	for rows.Next() {
		err = rows.StructScan(&onGoingSellTask)
		utils.HandleErrorCritical(err)
		ticker, err := b.GetTicker(onGoingSellTask.Market)
		utils.HandleError(err)
		current := ticker.Bid
		currenLowerBound := lib.CalculateLowerBound(onGoingSellTask.BoughtInPrice, current, current*0.95, onGoingSellTask.ProfitRange, onGoingSellTask.LowerBound)
		onGoingSellTask.CurrentBid = current
		onGoingSellTask.CurrentLowerBound = currenLowerBound

		ongoingTasks = append(ongoingTasks, onGoingSellTask)
	}

	return ongoingTasks
}
