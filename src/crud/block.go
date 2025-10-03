package crud

import (
	"sync"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/sb-icon/icon-go-api/models"
)

// BlockCrud - type for block table model
type BlockCrud struct {
	db    *gorm.DB
	model *models.Block
}

var blockCrud *BlockCrud
var blockCrudOnce sync.Once

// GetBlockCrud - create and/or return the blocks table model
func GetBlockCrud() *BlockCrud {
	blockCrudOnce.Do(func() {
		dbConn := getPostgresConn()
		if dbConn == nil {
			zap.S().Fatal("Cannot connect to postgres database")
		}

		blockCrud = &BlockCrud{
			db:    dbConn,
			model: &models.Block{},
		}
	})

	return blockCrud
}

// SelectMany - select from blocks table
// Returns: models, error (if present)
func (m *BlockCrud) SelectMany(
	limit int,
	skip int,
	number uint32,
	startNumber uint32,
	endNumber uint32,
	hash string,
	createdBy string,
	sort string,
) (*[]models.BlockList, error) {
	db := m.db

	// Latest blocks first
	if sort != "" {
		db = db.Order("number " + sort)
	}

	// Set table
	db = db.Model(&[]models.Block{})

	// Number
	if number != 0 {
		db = db.Where("number = ?", number)
	}

	// Start number and end number
	if startNumber != 0 && endNumber != 0 {
		db = db.Where("number BETWEEN ? AND ?", startNumber, endNumber)
	} else if startNumber != 0 {
		db = db.Where("number > ?", startNumber)
	} else if endNumber != 0 {
		db = db.Where("number < ?", endNumber)
	}

	// Hash
	if hash != "" {
		db = db.Where("hash = ?", hash)
	}

	// Created By (peer id)
	if createdBy != "" {
		db = db.Where("peer_id = ?", createdBy)
	}

	// Limit is required and defaulted to 1
	db = db.Limit(limit)

	// Skip
	if skip != 0 {
		db = db.Offset(skip)
	}

	blocks := &[]models.BlockList{}
	db = db.Find(blocks)

	return blocks, db.Error
}

// SelectOne - select from blocks table
func (m *BlockCrud) SelectOne(
	number uint32,
) (*models.Block, error) {
	db := m.db

	db = db.Order("number desc")

	if number != 0 {
		db = db.Where("number = ?", number)
	}

	block := &models.Block{}
	db = db.First(block)

	return block, db.Error
}

// SelectOne - select from blocks table
func (m *BlockCrud) SelectOneByTimestamp(timestamp uint64) (*models.Block, error) {
	db := m.db

	block := &models.Block{}
	db.Raw("SELECT * FROM blocks WHERE timestamp < ? ORDER BY timestamp DESC LIMIT 1;", timestamp).Scan(&block)

	return block, db.Error
}
