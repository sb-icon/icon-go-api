package crud

import (
	"strconv"
	"sync"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/sudoblockio/icon-go-api/models"
)

// TokenTransferCrud - type for tokenTransfer table model
type TokenTransferCrud struct {
	db    *gorm.DB
	model *models.TokenTransfer
}

var tokenTransferCrud *TokenTransferCrud
var tokenTransferCrudOnce sync.Once

// GetTokenTransferCrud - create and/or return the tokenTransfers table model
func GetTokenTransferCrud() *TokenTransferCrud {
	tokenTransferCrudOnce.Do(func() {
		dbConn := getPostgresConn()
		if dbConn == nil {
			zap.S().Fatal("Cannot connect to postgres database")
		}

		tokenTransferCrud = &TokenTransferCrud{
			db:    dbConn,
			model: &models.TokenTransfer{},
		}
	})

	return tokenTransferCrud
}

// SelectOne - select from token_transfers table
// Returns: models, error (if present)
func (m *TokenTransferCrud) SelectOne(
	transactionHash string,
	logIndex int32,
) (*models.TokenTransfer, error) {
	db := m.db

	// Set table
	db = db.Model(&[]models.TokenTransfer{})

	// Transaction Hash
	db = db.Where("transaction_hash = ?", transactionHash)

	// Log Index
	db = db.Where("log_index = ?", logIndex)

	tokenTransfer := &models.TokenTransfer{}
	db = db.First(tokenTransfer)

	return tokenTransfer, db.Error
}

// SelectMany - select from token_transfers table
// Returns: models, error (if present)
func (m *TokenTransferCrud) SelectMany(
	limit int,
	skip int,
	from string,
	to string,
	blockNumber int,
	startBlockNumber int,
	endBlockNumber int,
	transactionHash string,
	tokenContractAddress string,
) (*[]models.TokenTransfer, error) {
	db := m.db

	// Set table
	db = db.Model(&[]models.TokenTransfer{})

	// Latest transactions first
	db = db.Order("block_number desc")
	db = db.Order("transaction_index desc")
	db = db.Order("log_index desc")

	// from
	if from != "" {
		db = db.Where("from_address = ?", from)
	}

	// to
	if to != "" {
		db = db.Where("to_address = ?", to)
	}

	// block number
	if blockNumber != 0 {
		db = db.Where("block_number = ?", blockNumber)
	}

	// start block number
	if startBlockNumber != 0 {
		db = db.Where("block_number >= ?", startBlockNumber)
	}

	// end block number
	if endBlockNumber != 0 {
		db = db.Where("block_number <= ?", endBlockNumber)
	}

	// transaction hash
	if transactionHash != "" {
		db = db.Where("transaction_hash = ?", transactionHash)
	}

	// token contract address
	if tokenContractAddress != "" {
		db = db.Where("token_contract_address = ?", tokenContractAddress)
	}

	// Limit is required and defaulted to 1
	db = db.Limit(limit)

	// Skip
	if skip != 0 {
		db = db.Offset(skip)
	}

	tokenTransfers := &[]models.TokenTransfer{}
	db = db.Find(tokenTransfers)

	return tokenTransfers, db.Error
}

// SelectManyByAddress - select from token_transfers table by address
// Returns: models, error (if present)
func (m *TokenTransferCrud) SelectManyByAddress(
	limit int,
	skip int,
	address string,
) (*[]models.TokenTransfer, error) {
	db := m.db

	// Set table
	db = db.Model(&[]models.TokenTransfer{})

	// Latest transactions first
	db = db.Order("block_number desc")

	// Address
	db = db.Where(`(transaction_hash, log_index)
	IN (
		SELECT
			transaction_hash, log_index
		FROM
			token_transfer_by_addresses
		WHERE
			address = ?
		ORDER BY block_number desc
		LIMIT ?
		OFFSET ?
	)`, address, limit, skip)

	tokenTransfers := &[]models.TokenTransfer{}
	db = db.Find(tokenTransfers)

	return tokenTransfers, db.Error
}

func (m *TokenTransferCrud) SelectManyByAddressBlockRange(
	limit int,
	skip int,
	address string,
	startBlockNumber int,
	endBlockNumber int,
) (*[]models.TokenTransfer, error) {
	db := m.db

	// Set table
	db = db.Model(&[]models.TokenTransfer{})

	blockWhereCondition := ""
	if startBlockNumber != 0 || endBlockNumber != 0 {
		blockWhereCondition = blockWhereCondition + "and"
	}

	// start block number
	if startBlockNumber != 0 {
		blockWhereCondition = blockWhereCondition + " > " + strconv.Itoa(startBlockNumber)
	}

	// end block number
	if endBlockNumber != 0 {
		blockWhereCondition = blockWhereCondition + " > " + strconv.Itoa(startBlockNumber)
	}

	// Latest transactions first
	db = db.Order("block_number desc")

	// Address
	db = db.Where(`(transaction_hash, log_index)
	IN (
		SELECT
			transaction_hash, log_index
		FROM
			token_transfer_by_addresses
		WHERE
			address = ? ?
		ORDER BY block_number desc
		LIMIT ?
		OFFSET ?
	)`, address, blockWhereCondition, limit, skip)

	tokenTransfers := &[]models.TokenTransfer{}
	db = db.Find(tokenTransfers)

	return tokenTransfers, db.Error
}

// SelectManyByTokenContracAddress - select from token_transfers table by token contract address
// Returns: models, error (if present)
func (m *TokenTransferCrud) SelectManyByTokenContractAddress(
	limit int,
	skip int,
	tokenContractAddress string,
) (*[]models.TokenTransfer, error) {
	db := m.db

	// Set table
	db = db.Model(&[]models.TokenTransfer{})

	// Latest transactions first
	db = db.Order("block_number desc")

	// address
	db = db.Where("token_contract_address = ?", tokenContractAddress)

	// Limit is required and defaulted to 1
	db = db.Limit(limit)

	// Skip
	if skip != 0 {
		db = db.Offset(skip)
	}

	tokenTransfers := &[]models.TokenTransfer{}
	db = db.Find(tokenTransfers)

	return tokenTransfers, db.Error
}
