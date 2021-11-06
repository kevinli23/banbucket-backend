package banano

import (
	"banfaucetservice/pkg/logger"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"

	banano "github.com/BananoCoin/gobanano/nano"
	"github.com/BananoCoin/gobanano/nano/block"
	"github.com/BananoCoin/gobanano/nano/wallet"
)

const AMOUNT_TO_SEND = "3000000000000000000000000000"
const REDUCED_AMOUNT = "1000000000000000000000000000"
const SUPER_REDUCED = "100000000000000000000000000"
const PUBLIC_KEY = "4438BE58D6D6142F47F1F80C7A2EC050BAAAFE81024B68ABE720D67DB2162020"
const API_URL = "https://api-beta.banano.cc"

var specialRepresentatives = []string{
	"ban_19potasho7ozny8r1drz3u3hb3r97fw4ndm4hegdsdzzns1c3nobdastcgaa",
}

var badRepresentatives = []string{
	"ban_1ka1ium4pfue3uxtntqsrib8mumxgazsjf58gidh1xeo5te3whsq8z476goo",
	"ban_1bananobh5rat99qfgt1ptpieie5swmoth87thi74qgbfrij7dcgjiij94xr",
	"ban_1on1ybanskzzsqize1477wximtkdzrftmxqtajtwh4p4tg1w6awn1hq677cp",
}

func GetNewBalanceAndFrontier(addr string, dest string, destRepresentative string, unopened bool) (banano.Balance, block.Hash, banano.Balance, error) {
	balance, frontier, _, err := GetAccountInfo(addr)
	if err != nil {
		return banano.Balance{}, block.Hash{}, banano.Balance{}, err
	}

	isSpecialRepresentative := false
	for _, rep := range specialRepresentatives {
		if destRepresentative == rep {
			isSpecialRepresentative = true
			break
		}
	}

	// 28 characters is 0.01 - 0.09
	if len(balance) <= 28 {
		return banano.Balance{}, block.Hash{}, banano.Balance{}, fmt.Errorf("BanBucket is dry - please consider donating")
	}

	newBalance, err := banano.ParseBalance(balance, "raw")
	if err != nil {
		return banano.Balance{}, block.Hash{}, banano.Balance{}, err
	}

	var amount banano.Balance
	amount, err = banano.ParseBalance(AMOUNT_TO_SEND, "raw")
	if err != nil {
		return banano.Balance{}, block.Hash{}, banano.Balance{}, err
	}

	for _, rep := range badRepresentatives {
		if rep == destRepresentative {
			amount, err = banano.ParseBalance(REDUCED_AMOUNT, "raw")
			if err != nil {
				return banano.Balance{}, block.Hash{}, banano.Balance{}, err
			}
		}
	}

	if isSpecialRepresentative {
		roll := rand.Intn(10)
		logger.Info.Printf("%s has a special representative and rolled %d\n", dest, roll)

		if roll <= 2 {
			amount = amount.Add(amount)
		}
	}

	if unopened {
		amount, err = banano.ParseBalance(SUPER_REDUCED, "raw")
		if err != nil {
			return banano.Balance{}, block.Hash{}, banano.Balance{}, err
		}
	}

	newBalance = newBalance.Sub(amount)

	frontierHash := ParseFrontier(frontier)

	return newBalance, frontierHash, amount, nil
}

func ParseFrontier(frontier string) block.Hash {
	hash := new(block.Hash)

	frontierHash, _ := hex.DecodeString(frontier)

	for i, v := range frontierHash {
		hash[i] = v
	}

	return *hash
}

func GetAccountPK(seed string) (*wallet.Account, error) {
	seedBytes, err := hex.DecodeString(seed)

	if err != nil {
		return nil, err
	}

	ms := new(wallet.Seed)

	for i, v := range seedBytes {
		ms[i] = v
	}

	pkb, _ := ms.Key(0)

	acc := wallet.NewAccount(pkb)

	return acc, nil
}

func BlockHashStringToHash(hash string) (*block.Hash, error) {
	decodedHash, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	blockHash := new(block.Hash)

	for i, v := range decodedHash {
		blockHash[i] = v
	}

	return blockHash, nil
}

func GetDestinationHash(addr string) (*block.Hash, error) {
	receieve, err := banano.ParseAddress(addr)
	if err != nil {
		return nil, err
	}

	destPK, err := GetAccountKey(receieve.String())
	if err != nil {
		return nil, err
	}

	decoded_pk, err := hex.DecodeString(destPK)
	if err != nil {
		return nil, err
	}

	dest_pk_hash := new(block.Hash)

	for i, v := range decoded_pk {
		dest_pk_hash[i] = v
	}

	return dest_pk_hash, nil
}

func ReceiveNewAmount(hash string, currentAmount banano.Balance) (banano.Balance, string, string, error) {
	toReceive, donator, err := GetBlockInfoAmount(hash)
	if err != nil {
		return banano.Balance{}, "", "", err
	}

	amountReceived, err := banano.ParseBalance(toReceive, "raw")
	if err != nil {
		return banano.Balance{}, "", "", err
	}

	return amountReceived.Add(currentAmount), amountReceived.String(), donator, nil
}

func GetCoinGeckoPrice() (float64, float64, error) {
	coingeckoURL := "https://api.coingecko.com/api/v3/simple/price?ids=banano&vs_currencies=usd&include_24hr_change=true"
	response, err := http.Get(coingeckoURL)
	if err != nil {
		return 0, 0, err
	}

	defer response.Body.Close()

	responseBodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, 0, err
	}

	var res GetCoinGeckoPriceResponse
	if err := json.Unmarshal(responseBodyBytes, &res); err != nil {
		return 0, 0, err
	}

	return res.Banano.USD, res.Banano.Change, nil
}

// Return true for now as this is a test measure/check
func GetYellowSpyGlassAccountOpened(addr string) bool {
	yellowSpyGlassURL := fmt.Sprintf("https://api.yellowspyglass.com/yellowspyglass/account-overview/%s", addr)
	response, err := http.Get(yellowSpyGlassURL)
	if err != nil {
		return true
	}

	defer response.Body.Close()

	responseBodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return true
	}

	var res YellowSpyGlassAccountOverviewResponse
	if err := json.Unmarshal(responseBodyBytes, &res); err != nil {
		return true
	}

	return res.Opened
}
