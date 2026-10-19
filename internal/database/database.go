package database

import (
	"fmt"
	"os"
	"path/filepath"

	"catalogcore/internal/config"
	"catalogcore/internal/model"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch cfg.Driver {
	case "postgres":
		dialector = postgres.Open(cfg.PostgresDSN)
	case "sqlite":
		if err := os.MkdirAll(filepath.Dir(cfg.SQLitePath), 0o755); err != nil {
			return nil, err
		}
		dialector = sqlite.Open(cfg.SQLitePath)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}
	return gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
}

func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.Brand{},
		&model.Category{},
		&model.Series{},
		&model.Product{},
		&model.Variant{},
		&model.Attachment{},
		&model.CrawlJob{},
		&model.CrawlItem{},
	); err != nil {
		return err
	}
	if db.Dialector.Name() == "postgres" {
		return db.Exec(`
			CREATE INDEX IF NOT EXISTS idx_catalog_products_tenant_brand ON catalog_products (tenant_id, brand_id);
			CREATE INDEX IF NOT EXISTS idx_catalog_products_tenant_model ON catalog_products (tenant_id, model_no);
			CREATE INDEX IF NOT EXISTS idx_catalog_crawl_jobs_tenant_status ON catalog_crawl_jobs (tenant_id, status);
		`).Error
	}
	return nil
}
