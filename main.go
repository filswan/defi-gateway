package main

import (
	"fmt"
	"os"
	"payment-bridge/blockchain/browsersync"
	"payment-bridge/common/constants"
	"payment-bridge/config"
	"payment-bridge/database"
	"payment-bridge/logs"
	"payment-bridge/models"
	"payment-bridge/routers"
	"payment-bridge/routers/billing"
	"payment-bridge/routers/common"
	"payment-bridge/routers/storage"
	"payment-bridge/scheduler"
	"time"

	"github.com/gin-gonic/gin"
	cors "github.com/itsjamie/gin-cors"
	"github.com/joho/godotenv"
)

func main() {
	LoadEnv()
	// init database
	db := database.Init()

	initMethod()
	browsersync.Init()

	models.RunAllTheScan()

	scheduler.CreateTaskScheduler()

	scheduler.SendDealScheduler()

	scheduler.DAOUnlockPaymentSchedule()
	//scheduler.RefundUnlockPaymentSchedule()
	scheduler.ScanDealInfoScheduler()
	scheduler.ScanExpiredDealInfoScheduler()

	defer func() {
		err := db.Close()
		if err != nil {
			logs.GetLogger().Error(err)
		}
	}()

	r := gin.Default()
	r.MaxMultipartMemory = config.GetConfig().MaxMultipartMemory << 20
	r.Use(cors.Middleware(cors.Config{
		Origins:         "*",
		Methods:         "GET, PUT, POST, DELETE",
		RequestHeaders:  "Origin, Authorization, Content-Type",
		ExposedHeaders:  "",
		MaxAge:          50 * time.Second,
		Credentials:     true,
		ValidateHeaders: false,
	}))

	v1 := r.Group("/api/v1")
	common.HostManager(v1.Group(constants.URL_HOST_GET_COMMON))
	routers.EventLogManager(v1.Group(constants.URL_EVENT_PREFIX))
	billing.BillingManager(v1.Group(constants.URL_BILLING_PREFIX))
	storage.SendDealManager(v1.Group(constants.URL_STORAGE_PREFIX))

	err := r.Run(":" + config.GetConfig().Port)
	if err != nil {
		logs.GetLogger().Fatal(err)
	}

}

func initMethod() string {
	config.InitConfig("")
	return ""
}

func LoadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		logs.GetLogger().Error(err)
	}
	fmt.Println("name: ", os.Getenv("privateKey"))
}
