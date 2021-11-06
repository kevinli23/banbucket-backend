package middleware

import (
	"banfaucetservice/cmd/models"
	"banfaucetservice/pkg/app"
	"banfaucetservice/pkg/banano"
	"banfaucetservice/pkg/logger"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/BananoCoin/gobanano/nano"
	"github.com/pkg/errors"
)

const (
	torValidationAddr = "127.0.0.2"
	torDNSEL          = "%s.dnsel.torproject.org"

	VERIFY_HCAPTCHA_FLAG = true
	CUSTOM_IP_FLAG       = false
	IP_INTEL_FLAG        = false
	CLAIM_INTERVAL       = 15
)

func GetBasePayout(app *app.App) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		balance, err := nano.ParseBalance(banano.AMOUNT_TO_SEND, "raw")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{Message: errors.Wrapf(err, "Failed to compute base payout").Error()})
			return
		}
		amountFloatString, _ := strconv.ParseFloat(balance.String(), 64)
		w.WriteHeader(http.StatusAccepted)

		floatString := fmt.Sprintf("%.2f", amountFloatString*10)

		json.NewEncoder(w).Encode(Response{
			Message: floatString,
		})
	}
}

func GetBananoPrice(app *app.App) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(PriceResponse{
			Price:  fmt.Sprintf("%.5f", app.Price),
			Change: fmt.Sprintf("%.3f", app.PriceChange),
		})
	}
}

func GetFaucetAmount(app *app.App) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		f, _ := strconv.ParseFloat(app.Amount.String(), 64)
		json.NewEncoder(w).Encode(Response{Message: fmt.Sprintf("%.2f", f*10)})
	}
}

func GetDonators(app *app.App) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var donators []models.Donator
		var err error
		if len(app.Donators.Donators) > 0 {
			donators = app.Donators.Donators
		} else {
			donators, err = app.GetAllDonators()
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(Response{Message: errors.Wrapf(err, "Failed to retrieve list of donators").Error()})
				return
			}
			app.Donators = models.AllDonators{Donators: donators}
		}

		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(models.AllDonators{
			Donators: donators,
		})
	}
}

func ClaimBanano(app *app.App) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		var claimReq models.ClaimRequest

		if err := json.NewDecoder(r.Body).Decode(&claimReq); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{Message: errors.Wrapf(err, "Unable to decode the request body").Error()})
			return
		}

		if !CUSTOM_IP_FLAG {
			ip := r.Header.Get("X-REAL-IP")
			netIP := net.ParseIP(ip)

			if netIP == nil {
				ips := r.Header.Get("X-FORWARDED-FOR")
				ip = strings.Split(ips, ",")[0]
				netIP = net.ParseIP(ip)

				if netIP == nil {
					remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
					if err != nil {
						w.WriteHeader(http.StatusBadRequest)
						json.NewEncoder(w).Encode(Response{Message: "Request sent from an invalid IP"})
						return
					}

					netIP = net.ParseIP(remoteIP)
					if netIP == nil {
						w.WriteHeader(http.StatusBadRequest)
						json.NewEncoder(w).Encode(Response{Message: "Request sent from an invalid IP"})
						return
					}

					ip = remoteIP
				}
			}

			claimReq.IP = ip
		}

		isBanned, err := app.IsBanned(claimReq.BananoAddr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{Message: "Something went wrong..."})
			return
		}

		if isBanned {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{Message: "Sorry! BanBucket has identified you as an abuser of this faucet. If you think this is incorrect please reach out on reddit."})
			return
		}

		if strings.HasPrefix(claimReq.IP, "51.79") || strings.HasPrefix(claimReq.IP, "128.90") || strings.HasPrefix(claimReq.IP, "31.6") {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{Message: "Sorry! BanBucket doesn't support faucet claims from your region at this time due to abuse."})
			return
		}

		if isTorIP(claimReq.IP) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{Message: "Sorry! BanBucket doesn't support faucet claims from a TOR IP at this time."})
			return
		}

		isVPN, err := app.IsVPN(claimReq.IP)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{Message: "Something went wrong..."})
			return
		}
		if isVPN {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{Message: "Sorry! BanBucket doesn't support faucet claims from a VPN at this time."})
			return
		}

		valid, err := verifyHCaptcha(claimReq.Captcha, app.Config.GetHCapthcaSecret())
		if err != nil || !valid {
			logger.Error.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{Message: "Invalid HCaptcha verification"})
			return
		}

		lastClaims, err := getLastClaim(app, claimReq.BananoAddr, claimReq.IP, r.Context())
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{Message: errors.Wrapf(err, "Failed to retrieve last claim details").Error()})
			return
		}

		var lastClaim time.Time
		newClaim := len(lastClaims) == 0

		if len(lastClaims) >= 2 {
			for _, lClaim := range lastClaims {
				if time.Unix(lClaim.LastClaim, 0).After(lastClaim) {
					lastClaim = time.Unix(lClaim.LastClaim, 0)
				}
			}
		} else if len(lastClaims) == 1 {
			lastClaim = time.Unix(lastClaims[0].LastClaim, 0)
		}

		diff := time.Since(lastClaim)

		if diff.Hours() >= CLAIM_INTERVAL {

			// This is called here to reduce the # of API calls (limit of 500 daily)
			if IP_INTEL_FLAG {
				if isMaliciousIP := IsMaliciousIP(claimReq.IP); isMaliciousIP {
					// TODO: Save IP to Blacklist

					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(Response{Message: "Your IP is deemed to be malicious, sorry for the inconvenience"})
					return
				}
			}

			var err error

			hash, amountGiven, err := banano.SendBanano(claimReq.BananoAddr, app)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(Response{Message: errors.Wrapf(err, "Failed to send banano").Error()})
				return
			}

			logger.Info.Printf("Sent banano to %s with hash %s\n", claimReq.BananoAddr, hash)

			// update db
			if newClaim {
				err = insertClaim(app, claimReq)
			} else {
				err = updateClaims(app, claimReq)
			}

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(Response{Message: errors.Wrapf(err, "Failed to save claim to DB").Error()})
				return
			}

			amountFloatString, _ := strconv.ParseFloat(amountGiven.String(), 64)
			w.WriteHeader(http.StatusAccepted)

			floatString := fmt.Sprintf("%.2f", amountFloatString*10)

			if floatString == "0.02" {
				json.NewEncoder(w).Encode(Response{Message: fmt.Sprintf("Success! %s BAN has been sent your way. Consider switching your representative away from Kalium or BananoVault to earn more!", floatString)})
				return
			}

			json.NewEncoder(w).Encode(Response{Message: fmt.Sprintf("Success! %s BAN has been sent your way.", floatString)})
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		cooldown := time.Until(time.Now().Add((CLAIM_INTERVAL * time.Hour) - diff))
		json.NewEncoder(w).Encode(Response{Message: fmt.Sprintf("Last claim too soon. Please wait %s", cooldown)})
	}

}
