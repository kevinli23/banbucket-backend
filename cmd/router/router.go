package router

import (
	"banfaucetservice/cmd/handlers"
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

	r.HandleFunc("/api/v1/claim", handlers.ClaimBanano(app)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/v1/donators", handlers.GetDonators(app)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/amount", handlers.GetFaucetAmount(app)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/price", handlers.GetBananoPrice(app)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/payout", handlers.GetBasePayout(app)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/stats", handlers.GetStats(app)).Methods("GET", "OPTIONS")
	return r
}
