package models

type GetIPIntelResponse struct {
	Status  string `json:"status"`
	QueryIP string `json:"queryIP"`
	Result  string `json:"result"`
}
