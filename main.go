package main

import (
	"banfaucetservice/cmd/router"
	"banfaucetservice/pkg/app"
	"banfaucetservice/pkg/banano"
	"banfaucetservice/pkg/exithandler"
	"banfaucetservice/pkg/logger"
	"banfaucetservice/pkg/server"
	"banfaucetservice/pkg/stats"
	"context"
	"log"
	"time"

	"github.com/BananoCoin/gobanano/nano"
	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

func main() {
	if err := godotenv.Load(); err != nil {
		logger.Info.Println("failed to load env vars")
	}

	app, err := app.Get()
	if err != nil {
		log.Fatal(err.Error())
	}

	balance, _, _, err := banano.GetAccountInfo(app.FaucetAddress.String())
	if err != nil {
		log.Fatal(err.Error())
	}

	initBalance, err := nano.ParseBalance(balance, "raw")
	if err != nil {
		log.Fatal(err.Error())
	}

	price, change, err := banano.GetCoinGeckoPrice()
	if err != nil {
		logger.WarningLogger.Printf("Failed to get coingecko price %v", err)
	}

	app.Price = price
	app.PriceChange = change
	app.Amount = initBalance

	stats.GetNewTransactions(context.Background(), app.FBHandler)

	srv := server.
		Get().
		WithAddr(app.Config.GetAPIPort()).
		WithRouter(router.Get(app))

	go func() {
		logger.Info.Printf("connected to postgres db at %s", app.Config.GetDBConnStr())
		logger.Info.Printf("connected to mongo db at %s", app.Config.GetMongoURI())
		logger.Info.Printf("starting server at %s", app.Config.GetAPIPort())
		if err := srv.Start(); err != nil {
			log.Fatal(err.Error())
		}
	}()

	go func() {
		s := gocron.NewScheduler(time.UTC)
		s.Every(1).Hour().Do(func() {
			err := banano.ReceiveBanano(app.Config.GetFaucetAddress(), app)
			if err != nil {
				logger.Info.Println("Nothing was received")
			}
		})

		s.Every(120).Minutes().Do(func() {
			err := stats.GetNewTransactions(context.Background(), app.FBHandler)
			if err != nil {
				logger.Error.Println(errors.Wrap(err, "Failed to retrieve new transactions"))
			}
			err = app.FBHandler.GenerateStats(context.Background(), app.MongoClient.Database("faucet").Collection("claim"))
			if err != nil {
				logger.Error.Println(errors.Wrap(err, "Failed to generate stats"))
			}
		})

		s.Every(5).Minute().Do(func() {
			price, change, err := banano.GetCoinGeckoPrice()
			if err != nil {
				logger.WarningLogger.Printf("Failed to get coingecko price %v", err)
			}

			app.Price = price
			app.PriceChange = change
		})

		s.StartAsync()
	}()

	exithandler.Init(func() {
		if err := srv.Close(); err != nil {
			log.Fatal(err.Error())
		}

		if err := app.DB.Close(); err != nil {
			log.Fatal(err.Error())
		}

		if err := app.MongoClient.Disconnect(context.TODO()); err != nil {
			log.Fatal(err.Error())
		}

		if err := app.FBHandler.Client.Close(); err != nil {
			log.Fatal(err.Error())
		}
	})
}
