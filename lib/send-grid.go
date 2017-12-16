package lib

import (
	"BitGainz-Bittrex/lib/utils"
	"os"
	"strconv"
	"time"

	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var from = mail.NewEmail("JeltaFund Monitoring", "jeltafund@gmail.com")

type emailNotif struct {
	email            string
	subject          string
	plainTextContent string
	htmlContent      string
}

// SellOrderEmailNotification : give user email notification
func SellOrderEmailNotification(email string, market string, boughtIn float64, soldAt float64, taskID int64) {
	subject := "Sell Order Placed For Market " + market + " at " + time.Now().Format("2006-01-02 15:04:05")
	plainTextContent := "Task " + string(taskID) + " placed LIMIT_SELL order for market " + market + " at " + time.Now().Format("2006-01-02 15:04:05") + "\n" + "Bought In Price: " + strconv.FormatFloat(boughtIn, 'f', -1, 64) + "\n" + "Sold At Price: " + strconv.FormatFloat(soldAt, 'f', -1, 64)
	htmlContent := "<p>" + plainTextContent + "</p>"
	e := emailNotif{email: email, subject: subject, plainTextContent: plainTextContent, htmlContent: htmlContent}
	e.send()
}

// BuyOrderEmailNotification : give user email notification
func BuyOrderEmailNotification(email string, market string, boughtIn float64, taskID int64) {
	subject := "Buy Order Placed For Market " + market + " at " + time.Now().Format("2006-01-02 15:04:05")
	plainTextContent := "Task " + string(taskID) + " placed LIMIT_BUY order for market " + market + " at " + time.Now().Format("2006-01-02 15:04:05") + "\n" + "Bought At Price: " + strconv.FormatFloat(boughtIn, 'f', -1, 64) + "\n"
	htmlContent := "<p>" + plainTextContent + "</p>"
	e := emailNotif{email: email, subject: subject, plainTextContent: plainTextContent, htmlContent: htmlContent}
	e.send()
}

// BittrexErrorEmailNotification : when bittrex errors out
func BittrexErrorEmailNotification(email string, market string, taskID int64, errors string) {
	subject := "Bittrex Api Error for " + market + " at " + time.Now().Format("2006-01-02 15:04:05")
	plainTextContent := "Task " + string(taskID) + " errored due to an internal error from Bittrex API, the task is now terminated " + "Error trace: " + errors
	htmlContent := "<p>Task " + string(taskID) + " errored due to an internal error from Bittrex API, the task is now terminated <br /><br />" + "Error trace: <br />" + errors + "</p>"
	e := emailNotif{email: email, subject: subject, plainTextContent: plainTextContent, htmlContent: htmlContent}
	e.send()
}

func (e *emailNotif) send() {
	to := mail.NewEmail("User", e.email)
	message := mail.NewSingleEmail(from, e.subject, to, e.plainTextContent, e.htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SEND_GRID_KEY"))
	_, err := client.Send(message)
	utils.HandleError(err)
}
