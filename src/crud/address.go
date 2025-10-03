package crud

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/sb-icon/icon-go-api/models"
)

// AddressCrud - type for address table model
type AddressCrud struct {
	db    *gorm.DB
	model *models.Address
}

var addressCrud *AddressCrud
var addressCrudOnce sync.Once

// GetAddressCrud - create and/or return the addresss table model
func GetAddressCrud() *AddressCrud {
	addressCrudOnce.Do(func() {
		dbConn := getPostgresConn()
		if dbConn == nil {
			zap.S().Fatal("Cannot connect to postgres database")
		}

		addressCrud = &AddressCrud{
			db:    dbConn,
			model: &models.Address{},
		}
	})

	return addressCrud
}

// SelectOne - select one from addresses table
func (m *AddressCrud) SelectOne(
	_address string,
) (*models.Address, error) {
	db := m.db

	// Set table
	db = db.Model(&models.Address{})

	// Address
	db = db.Where("address = ?", _address)

	address := &models.Address{}
	db = db.First(address)

	return address, db.Error
}

// SelectMany - select many from addreses table
func (m *AddressCrud) SelectMany(
	limit int,
	skip int,
	address string,
	isContract *bool,
	isToken *bool,
	isNft *bool,
	tokenStandard string,
	sort string,
) (*[]models.AddressList, error) {
	db := m.db

	// Set table
	db = db.Model(&models.Address{})

	if isContract != nil {
		db = db.Where("is_contract = ?", &isContract)
	}

	if isToken != nil {
		db = db.Where("is_token = ?", &isToken)
	}

	if isNft != nil {
		db = db.Where("is_nft = ?", &isNft)
	}

	// Address
	if address != "" {
		db = db.Where("address = ?", address)
	}

	// Token Standard
	if tokenStandard != "" {
		db = db.Where("token_standard = ?", tokenStandard)
	}

	// Limit
	db = db.Limit(limit)

	// Skip
	if skip != 0 {
		db = db.Offset(skip)
	}

	// Order by
	if sort == "" {
		db = db.Order("balance DESC")
	} else {
		orderVal := sort[0:1]

		var orderCol string
		var order string

		if orderVal == "-" {
			orderCol = sort[1:]
			order = "ASC"
		} else {
			orderCol = sort
			order = "DESC"
		}
		db = db.Order(fmt.Sprintf("%s %s", orderCol, order))
	}

	addresses := &[]models.AddressList{}
	db = db.Find(addresses)

	return addresses, db.Error
}

func (m *AddressCrud) CountWithParamsSearch(
	search string,
	tokenStandard string,
	status string,
	isToken *bool,
	isNft *bool,
	isContract *bool,
) (int64, error) {
	db := m.db
	db = db.Model(&models.Address{})

	// Support search functionality
	if search != "" {
		db.Where("LOWER(name) LIKE LOWER(?)", fmt.Sprintf("%%%s%%", search))
	}

	if tokenStandard != "" {
		db.Where("is_token = true")
		db.Where("token_standard = ?", tokenStandard)
	}

	if status != "" {
		db.Where("status = ?", status)
	}

	if isToken != nil {
		db = db.Where("is_token = ?", &isToken)
	}

	if isNft != nil {
		db = db.Where("is_nft = ?", &isNft)
	}

	if isContract != nil {
		db = db.Where("is_contract = ?", &isContract)
	}

	// Is contract
	db = db.Where("is_contract = true")

	var count int64
	db = db.Count(&count)
	return count, db.Error
}

// SelectManyContracts - select many from addresses table
func (m *AddressCrud) SelectManyContracts(
	search string,
	tokenStandard string,
	isToken *bool,
	isNft *bool,
	status string,
	limit int,
	skip int,
	sort string,
) (*[]models.ContractList, error) {
	db := m.db
	db = db.Model(&models.Address{})

	// Support search functionality
	if search != "" {
		db.Where("LOWER(name) LIKE LOWER(?)", fmt.Sprintf("%%%s%%", search))
	}

	if tokenStandard != "" {
		db.Where("is_token = true")
		db.Where("token_standard = ?", tokenStandard)
	}

	if status != "" {
		db.Where("status = ?", status)
	}

	if isToken != nil {
		db = db.Where("is_token = ?", &isToken)
	}

	if isNft != nil {
		db = db.Where("is_nft = ?", &isNft)
	}

	// Is contract
	db = db.Where("is_contract = true")

	// Order by
	if sort == "" {
		// Order by name with nulls in the back
		db = db.Order("nullif(name, '') nulls last")
	} else {
		orderVal := sort[0:1]

		var orderCol string
		var order string

		if orderVal == "-" {
			orderCol = sort[1:]
			order = "ASC"
		} else {
			orderCol = sort
			order = "DESC"
		}
		db = db.Order(fmt.Sprintf("%s %s", orderCol, order))
	}

	// Limit
	db = db.Limit(limit)

	// Skip
	if skip != 0 {
		db = db.Offset(skip)
	}

	contracts := &[]models.ContractList{}
	db = db.Find(contracts)

	return contracts, db.Error
}
