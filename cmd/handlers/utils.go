package handlers

import (
	"banfaucetservice/cmd/models"
	"banfaucetservice/pkg/app"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

const HCAPTCHA_VERIFY_URL = "https://hcaptcha.com/siteverify"

type HCaptchaResponse struct {
	Success     bool   `json:"success"`
	ChallengeTS string `json:"challenge_ts"`
	HostName    string `json:"hostname"`
	ErrorCodes  string `json:"error_codes"`
	Credit      bool   `json:"credit"`
}

type Response struct {
	Message string `json:"message"`
}

type PriceResponse struct {
	Price  string `json:"price"`
	Change string `json:"change"`
}

func verifyHCaptcha(responseToken string, secret string) (bool, error) {
	if !VERIFY_HCAPTCHA_FLAG {
		return true, nil
	}

	data := url.Values{}

	data.Set("response", responseToken)
	data.Set("secret", secret)

	client := &http.Client{}
	req, err := http.NewRequest("POST", HCAPTCHA_VERIFY_URL, strings.NewReader(data.Encode()))
	if err != nil {
		return false, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	response, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()

	responseBodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, err
	}

	var res HCaptchaResponse
	if err := json.Unmarshal(responseBodyBytes, &res); err != nil {
		return false, err
	}

	return res.Success, nil
}

func updateClaims(app *app.App, claimReq models.ClaimRequest) error {
	filter := bson.M{
		"$or": []bson.M{
			{"addr": claimReq.BananoAddr},
			{"ip": claimReq.IP},
		},
	}

	update := bson.M{
		"$set": bson.M{"last_claim": time.Now().Unix()},
		"$inc": bson.M{"claims": 1},
	}

	_, err := app.MongoClient.Database("faucet").Collection("claim").UpdateMany(context.TODO(), filter, update)

	return err

	// stmt := `UPDATE claim SET last_claim=$1, claims=claims + 1 WHERE addr=$2 OR ip=$3`

	// _, err := app.DB.Exec(stmt, claimReq.BananoAddr, claimReq.IP, time.Now().Unix())

	// return err
}

func insertClaim(app *app.App, claimReq models.ClaimRequest) error {
	claim := models.Claim{
		BananoAddr: claimReq.BananoAddr,
		IP:         claimReq.IP,
		LastClaim:  time.Now().Unix(),
		Claims:     1,
	}

	_, err := app.MongoClient.Database("faucet").Collection("claim").InsertOne(context.TODO(), claim)

	return err

	// stmt := `INSERT INTO claim(addr, ip, last_claim, claims) VALUES ($1, $2, $3, $4)`

	// _, err := app.DB.Exec(stmt, claimReq.BananoAddr, claimReq.IP, time.Now().Unix(), 1)

	// return err
}

func getLastClaim(app *app.App, addr string, ip string, ctx context.Context) ([]models.Claim, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"addr": addr},
			{"ip": ip},
		},
	}

	cursor, err := app.MongoClient.Database("faucet").Collection("claim").Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var lastClaims []models.Claim

	for cursor.Next(ctx) {
		var c models.Claim
		err := cursor.Decode(&c)
		if err != nil {
			return nil, err
		}

		lastClaims = append(lastClaims, c)
	}

	return lastClaims, nil

	// stmt := `SELECT * FROM claim WHERE addr=$1 OR ip=$2`

	// var lastClaims []models.Claim

	// rows, err := app.DB.Query(stmt, addr, ip)
	// if err != nil {
	// 	return nil, err
	// }
	// defer rows.Close()

	// for rows.Next() {
	// 	var lc models.Claim
	// 	if err := rows.Scan(&lc.BananoAddr, &lc.IP, &lc.LastClaim, &lc.Claims); err != nil {
	// 		return nil, err
	// 	}
	// 	lastClaims = append(lastClaims, lc)
	// }
	// if err = rows.Err(); err != nil {
	// 	return nil, err
	// }
	// return lastClaims, nil
}

func reverseOctets(ip string) string {
	oct := strings.Split(ip, ".")
	for i := len(oct)/2 - 1; i >= 0; i-- {
		opp := len(oct) - 1 - i
		oct[i], oct[opp] = oct[opp], oct[i]
	}
	return strings.Join(oct, ".")
}

func isTorIP(clientIP string) bool {
	torExt := fmt.Sprintf(torDNSEL, reverseOctets(clientIP))

	addrs, _ := net.LookupHost(torExt)
	for _, addr := range addrs {
		if addr == torValidationAddr {
			return true
		}
	}

	return false
}

func IsMaliciousIP(clientIP string) bool {
	requestURL := fmt.Sprintf("https://check.getipintel.net/check.php?ip=%s&contact=aiglette_carbone@aleeas.com&flags=m&format=json", clientIP)

	response, err := http.Get(requestURL)
	if err != nil {
		return false
	}

	defer response.Body.Close()

	responseBodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false
	}

	var res models.GetIPIntelResponse
	if err := json.Unmarshal(responseBodyBytes, &res); err != nil {
		return false
	}

	result, err := strconv.ParseFloat(res.Result, 64)
	if err != nil {
		return false
	}

	return result > 0.99
}

// func getISP(clientIP string) string {

// }
