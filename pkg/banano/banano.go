package banano

import (
	"banfaucetservice/pkg/app"
	"banfaucetservice/pkg/logger"
	"strconv"

	"github.com/BananoCoin/gobanano/nano"
	"github.com/BananoCoin/gobanano/nano/block"
)

func SendBanano(dest string, app *app.App) (string, nano.Balance, error) {
	address := app.FaucetAddress
	rep := app.FaucetRep

	account, err := GetAccountPK(app.Config.GetSeed())
	if err != nil {
		return "", nano.Balance{}, err
	}

	destPubKey, err := GetDestinationHash(dest)
	if err != nil {
		return "", nano.Balance{}, err
	}

	destBalance, _, destRepresentative, err := GetAccountInfo(dest)
	unopened := destBalance == "0" || err != nil || !GetYellowSpyGlassAccountOpened(dest)

	app.Lock.Lock()

	newBalance, frontier, amountGiven, err := GetNewBalanceAndFrontier(address.String(), dest, destRepresentative, unopened)
	if err != nil {
		app.Lock.Unlock()
		return "", nano.Balance{}, err
	}

	// work, err := BananoGenerateWork(frontier.String())
	// if err != nil {
	// 	logger.Error.Printf("Failed to generate work: %x", err)
	// 	app.Lock.Unlock()
	// 	return "", nano.Balance{}, err
	// }

	sendBlock := block.StateBlock{
		Address:        address,
		Representative: rep,
		Balance:        newBalance,
		PreviousHash:   frontier,
		Link:           *destPubKey,
		// Work:           block.Work(work),
	}

	sendBlock.Signature = account.Sign(sendBlock.Hash())

	logger.Info.Printf("Begin sending for %s\n", dest)
	newHash, err := BananoFaucetSend(sendBlock)
	if err != nil {
		logger.Error.Printf("Failed to send: %x", err)
		app.Lock.Unlock()
		return "", nano.Balance{}, err
	}

	app.Lock.Unlock()

	// go func() {
	// 	logger.Info.Printf("Starting computation of work for next hash: %s\n", newHash)
	// 	BananoGenerateWork(newHash)
	// }()

	app.Amount = newBalance

	return newHash, amountGiven, nil
}

func ReceiveBanano(addr string, app *app.App) error {
	logger.Info.Printf("Started receive process for %s\n", addr)
	pendings, err := GetAccountPendings(addr)
	if err != nil {
		return err
	}

	logger.Info.Printf("Found the following blocks %v\n", pendings)

	for _, pending := range pendings {
		logger.Info.Printf("Started receiving block: %s\n", pending)
		receiveHash, err := ReceiveBananoFromSpecificHash(addr, pending, app)
		if err != nil || receiveHash == "" {
			logger.Info.Printf("Failed to receive hash %s: %v\n", pending, err)
			return err
		} else {
			logger.Info.Printf("Successfully received banano: %s\n", receiveHash)
		}
	}

	return nil
}

func ReceiveBananoFromSpecificHash(addr string, hash string, app *app.App) (string, error) {
	address := app.FaucetAddress
	rep := app.FaucetRep

	account, err := GetAccountPK(app.Config.GetSeed())
	if err != nil {
		return "", err
	}

	blochHash, err := BlockHashStringToHash(hash)
	if err != nil {
		return "", err
	}

	app.Lock.Lock()

	balance, frontier, _, err := GetAccountInfo(addr)
	if err != nil {
		return "", err
	}

	parsedFrontier := ParseFrontier(frontier)

	parsedBalance, err := nano.ParseBalance(balance, "raw")
	if err != nil {
		return "", err
	}

	newBalance, amountReceived, donator, err := ReceiveNewAmount(hash, parsedBalance)
	if err != nil {
		return "", err
	}

	work, _ := BananoGenerateWork(parsedFrontier.String())

	receiveBlock := block.StateBlock{
		Address:        address,
		Representative: rep,
		Balance:        newBalance,
		PreviousHash:   parsedFrontier,
		Link:           *blochHash,
		Work:           block.Work(work),
	}

	receiveBlock.Signature = account.Sign(receiveBlock.Hash())

	newHash, err := BananoFaucetReceive(receiveBlock)
	if err != nil {
		return "", err
	}

	amountReceivedFloat, _ := strconv.ParseFloat(amountReceived, 64)
	err = app.ProcessBananoReceive(donator, amountReceivedFloat*10)
	if err != nil {
		logger.Error.Printf("Failed To Cache Donators: %v\n", err)
	}

	app.Lock.Unlock()

	app.Amount = newBalance

	return newHash, nil
}
