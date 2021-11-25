package config

import (
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	// -- Admin (or the PDS) account --

	AdminAddress           string `env:"FLOW_PDS_ADMIN_ADDRESS,notEmpty"`
	AdminPrivateKey        string `env:"FLOW_PDS_ADMIN_PRIVATE_KEY,notEmpty"`
	AdminPrivateKeyIndexes []int  `env:"FLOW_PDS_ADMIN_PRIVATE_KEY_INDEXES,notEmpty" envDefault:"0" envSeparator:","`
	AdminPrivateKeyType    string `env:"FLOW_PDS_ADMIN_PRIVATE_KEY_TYPE,notEmpty" envDefault:"local"`

	// -- Flow addresses --
	// Address of the PDS account, usually this should equal to 'AdminAddress'
	PDSAddress              string `env:"PDS_ADDRESS,notEmpty"`
	NonFungibleTokenAddress string `env:"NON_FUNGIBLE_TOKEN_ADDRESS,notEmpty"`

	// -- Database --

	DatabaseDSN  string `env:"FLOW_PDS_DATABASE_DSN" envDefault:"pds.db"`
	DatabaseType string `env:"FLOW_PDS_DATABASE_TYPE" envDefault:"sqlite"`

	// -- Host and chain access --

	Host          string `env:"FLOW_PDS_HOST"`
	Port          int    `env:"FLOW_PDS_PORT" envDefault:"3000"`
	AccessAPIHost string `env:"FLOW_PDS_ACCESS_API_HOST" envDefault:"localhost:3569"`

	// -- Rates etc. ---

	// How many transactions to send per second at max
	TransactionSendRate int    `env:"FLOW_PDS_SEND_RATE" envDefault:"10"`
	TransactionGasLimit uint64 `env:"FLOW_PDS_GAS_LIMIT" envDefault:"9999"`
	// Going much above 40 will cause the transactions to use more than 9999 gas
	SettlementBatchSize int `env:"FLOW_PDS_SETTLEMENT_BATCH_SIZE" envDefault:"40"`
	MintingBatchSize    int `env:"FLOW_PDS_MINTING_BATCH_SIZE" envDefault:"40"`

	// The batchSize for database batch handling (big inserts or batch processing)
	QueryBatchSize int `env:"FLOW_PDS_QUERY_BATCH_SIZE" envDefault:"1000"`

	// -- Testing --

	TestPackCount int `env:"TEST_PACK_COUNT" envDefault:"4"`
}

type ConfigOptions struct {
	EnvFilePath string
}

// ParseConfig parses environment variables and flags to a valid Config.
func ParseConfig(opt *ConfigOptions) (*Config, error) {
	if opt != nil && opt.EnvFilePath != "" {
		// Load variables from a file to the environment of the process
		if err := godotenv.Load(opt.EnvFilePath); err != nil {
			log.Printf("Could not load environment variables from file.\n%s\nIf running inside a docker container this can be ignored.\n\n", err)
		}
	}

	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
