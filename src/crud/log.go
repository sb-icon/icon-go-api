package crud

import (
	"sync"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/sb-icon/icon-go-api/models"
)

// LogCrud - type for log table model
type LogCrud struct {
	db    *gorm.DB
	model *models.Log
}

var logCrud *LogCrud
var logCrudOnce sync.Once

// GetLogCrud - create and/or return the logs table model
func GetLogCrud() *LogCrud {
	logCrudOnce.Do(func() {
		dbConn := getPostgresConn()
		if dbConn == nil {
			zap.S().Fatal("Cannot connect to postgres database")
		}

		logCrud = &LogCrud{
			db:    dbConn,
			model: &models.Log{},
		}
	})

	return logCrud
}

// Select - select from logs table
// Returns: models, error (if present)
func (m *LogCrud) SelectMany(
	limit int,
	skip int,
	blockNumber uint32,
	blockStart uint32,
	blockEnd uint32,
	transactionHash string,
	scoreAddress string,
	method string,
) (*[]models.Log, error) {
	db := m.db

	// Set table
	db = db.Model(&models.Log{})

	// Latest logs first
	db = db.Order("block_number desc")

	// Number
	if blockNumber != 0 {
		db = db.Where("block_number = ?", blockNumber)
	}

	// Number Start
	if blockStart != 0 {
		db = db.Where("block_number >= ?", blockStart)
	}

	// Number End
	if blockEnd != 0 {
		db = db.Where("block_number <= ?", blockEnd)
	}

	// Hash
	if transactionHash != "" {
		db = db.Where("transaction_hash = ?", transactionHash)
	}

	// Address
	if scoreAddress != "" {
		db = db.Where("address = ?", scoreAddress)
	}

	// Method
	if method != "" {
		db = db.Where("method = ?", method)
	}

	// Limit is required and defaulted to 1
	db = db.Limit(limit)

	// Skip
	if skip != 0 {
		db = db.Offset(skip)
	}

	logs := &[]models.Log{}
	db = db.Find(logs)

	return logs, db.Error
}
