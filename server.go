package main

import (
	"net/http"
	"os"

	"BitGainz-Bittrex/handlers"
	"BitGainz-Bittrex/lib/utils"

	l4g "github.com/alecthomas/log4go"
	raven "github.com/getsentry/raven-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
)

func main() {

	// setup the Chi router
	r := chi.NewRouter()

	// Use Chi built-in middlewares
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// When a client closes their connection midway through a request, the
	// http.CloseNotifier will cancel the request context (ctx).
	r.Use(middleware.CloseNotify)

	// Mount endpoints
	r.Mount("/", handlers.APIRouter())

	//run server
	l4g.Info("Starting server at PORT 8080")
	err := http.ListenAndServe(":8080", r)

	utils.HandleErrorFatal(err)

}

func init() {
	raven.SetDSN(os.Getenv("SENTRY_DSN"))
}
