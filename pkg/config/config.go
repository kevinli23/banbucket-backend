package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	dbUser         string
	dbPswd         string
	dbHost         string
	dbPort         string
	dbName         string
	mongoUser      string
	mongoPswd      string
	apiPort        string
	faucetSeed     string
	faucetAddr     string
	faucetRep      string
	hcaptchaSecret string
}

func Get() *Config {
	conf := &Config{}

	flag.StringVar(&conf.dbUser, "dbuser", os.Getenv("POSTGRES_USER"), "DB user name")
	flag.StringVar(&conf.dbPswd, "dbpswd", os.Getenv("POSTGRES_PASSWORD"), "DB passord")
	flag.StringVar(&conf.dbPort, "dbport", os.Getenv("POSTGRES_PORT"), "DB port")
	flag.StringVar(&conf.dbHost, "dbhost", os.Getenv("POSTGRES_HOST"), "DB host")
	flag.StringVar(&conf.dbName, "dbname", os.Getenv("POSTGRES_DB"), "DB name")
	flag.StringVar(&conf.mongoUser, "mongouser", os.Getenv("MONGO_USER"), "Mongo username")
	flag.StringVar(&conf.mongoPswd, "mongopswd", os.Getenv("MONGO_PASSWORD"), "Mongo password")
	flag.StringVar(&conf.apiPort, "apiPort", os.Getenv("PORT"), "API Port")
	flag.StringVar(&conf.faucetAddr, "faucetaddr", os.Getenv("FAUCET_ADDR"), "Faucet Address")
	flag.StringVar(&conf.faucetRep, "faucetrep", os.Getenv("FAUCET_REP"), "Faucet Representative")
	flag.StringVar(&conf.faucetSeed, "faucetseed", os.Getenv("FAUCET_SEED"), "Faucet Seed")
	flag.StringVar(&conf.hcaptchaSecret, "hcaptchasecret", os.Getenv("HCAPTCHA_SECRET"), "Hcaptcha Secret")
	flag.Parse()

	return conf
}

func (c *Config) GetSeed() string {
	return c.faucetSeed
}

func (c *Config) GetHCapthcaSecret() string {
	return c.hcaptchaSecret
}

func (c *Config) GetDBConnStr() string {
	return fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", c.dbHost, c.dbPort, c.dbUser, c.dbName, c.dbPswd)
}

func (c *Config) GetAPIPort() string {
	return ":" + c.apiPort
}

func (c *Config) GetFaucetAddress() string {
	return c.faucetAddr
}

func (c *Config) GetFaucetRepAddress() string {
	return c.faucetRep
}

func (c *Config) GetMongoURI() string {
	return fmt.Sprintf("mongodb+srv://%s:%s@cluster0.8qiqc.mongodb.net/faucet?retryWrites=true&w=majority", c.mongoUser, c.mongoPswd)
}
