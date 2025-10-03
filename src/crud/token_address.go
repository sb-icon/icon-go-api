package crud

import (
	"sync"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/sb-icon/icon-go-api/models"
)

// TokenAddressCrud - type for tokenAddress table model
type TokenAddressCrud struct {
	db    *gorm.DB
	model *models.TokenAddress
}

var tokenAddressCrud *TokenAddressCrud
var tokenAddressCrudOnce sync.Once

// GetTokenAddressCrud - create and/or return the tokenAddresses table model
func GetTokenAddressCrud() *TokenAddressCrud {
	tokenAddressCrudOnce.Do(func() {
		dbConn := getPostgresConn()
		if dbConn == nil {
			zap.S().Fatal("Cannot connect to postgres database")
		}

		tokenAddressCrud = &TokenAddressCrud{
			db:    dbConn,
			model: &models.TokenAddress{},
		}
	})

	return tokenAddressCrud
}

// SelectMany - select from token_transfers table
// Returns: models, error (if present)
func (m *TokenAddressCrud) SelectMany(
	limit int,
	skip int,
) (*[]models.TokenAddress, error) {
	db := m.db

	// Set table
	db = db.Model(&[]models.TokenAddress{})

	db = db.Order("balance desc")

	// Limit is required and defaulted to 1
	db = db.Limit(limit)

	// Skip
	if skip != 0 {
		db = db.Offset(skip)
	}

	tokenAddresses := &[]models.TokenAddress{}
	db = db.Find(tokenAddresses)

	return tokenAddresses, db.Error
}

// SelectTokensByPublicKey - select token contract addresses by address
func (m *TokenAddressCrud) SelectManyByAddress(
	address string,
) (*[]models.TokenAddress, error) {
	db := m.db

	// Set table
	db = db.Model(&models.TokenAddress{})

	// Public key
	db = db.Where("address = ?", address)

	tokenAddresses := &[]models.TokenAddress{}
	db = db.Find(tokenAddresses)

	return tokenAddresses, db.Error
}

// SelectMany - select from token_transfers table
// Returns: models, error (if present)
func (m *TokenAddressCrud) SelectManyByTokenContractAddress(
	limit int,
	skip int,
	tokenContractAddress string,
) (*[]models.TokenAddress, error) {
	db := m.db

	// Set table
	db = db.Model(&[]models.TokenAddress{})

	db = db.Order("balance desc")

	// Token Contract Address
	db = db.Where("token_contract_address = ?", tokenContractAddress)

	// Limit is required and defaulted to 1
	db = db.Limit(limit)

	// Skip
	if skip != 0 {
		db = db.Offset(skip)
	}

	tokenAddresses := &[]models.TokenAddress{}
	db = db.Find(tokenAddresses)

	return tokenAddresses, db.Error
}

func (m *TokenAddressCrud) CountBy(
	address string,
	tokenContractAddress string,
) (int64, error) {
	db := m.db
	db = db.Model(&models.TokenAddress{})

	if address != "" {
		db = db.Where("address = ?", address)
	}

	if tokenContractAddress != "" {
		db = db.Where("token_contract_address = ?", tokenContractAddress)
	}

	var count int64
	db = db.Count(&count)
	return count, db.Error
}
