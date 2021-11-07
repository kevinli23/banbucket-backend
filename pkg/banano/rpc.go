package banano

import (
	"banfaucetservice/pkg/logger"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/BananoCoin/gobanano/nano/block"
)

func GetAccountKey(addr string) (string, error) {
	requestBody, _ := json.Marshal(AccountKeyRequest{
		Action:  "account_key",
		Account: addr,
	})

	response, err := http.Post(API_URL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	responseBodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	var res AccountKeyResponse
	if err := json.Unmarshal(responseBodyBytes, &res); err != nil {
		return "", err
	}

	return res.Key, nil
}

func GetAccountRepresentative(addr string) (string, error) {
	requestBody, _ := json.Marshal(AccountKeyRequest{
		Action:  "account_representative",
		Account: addr,
	})

	response, err := http.Post(API_URL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	responseBodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	var res AccountRepresentativeResponse
	if err := json.Unmarshal(responseBodyBytes, &res); err != nil {
		return "", err
	}

	return res.Representative, nil
}

func BananoGenerateWork(hash string) (uint64, error) {
	logger.Info.Printf("Started work generation for %s\n", strings.ToUpper(hash))
	requestBody, _ := json.Marshal(map[string]string{
		"action": "work_generate",
		"hash":   strings.ToUpper(hash),
	})

	response, err := http.Post(API_URL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return 0, err
	}

	defer response.Body.Close()

	responseBodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}

	defer response.Body.Close()

	var res WorkResponse
	if err := json.Unmarshal(responseBodyBytes, &res); err != nil {
		return 0, err
	}

	work, err := strconv.ParseUint(res.Work, 16, 64)
	if err != nil {
		return 0, err
	}

	logger.Info.Printf("Finished work generation for %s\n", hash)
	return work, nil
}

func BananoFaucetReceive(block block.StateBlock) (string, error) {
	return BananoFaucetProcess(block, "receive")
}

func BananoFaucetSend(block block.StateBlock) (string, error) {
	return BananoFaucetProcess(block, "send")
}

func BananoFaucetProcess(sBlock block.StateBlock, subtype string) (string, error) {

	requestBody, _ := json.Marshal(SendRequest{
		Action:    "process",
		JsonBlock: "true",
		Subtype:   subtype,
		Block: SendRequestBlock{
			Type:           "state",
			Account:        sBlock.Address.String(),
			Previous:       sBlock.PreviousHash.String(),
			Representative: sBlock.Representative.String(),
			Balance:        sBlock.Balance.BigInt().String(),
			Link:           sBlock.Link.String(),
			Work:           sBlock.Work.String(),
			Signature:      sBlock.Signature.String(),
		},
		DoWork: "true",
	})

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	response, err := client.Post(API_URL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	responseBodyBytes, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	var res SendBlockResponse

	if err := json.Unmarshal(responseBodyBytes, &res); err != nil {
		return "", err
	}

	if res.Error != "" {
		return "", fmt.Errorf(res.Error)
	}

	return res.Hash, nil
}

func GetAccountInfo(addr string) (string, string, string, error) {
	requestBody, _ := json.Marshal(map[string]string{
		"action":         "account_info",
		"account":        addr,
		"representative": "true",
	})

	response, err := http.Post(API_URL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", "", "", err
	}

	defer response.Body.Close()

	responseBodyBytes, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return "", "", "", err
	}

	defer response.Body.Close()

	var res AccountInfo

	if err := json.Unmarshal(responseBodyBytes, &res); err != nil {
		return "", "", "", err
	}

	return res.Balance, res.Frontier, res.Representative, nil
}

func GetAccountPendings(addr string) ([]string, error) {
	requestBody, _ := json.Marshal(AccountsPendingRequest{
		Action:   "accounts_pending",
		Accounts: []string{addr},
	})

	response, err := http.Post(API_URL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	responseBodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	var res AccountsPendingResponse

	if err := json.Unmarshal(responseBodyBytes, &res); err != nil {
		return nil, err
	}

	return res.Blocks.Hashes, nil
}

func GetBlockInfoAmount(blockHash string) (string, string, error) {
	requestBody, _ := json.Marshal(BlockInfoRequest{
		Action:    "block_info",
		JsonBlock: "true",
		Hash:      blockHash,
	})

	response, err := http.Post(API_URL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", "", err
	}

	defer response.Body.Close()

	responseBodyBytes, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return "", "", err
	}

	defer response.Body.Close()

	var res BlockInfoResponse
	if err := json.Unmarshal(responseBodyBytes, &res); err != nil {
		return "", "", err
	}

	return res.Amount, res.BlockAccount, nil
}
