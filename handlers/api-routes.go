package handlers

import (
	"github.com/go-chi/chi"
)

// APIRouter : api routing
func APIRouter() chi.Router {
	r := chi.NewRouter()

	r.Post("/start_monitor_sell_session", StartMonitorSellSession)
	r.Post("/start_monitor_buy_session", StartMonitorBuySession)
	r.Post("/end_sell_session", EndSellSession)
	r.Post("/end_buy_session", EndBuySession)
	r.Get("/get_ongoing_sessions", GetOngoingTasks)

	return r
}
