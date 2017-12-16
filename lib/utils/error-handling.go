package utils

import (
	"fmt"
	"log"

	raven "github.com/getsentry/raven-go"
)

// HandleError : handle panic errors
func HandleError(err error) {
	if err != nil {
		raven.CaptureError(err, nil)
		fmt.Print(err)
	}
}

// HandleErrorCritical : handle critical errors
func HandleErrorCritical(err error) {
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		log.Panic(err)
	}
}

// HandleErrorFatal : handle fatal errors
func HandleErrorFatal(err error) {
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		log.Fatal(err)
	}
}
