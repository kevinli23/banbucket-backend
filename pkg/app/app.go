package app

import (
	"banfaucetservice/cmd/models"
	"banfaucetservice/pkg/config"
	"banfaucetservice/pkg/db"
	"banfaucetservice/pkg/logger"
	"database/sql"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/BananoCoin/gobanano/nano"
	"github.com/newrelic/go-agent/v3/newrelic"
	"go.mongodb.org/mongo-driver/mongo"
)

type App struct {
	DB            *sql.DB
	MongoClient   *mongo.Client
	Config        *config.Config
	BannedIPs     []byte
	BannedAddrs   []byte
	Amount        nano.Balance
	FaucetAddress nano.Address
	FaucetRep     nano.Address
	Donators      models.AllDonators
	Lock          sync.Mutex
	Price         float64
	PriceChange   float64
	NewRelicApp   *newrelic.Application
	FBHandler     *models.FirestoreHandler
}

func Get() (*App, error) {
	cfg := config.Get()

	sqlDB, err := db.GetPostgresDB(cfg.GetDBConnStr())
	if err != nil {
		return nil, err
	}

	client, err := db.GetMongoDB(cfg.GetMongoURI())
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadFile("bannedips.txt")
	if err != nil {
		return nil, err
	}

	bannedAddrs, err := ioutil.ReadFile("bannedaddr.txt")
	if err != nil {
		return nil, err
	}

	address, err := nano.ParseAddress(cfg.GetFaucetAddress())
	if err != nil {
		return nil, err
	}

	rep, err := nano.ParseAddress(cfg.GetFaucetRepAddress())
	if err != nil {
		return nil, err
	}

	nrApp, err := newrelic.NewApplication(
		newrelic.ConfigAppName("banbucket"),
		newrelic.ConfigLicense(os.Getenv("NEW_RELIC_LICENSE")),
		newrelic.ConfigDistributedTracerEnabled(true),
	)

	firebaseHandler := &models.FirestoreHandler{}
	firebaseHandler.New()

	if err != nil {
		return nil, err
	}

	return &App{
		DB:            sqlDB,
		MongoClient:   client,
		Config:        cfg,
		BannedIPs:     b,
		BannedAddrs:   bannedAddrs,
		FaucetAddress: address,
		FaucetRep:     rep,
		Lock:          sync.Mutex{},
		NewRelicApp:   nrApp,
		FBHandler:     firebaseHandler,
	}, nil
}

func (app *App) IsBanned(addr string) (bool, error) {
	isBanned, err := regexp.Match(addr, app.BannedAddrs)
	if err != nil {
		logger.Error.Println("Failed to compute valid IP")
		return false, err
	}
	return isBanned, nil
}

func (app *App) IsVPN(ip string) (bool, error) {

	var network strings.Builder

	nums := strings.Split(ip, ".")

	for i, num := range nums {
		if i == 3 {
			continue
		}
		network.WriteString(num)
		network.WriteString(".")
	}

	network.WriteString("0/24")

	networkVPN, err := regexp.Match(network.String(), app.BannedIPs)
	if err != nil {
		logger.Error.Println("Failed to compute network IP")
		return false, err
	}

	ipVPN, err := regexp.Match(ip, app.BannedIPs)
	if err != nil {
		logger.Error.Println("Failed to compute valid IP")
		return false, err
	}
	return ipVPN || networkVPN, nil
}

func (app *App) ProcessBananoReceive(addr string, amount float64) error {
	queryStatement := `SELECT * FROM donation WHERE addr = $1`

	rows := app.DB.QueryRow(queryStatement, addr)

	var donator models.Donator

	err := rows.Scan(&donator.BananoAddr, &donator.Amount)

	if err == sql.ErrNoRows {
		err2 := app.InsertDonator(addr, amount)
		if err2 != nil {
			return err2
		}

		donators, err := app.GetAllDonators()
		if err != nil {
			return err
		}

		app.Donators.Donators = donators
		return nil
	}

	if err != nil {
		return err
	}

	if err := app.UpdateDonator(addr, amount); err != nil {
		return err
	}

	// Retrieve and cache donators
	donators, err := app.GetAllDonators()
	if err != nil {
		return err
	}

	app.Donators.Donators = donators
	return nil
}

func (app *App) UpdateDonator(addr string, amount float64) error {
	logger.Info.Printf("Updating donation %s into db with amount %.2f\n", addr, amount)
	stmt := `UPDATE donation SET amount=amount + $1 WHERE addr = $2`

	_, err := app.DB.Exec(stmt, amount, addr)

	return err
}

func (app *App) InsertDonator(addr string, amount float64) error {
	logger.Info.Printf("Inserting donator %s into db with amount %.2f\n", addr, amount)
	stmt := `INSERT INTO donation(addr, amount) VALUES ($1, $2)`

	_, err := app.DB.Exec(stmt, addr, amount)

	return err
}

func (app *App) GetAllDonators() ([]models.Donator, error) {
	stmt := `SELECT * FROM donation`

	rows, err := app.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	donators := []models.Donator{}

	for rows.Next() {
		var dono models.Donator
		if err := rows.Scan(&dono.BananoAddr, &dono.Amount); err != nil {
			return nil, err
		}
		donators = append(donators, dono)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return donators, nil
}
