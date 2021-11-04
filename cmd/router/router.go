package router

import (
	"banfaucetservice/cmd/middleware"
	"banfaucetservice/pkg/app"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/newrelic/go-agent/v3/integrations/nrgorilla"
)

func Get(app *app.App) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("🍌🐒🍌"))
	})

	r.Use(nrgorilla.Middleware(app.NewRelicApp))
	r.Use(middleware.CORSMiddleware)

	r.HandleFunc("/api/v1/claim", middleware.ClaimBanano(app)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/v1/donators", middleware.GetDonators(app)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/amount", middleware.GetFaucetAmount(app)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/price", middleware.GetBananoPrice(app)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/payout", middleware.GetBasePayout(app)).Methods("GET", "OPTIONS")
	return r
}
