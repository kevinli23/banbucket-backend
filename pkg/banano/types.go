package banano

type AccountHistoryResponse struct {
	Account string `json:"account"`
	History []struct {
		Type           string `json:"type"`
		Account        string `json:"account"`
		Amount         string `json:"amount"`
		LocalTimestamp string `json:"local_timestamp"`
		Height         string `json:"height"`
		Hash           string `json:"hash"`
	} `json:"history"`
	Previous string `json:"previous"`
}

type AccountInfo struct {
	Frontier       string `json:"frontier"`
	Balance        string `json:"balance"`
	Representative string `json:"representative"`
}

type AccountRepresentativeResponse struct {
	Representative string `json:"representative"`
}

type Block struct {
	Type           string `json:"type"`
	Account        string `json:"account"`
	Previous       string `json:"previous"`
	Representative string `json:"representative"`
	Balance        string `json:"balance"`
	Work           string `json:"work"`
	Link           string `json:"link"`
	Signature      string `json:"signature"`
}

type WorkResponse struct {
	Work       string `json:"work"`
	Difficulty string `json:"difficulty"`
	Multiplier string `json:"multiplier"`
	Hash       string `json:"hash"`
}

type SendBlockResponse struct {
	Hash  string `json:"hash"`
	Error string `json:"error"`
}

type AccountKeyResponse struct {
	Key string `json:"key"`
}

type SendRequestBlock struct {
	Type           string `json:"type"`
	Account        string `json:"account"`
	Previous       string `json:"previous"`
	Representative string `json:"representative"`
	Balance        string `json:"balance"`
	Link           string `json:"link"`
	Work           string `json:"work"`
	Signature      string `json:"signature"`
}

type SendRequest struct {
	Action    string           `json:"action"`
	JsonBlock string           `json:"json_block"`
	Subtype   string           `json:"subtype"`
	Block     SendRequestBlock `json:"block"`
	DoWork    string           `json:"do_work"`
}

type AccountKeyRequest struct {
	Action  string `json:"action"`
	Account string `json:"account"`
}

type AccountsPendingRequest struct {
	Action   string   `json:"action"`
	Accounts []string `json:"accounts"`
	Count    string   `json:"counts"`
}

type AccountsPendingBlock struct {
	Hashes []string `json:"ban_1j3rqseffoin7x5z5y1ehaqe1n7todza41kdf4oyga8phps3ea31u39ruchu"`
}

type AccountsPendingResponse struct {
	Blocks AccountsPendingBlock `json:"blocks"`
}

type BlockInfoRequest struct {
	Action    string `json:"action"`
	JsonBlock string `json:"json_block"`
	Hash      string `json:"hash"`
}

type BlockInfoResponse struct {
	BlockAccount   string      `json:"block_account"`
	Amount         string      `json:"amount"`
	Confirmed      string      `json:"confirmed"`
	Height         string      `json:"height"`
	LocalTimestamp string      `json:"local_timestamp"`
	Subtype        string      `json:"subtype"`
	Contents       interface{} `json:"contents"`
}

type CoinGeckoPrice struct {
	USD    float64 `json:"usd"`
	Change float64 `json:"usd_24h_change"`
}

type GetCoinGeckoPriceResponse struct {
	Banano CoinGeckoPrice `json:"banano"`
}

type YellowSpyGlassAccountOverviewResponse struct {
	Opened bool `json:"opened"`
}
