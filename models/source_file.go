package models

import (
	"fmt"
	"payment-bridge/common/constants"
	"payment-bridge/database"
	"strings"

	"github.com/filswan/go-swan-lib/logs"
	"github.com/shopspring/decimal"
)

type SourceFile struct {
	ID          int64  `json:"id"`
	ResourceUri string `json:"resource_uri"`
	Status      string `json:"status"`
	FileSize    int64  `json:"file_size"`
	Dataset     string `json:"dataset"`
	IpfsUrl     string `json:"ipfs_url"`
	PinStatus   string `json:"pin_status"`
	PayloadCid  string `json:"payload_cid"`
	NftTxHash   string `json:"nft_tx_hash"`
	TokenId     string `json:"token_id"`
	MintAddress string `json:"mint_address"`
	FileType    int    `json:"file_type"`
	CreateAt    int64  `json:"create_at"`
	UpdateAt    int64  `json:"update_at"`
}

type SourceFileExt struct {
	SourceFile
	DealFileId   int64            `json:"deal_file_id"`
	Duration     int              `json:"duration"`
	LockedFee    *decimal.Decimal `json:"locked_fee"`
	OfflineDeals []*OfflineDeal   `json:"offline_deals"`
}

func GetSourceFileById(id int64) (*SourceFile, error) {
	var sourceFile SourceFile

	err := database.GetDB().Where("id=?", id).First(&sourceFile).Error
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return &sourceFile, nil
}

func GetSourceFilesByPayloadCid(payloadCid string) ([]*SourceFile, error) {
	var sourceFiles []*SourceFile

	err := database.GetDB().Where("payload_cid=?", payloadCid).Order("create_at").Find(&sourceFiles).Error
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return sourceFiles, nil
}

func GetSourceFileByPayloadCid(payloadCid string) ([]*SourceFile, error) {
	var sourceFiles []*SourceFile

	err := database.GetDB().Where("payload_cid=?", payloadCid).Order("create_at").Find(&sourceFiles).Error
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return sourceFiles, nil
}

func GetSourceFileByPayloadCidWalletAddress(payloadCid, walletAddress string) (*SourceFile, error) {
	var sourceFiles []*SourceFile

	err := database.GetDB().Where("payload_cid=? and wallet_address=?", payloadCid, walletAddress).Order("create_at").Find(&sourceFiles).Error
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if len(sourceFiles) > 0 {
		return sourceFiles[0], nil
	}

	err = fmt.Errorf("source file with payload_cid:%s, wallet_address:%s not exists", payloadCid, walletAddress)
	logs.GetLogger().Error(err)
	return nil, err
}

func GetSourceFilesNeed2Car() ([]*SourceFileExt, error) {
	var sourceFiles []*SourceFileExt
	sql := "select a.*,b.locked_fee from source_file a, event_lock_payment b where b.source_file_id=a.id and a.status=? and a.file_type=?"
	err := database.GetDB().Raw(sql, constants.SOURCE_FILE_STATUS_CREATED, constants.SOURCE_FILE_TYPE_NORMAL).Scan(&sourceFiles).Error

	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return sourceFiles, nil
}

func CreateSourceFile(sourceFile SourceFile) (*SourceFile, error) {
	value, err := database.SaveOneWithResult(&sourceFile)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	sourceFileCreated := value.(*SourceFile)

	return sourceFileCreated, nil
}

func GetSourceFiles(limit, offset string, walletAddress, payloadCid string) ([]*SourceFileExt, error) {
	sql := "select s.id, h.file_name,s.file_size,s.pin_status,s.create_at,s.payload_cid,s.ipfs_url,h.wallet_address,s.mint_address, s.nft_tx_hash, s.token_id,df.id deal_file_id,df.lock_payment_status status,df.duration, evpm.locked_fee from source_file s "
	sql = sql + "left join source_file_upload_history h on s.id=h.source_file_id "
	sql = sql + "left join source_file_deal_file_map sfdfm on s.id = sfdfm.source_file_id "
	sql = sql + "left join deal_file df on sfdfm.deal_file_id = df.id "

	params := []interface{}{}

	if strings.Trim(payloadCid, " ") != "" {
		sql = sql + " and s.payload_cid=?"
		params = append(params, payloadCid)
	}

	sql = sql + "left outer join event_lock_payment evpm on evpm.payload_cid = s.payload_cid "
	sql = sql + "where h.wallet_address=? and s.file_type=?"
	params = append(params, walletAddress, constants.SOURCE_FILE_TYPE_NORMAL)

	var results []*SourceFileExt

	err := database.GetDB().Raw(sql, params...).Order("create_at desc").Limit(limit).Offset(offset).Scan(&results).Error
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return results, nil
}

func GetSourceFilesByWalletAddress(walletAddress string) ([]*SourceFileExt, error) {
	sql := "select s.id, h.file_name,s.file_size,s.pin_status,s.create_at,s.payload_cid,s.ipfs_url,h.wallet_address,s.mint_address, s.nft_tx_hash, s.token_id,df.id deal_file_id,df.lock_payment_status status,df.duration, evpm.locked_fee from source_file s "
	sql = sql + "left join source_file_upload_history h on s.id=h.source_file_id "
	sql = sql + "left join source_file_deal_file_map sfdfm on s.id = sfdfm.source_file_id "
	sql = sql + "left join deal_file df on sfdfm.deal_file_id = df.id "

	params := []interface{}{}
	sql = sql + "left outer join event_lock_payment evpm on evpm.payload_cid = s.payload_cid "
	sql = sql + "where h.wallet_address=? and s.file_type=?"
	params = append(params, walletAddress, constants.SOURCE_FILE_TYPE_NORMAL)

	var results []*SourceFileExt

	err := database.GetDB().Raw(sql, params...).Order("create_at desc").Scan(&results).Error
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return results, nil
}

func GetSourceFilesByDealFileId(dealFileId int64) ([]*SourceFile, error) {
	var sourceFiles []*SourceFile

	sql := "select a.* from source_file a, source_file_deal_file_map b where a.id=b.source_file_id and b.deal_file_id=?"

	err := database.GetDB().Raw(sql, dealFileId).Scan(&sourceFiles).Error
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return sourceFiles, nil
}
