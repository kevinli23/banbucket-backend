package stats

import (
	"banfaucetservice/cmd/models"
	"banfaucetservice/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/BananoCoin/gobanano/nano"
)

func GetYellowGlassTransactions(offset uint32) ([]models.YellowGlassTransactionsResponse, error) {
	url := fmt.Sprintf("https://api.yellowspyglass.com/yellowspyglass/confirmed-transactions?address=ban_1j3rqseffoin7x5z5y1ehaqe1n7todza41kdf4oyga8phps3ea31u39ruchu&offset=%d&size=50", offset)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res []models.YellowGlassTransactionsResponse

	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return res, nil
}

func GetNewTransactions(ctx context.Context, f *models.FirestoreHandler) error {
	offset, totalclaims, err := f.GetMetadata()
	if err != nil {
		return err
	}

	var newTransactions []models.YellowGlassTransactionsResponse

	var off uint32
	var count uint32

	for {
		logger.Info.Printf("Receiving transactions with offset %d", off)
		transactions, err := GetYellowGlassTransactions(off)
		if err != nil {
			return err
		}

		if len(transactions) == 0 {
			break
		}

		newTransactions = append(newTransactions, transactions...)
		if newTransactions[len(newTransactions)-1].Height <= offset {
			break
		}

		off += 50
	}

	batch := f.Client.Batch()
	counts := map[string]uint32{}

	logger.Info.Println("Storing information in firebase...")
	for _, transaction := range newTransactions {
		if transaction.Type != "send" || transaction.Height <= offset {
			continue
		}

		count += 1
		amount, err := nano.ParseBalance(transaction.BalanceRaw, "raw")
		if err != nil {
			return err
		}

		amountFloat, err := strconv.ParseFloat(amount.String(), 64)
		if err != nil {
			return err
		}

		year, month, day := time.Unix(int64(transaction.Timestamp), 0).Date()

		date := fmt.Sprintf("%d-%02d-%02d", year, int(month), day)

		ref := f.Client.Collection("transactions").Doc(date).Collection("tx").Doc(transaction.Hash)

		batch.Create(ref, models.FirebaseTransaction{
			Hash:      transaction.Hash,
			Address:   transaction.Address,
			Height:    transaction.Height,
			Timestamp: transaction.Timestamp,
			Amount:    amountFloat * 10,
		})

		if count%500 == 0 {
			batch.Commit(ctx)
			// Ignore error (probably duplicate error) for now
			batch = f.Client.Batch()
		}

		if _, found := counts[date]; !found {
			var c uint32
			doc, err := f.Client.Collection("transactions").Doc(date).Get(ctx)
			if err != nil {
				c = 0
			} else {
				jsonBytes, _ := json.Marshal(doc.Data())

				var res map[string]uint32

				if err := json.Unmarshal(jsonBytes, &res); err != nil {
					return err
				}

				c = res["count"]
			}

			counts[date] = c
		}

		counts[date] += 1
	}

	batch.Commit(ctx)
	batch = f.Client.Batch()

	for date, value := range counts {
		batch.Set(f.Client.Collection("transactions").Doc(date), map[string]interface{}{
			"count": value,
		})
	}

	batch.Commit(ctx)

	if newTransactions[0].Height > offset {
		offset = newTransactions[0].Height
	}

	logger.Info.Printf("Updating metadata: offset=%d totalclaims=%d\n", offset, totalclaims+count)
	f.SetMetadata(offset, totalclaims+count)

	f.CachedStats.TotalClaims = totalclaims + count

	return nil
}
