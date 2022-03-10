package config

import (
	"log"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/shopspring/decimal"
)

type Configuration struct {
	Port                 string       `toml:"port"`
	Database             database     `toml:"database"`
	Dev                  bool         `toml:"dev"`
	SwanApi              swanApi      `toml:"swan_api"`
	IpfsServer           ipfsServer   `toml:"ipfs_server"`
	MaxMultipartMemory   int64        `toml:"max_multipart_memory"`
	Lotus                lotus        `toml:"lotus"`
	SwanTask             swanTask     `toml:"swan_task"`
	ScheduleRule         ScheduleRule `toml:"schedule_rule"`
	AdminWalletOnPolygon string       `toml:"admin_wallet_on_polygon"`
	FileCoinWallet       string       `toml:"file_coin_wallet"`
	FilinkUrl            string       `toml:"filink_url"`
	FilecoinNetwork      string       `toml:"filecoin_network"`
}

type database struct {
	DbUsername   string `toml:"db_username"`
	DbPwd        string `toml:"db_pwd"`
	DbHost       string `toml:"db_host"`
	DbPort       string `toml:"db_port"`
	DbSchemaName string `toml:"db_schema_name"`
	DbArgs       string `toml:"db_args"`
}

type lotus struct {
	ApiUrl              string `toml:"api_url"`
	AccessToken         string `toml:"access_token"`
	FullNodeUrl         string `toml:"full_node_url"`
	FullNodeAccessToken string `toml:"full_node_access_token"`
	FinalStatusList     string `toml:"final_status_list"`
}

type swanTask struct {
	DirDeal                      string          `toml:"dir_deal"`
	Description                  string          `toml:"description"`
	CuratedDataset               string          `toml:"curated_dataset"`
	Tags                         string          `toml:"tags"`
	MinPrice                     decimal.Decimal `toml:"min_price"`
	MaxPrice                     decimal.Decimal `toml:"max_price"`
	ExpireDays                   int             `toml:"expire_days"`
	VerifiedDeal                 bool            `toml:"verified_deal"`
	FastRetrieval                bool            `toml:"fast_retrieval"`
	StartEpochHours              int             `toml:"start_epoch_hours"`
	MinerId                      string          `toml:"miner_id"`
	RelativeEpochFromMainNetwork int             `toml:"relative_epoch_from_main_network"`
}

type swanApi struct {
	ApiUrl                     string `toml:"api_url"`
	ApiKey                     string `toml:"api_key"`
	AccessToken                string `toml:"access_token"`
	GetShouldSendTaskUrlSuffix string `toml:"get_should_send_task_url_suffix"`
}

type ipfsServer struct {
	DownloadUrlPrefix string `toml:"download_url_prefix"`
	UploadUrl         string `toml:"upload_url"`
}

type ScheduleRule struct {
	UnlockPaymentRule   string `toml:"unlock_payment_rule"`
	SendDealRule        string `toml:"send_deal_rule"`
	CreateTaskRule      string `toml:"create_task_rule"`
	ScanDealStatusRule  string `toml:"scan_deal_status_rule"`
	UpdatePayStatusRule string `toml:"update_pay_status_rule"`
}

var config *Configuration

func InitConfig(configFile string) {
	if strings.Trim(configFile, " ") == "" {
		configFile = "./config/config.toml"
	}
	if metaData, err := toml.DecodeFile(configFile, &config); err != nil {
		log.Fatal("error:", err)
	} else {
		if !requiredFieldsAreGiven(metaData) {
			log.Fatal("required fields not given")
		}
	}
}

func GetConfig() Configuration {
	if config == nil {
		InitConfig("")
	}
	return *config
}

func requiredFieldsAreGiven(metaData toml.MetaData) bool {
	requiredFields := [][]string{
		{"port"},
		{"admin_wallet_on_polygon"},
		{"file_coin_wallet"},
		{"filink_url"},
		{"filecoin_network"},

		{"database", "db_host"},
		{"database", "db_port"},
		{"database", "db_username"},
		{"database", "db_schema_name"},
		{"database", "db_pwd"},

		{"swan_api", "api_url"},
		{"swan_api", "api_key"},
		{"swan_api", "get_should_send_task_url_suffix"},

		{"ipfs_server", "download_url_prefix"},
		{"ipfs_server", "upload_url"},

		{"lotus", "api_url"},
		{"lotus", "access_token"},
		{"lotus", "final_status_list"},

		{"swan_task", "relative_epoch_from_main_network"},
	}

	for _, v := range requiredFields {
		if !metaData.IsDefined(v...) {
			log.Fatal("required fields ", v)
		}
	}

	return true
}
