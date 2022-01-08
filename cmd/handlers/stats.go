package handlers

import (
	"banfaucetservice/pkg/app"
	"encoding/json"
	"net/http"
	"strconv"
)

func GetStats(app *app.App) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(app.FBHandler.CachedStats)
	}
}

func GetTotalClaims(app *app.App) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strconv.FormatUint(uint64(app.FBHandler.CachedStats.TotalClaims), 10)))
	}
}
