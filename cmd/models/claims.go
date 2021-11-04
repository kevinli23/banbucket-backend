package models

import _ "go.mongodb.org/mongo-driver/bson"

type Claim struct {
	BananoAddr string `json:"addr" bson:"addr"`
	IP         string `json:"ip" bson:"ip"`
	LastClaim  int64  `json:"lastclaim" bson:"last_claim"`
	Claims     int    `json:"claims" bson:"claims"`
}

type Donator struct {
	BananoAddr string  `json:"addr"`
	Amount     float64 `json:"amount"`
}

type AllDonators struct {
	Donators []Donator `json:"donators"`
}

type ClaimRequest struct {
	BananoAddr string `json:"addr"`
	IP         string `json:"ip"`
	Captcha    string `json:"captcha"`
}
