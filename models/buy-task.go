package models

import (
	bittrex "BitGainz-Bittrex/lib/go-bittrex"
	"BitGainz-Bittrex/lib/utils"

	"github.com/jmoiron/sqlx"
)

// BuyTask :
// Aims for a target "Last" value then buy in an fixed amount in terms of btc
// Possible States:
//     1 : ongoing
//     2 : ended
//     3 : ended bought
//     4 : ended with interuption
//     5 : bittrex api busted lmao
type BuyTask struct {
	TaskID        int     `db:"task_id"`
	APIKey        string  `db:"apikey"`
	Market        string  `db:"market"`
	TargetPrice   float64 `db:"target_price"`
	StartingPrice float64 `db:"starting_price"`
	CurrentLast   float64
	QuantityBtc   float64 `db:"quantity_btc"`
	State         int     `db:"state"`
	Duration      int     `db:"duration"`
	CreatedAt     string  `db:"created_at"`
}

// CreateTask : create a monitor buy task
func (t BuyTask) CreateTask(db *sqlx.DB) int64 {
	newBuyTask := `INSERT INTO monitor_buy_tasks (apikey, market, target_price, starting_price, quantity_btc, state, duration) VALUES (?, ?, ?, ? ,? ,? ,?)`
	result := db.MustExec(newBuyTask, t.APIKey, t.Market, t.TargetPrice, t.StartingPrice, t.QuantityBtc, t.State, t.Duration)

	id, err := result.LastInsertId()
	utils.HandleErrorCritical(err)

	return id
}

//CheckBuyTaskOngoing : check if monotoring for a specific market of key is already on going
func CheckBuyTaskOngoing(apikey string, market string, db *sqlx.DB) int {
	rows, err := db.Queryx("SELECT * FROM monitor_buy_tasks WHERE apikey = ? AND market = ?", apikey, market)
	utils.HandleError(err)
	defer rows.Close()

	buyTask := BuyTask{}

	for rows.Next() {
		err := rows.StructScan(&buyTask)
		utils.HandleError(err)

		if buyTask.State == 1 {
			return buyTask.TaskID
		}
	}

	return 0
}

// UpdateBuyTaskState : updating task state
func UpdateBuyTaskState(taskID int64, state int, db *sqlx.DB) {
	stmt, err := db.Prepare("UPDATE monitor_buy_tasks SET state=? WHERE task_id=?")
	utils.HandleError(err)

	defer stmt.Close()

	_, err = stmt.Exec(state, taskID)
	utils.HandleError(err)

}

// GetOngoingBuyTasks : get ongoing tasks for a user
func GetOngoingBuyTasks(apikey string, db *sqlx.DB) []BuyTask {
	b := bittrex.New(apikey, "nil")
	ongoingTasks := []BuyTask{}
	onGoingBuyTask := BuyTask{}

	rows, err := db.Queryx("SELECT * FROM monitor_buy_tasks WHERE apikey=? AND state=?", apikey, 1)
	utils.HandleErrorCritical(err)
	defer rows.Close()

	for rows.Next() {
		err = rows.StructScan(&onGoingBuyTask)
		utils.HandleErrorCritical(err)
		ticker, err := b.GetTicker(onGoingBuyTask.Market)
		utils.HandleError(err)
		current := ticker.Last
		onGoingBuyTask.CurrentLast = current

		ongoingTasks = append(ongoingTasks, onGoingBuyTask)
	}

	return ongoingTasks
}
