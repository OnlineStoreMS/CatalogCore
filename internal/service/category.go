package service

import (
	"strings"

	"catalogcore/internal/dto"
	"catalogcore/internal/model"
	"catalogcore/internal/repo"

	"gorm.io/gorm"
)

type standardCategory struct {
	Code   string
	Name   string
	NameEn string
	Sort   int
}

// 平台标准品类。code 稳定，各品牌的系列挂到同一套编码上。
var standardCategories = []standardCategory{
	{Code: "road", Name: "公路", NameEn: "ROAD", Sort: 10},
	{Code: "mtb", Name: "山地", NameEn: "MTB", Sort: 20},
	{Code: "gravel", Name: "砾石", NameEn: "GRAVEL", Sort: 30},
	{Code: "lifestyle", Name: "休闲城市", NameEn: "LIFESTYLE", Sort: 40},
	{Code: "ebike", Name: "电助力", NameEn: "E-BIKE", Sort: 50},
	{Code: "apparel", Name: "服饰", NameEn: "APPAREL", Sort: 60},
}

// 禧玛诺已入库系列的归类。CUES 跨场景、通用 SHIMANO 不是官网品类，不在这里。
var seriesCategoryByCode = map[string]string{
	"dura-ace":           "road",
	"ultegra":            "road",
	"shimano105":         "road",
	"tiagra":             "road",
	"sora":               "road",
	"claris":             "road",
	"xtr":                "mtb",
	"saint":              "mtb",
	"deorext":            "mtb",
	"deore":              "mtb",
	"altus-acera-alivio": "mtb",
	"grx":                "gravel",
	"essa":               "gravel",
	"tourney":            "lifestyle",
	"nexus":              "lifestyle",
	"alfine":             "lifestyle",
	"ep8":                "ebike",
	"ep6":                "ebike",
	"ep5":                "ebike",
	"e5100":              "ebike",
}

// SeedStandardCategories 为已有租户补齐标准品类，并把尚未归类、编码匹配的系列挂上。
func SeedStandardCategories(db *gorm.DB, tenantIDs ...uint64) error {
	ids := tenantIDs
	if len(ids) == 0 {
		var found []uint64
		if err := db.Model(&model.Brand{}).Distinct("tenant_id").Pluck("tenant_id", &found).Error; err != nil {
			return err
		}
		ids = found
		if len(ids) == 0 {
			ids = []uint64{1}
		}
	}
	seen := map[uint64]struct{}{}
	for _, id := range ids {
		id = repo.NormalizeTenantID(id)
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		if err := ensureTenantCategories(db, id); err != nil {
			return err
		}
		if err := backfillSeriesCategories(db, id); err != nil {
			return err
		}
	}
	return nil
}

func ensureTenantCategories(db *gorm.DB, tenantID uint64) error {
	for _, std := range standardCategories {
		var row model.Category
		err := db.Where("tenant_id = ? AND code = ?", tenantID, std.Code).First(&row).Error
		if err == nil {
			continue
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}
		row = model.Category{
			TenantID: tenantID,
			Name:     std.Name,
			NameEn:   std.NameEn,
			Code:     std.Code,
			Sort:     std.Sort,
			Status:   1,
		}
		if err := db.Create(&row).Error; err != nil {
			return err
		}
	}
	return nil
}

func backfillSeriesCategories(db *gorm.DB, tenantID uint64) error {
	var cats []model.Category
	if err := db.Where("tenant_id = ?", tenantID).Find(&cats).Error; err != nil {
		return err
	}
	byCode := map[string]uint64{}
	for _, c := range cats {
		byCode[c.Code] = c.ID
	}
	var series []model.Series
	if err := db.Where("tenant_id = ? AND (category_id = 0 OR category_id IS NULL)", tenantID).Find(&series).Error; err != nil {
		return err
	}
	for _, row := range series {
		catCode, ok := seriesCategoryByCode[strings.ToLower(strings.TrimSpace(row.Code))]
		if !ok {
			continue
		}
		catID := byCode[catCode]
		if catID == 0 {
			continue
		}
		if err := db.Model(&model.Series{}).Where("id = ?", row.ID).Update("category_id", catID).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *CatalogService) ListCategories(tenantID uint64) ([]dto.CategoryDTO, error) {
	if err := SeedStandardCategories(s.repos.DB, tenantID); err != nil {
		return nil, err
	}
	var list []model.Category
	if err := s.db(tenantID).Order("sort asc, id asc").Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]dto.CategoryDTO, 0, len(list))
	for _, row := range list {
		out = append(out, dto.CategoryDTO{
			ID: row.ID, Name: row.Name, NameEn: row.NameEn, Code: row.Code,
			Sort: row.Sort, Status: row.Status,
		})
	}
	return out, nil
}

func (s *CatalogService) categoryByID(tenantID uint64) (map[uint64]model.Category, error) {
	var list []model.Category
	if err := s.db(tenantID).Find(&list).Error; err != nil {
		return nil, err
	}
	out := make(map[uint64]model.Category, len(list))
	for _, row := range list {
		out[row.ID] = row
	}
	return out, nil
}
