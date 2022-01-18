package scheduler

import (
	"fmt"
	"payment-bridge/common/constants"
	"payment-bridge/common/utils"
	"payment-bridge/config"
	"payment-bridge/database"
	"payment-bridge/logs"
	"payment-bridge/models"
	"strconv"
	"strings"
	"time"

	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/robfig/cron"
)

func ScanDealInfoScheduler() {
	c := cron.New()
	err := c.AddFunc(config.GetConfig().ScheduleRule.ScanDealStatusRule, func() {
		logs.GetLogger().Info("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ scan deal info from chain scheduler is running at " + time.Now().Format("2006-01-02 15:04:05"))
		err := GetDealInfoByLotusClientAndUpdateInfoToDB()
		if err != nil {
			logs.GetLogger().Error(err)
			return
		}
	})
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}
	c.Start()
}

func GetDealInfoByLotusClientAndUpdateInfoToDB() error {
	inList := "'" + strings.Join(strings.Split(config.GetConfig().Lotus.FinalStatusList, ","), "', '") + "'"
	fmt.Println(inList)
	whereCondition := "deal_cid != '' and task_uuid != '' and lower(lock_payment_status) not in (lower('" + constants.LOCK_PAYMENT_STATUS_SUCCESS + "'), lower('" + constants.LOCK_PAYMENT_STATUS_REFUNDED + "'), lower('" + constants.LOCK_PAYMENT_STATUS_REFUNDING + "'))"
	//" and deal_status not in (" + inList + ")"
	dealList, err := models.FindDealFileList(whereCondition, "create_at desc", "100", "0")
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	for _, v := range dealList {
		lotusClient, err := lotus.LotusGetClient(config.GetConfig().Lotus.ApiUrl, config.GetConfig().Lotus.AccessToken)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		dealInfo, err := lotusClient.LotusClientGetDealInfo(v.DealCid)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		paymentStatus := ""
		if strings.ToLower(dealInfo.Status) == strings.ToLower(constants.DEAL_STATUS_ACTIVE) {
			paymentStatus = constants.LOCK_PAYMENT_STATUS_SUCCESS
		} else if strings.ToLower(dealInfo.Status) == strings.ToLower(constants.DEAL_STATUS_ERROR) {
			paymentStatus = constants.LOCK_PAYMENT_STATUS_REFUNDING
		} else {
			paymentStatus = constants.LOCK_PAYMENT_STATUS_PROCESSING
		}
		v.Verified = dealInfo.Verified
		v.DealStatus = dealInfo.Status
		v.DealId = dealInfo.DealId
		v.Cost = dealInfo.CostComputed
		v.LockPaymentStatus = paymentStatus
		v.UpdateAt = strconv.FormatInt(utils.GetEpochInMillis(), 10)
		err = database.SaveOne(v)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
	}
	return nil
}
