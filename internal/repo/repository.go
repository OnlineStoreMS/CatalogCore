package repo

import "gorm.io/gorm"

type Repos struct {
	DB *gorm.DB
}

func New(db *gorm.DB) *Repos {
	return &Repos{DB: db}
}

func NormalizeTenantID(id uint64) uint64 {
	if id == 0 {
		return 1
	}
	return id
}

func (r *Repos) ScopeTenant(tenantID uint64) *gorm.DB {
	return r.DB.Where("tenant_id = ?", NormalizeTenantID(tenantID))
}
