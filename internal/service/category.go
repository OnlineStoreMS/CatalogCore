package service

import (
	"fmt"
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
	{Code: "apparel", Name: "装备", NameEn: "APPAREL", Sort: 60},
}

// 禧玛诺已入库系列的归类。通用桶系列 SHIMANO 不保留。
var seriesCategoryByCode = map[string]string{
	"dura-ace":           "road",
	"ultegra":            "road",
	"shimano105":         "road",
	"tiagra":             "road",
	"sora":               "road",
	"claris":             "road",
	"cues":               "mtb",
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

func (s *CatalogService) SaveCategory(tenantID, id uint64, req dto.CategorySaveReq) (*dto.CategoryDTO, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name required", ErrBadRequest)
	}
	if err := SeedStandardCategories(s.repos.DB, tenantID); err != nil {
		return nil, err
	}
	code := strings.TrimSpace(req.Code)
	if code == "" {
		code = strings.ToLower(strings.TrimSpace(req.NameEn))
		code = strings.ReplaceAll(code, " ", "-")
	}
	status := int8(1)
	if req.Status != nil {
		status = *req.Status
	}
	var row model.Category
	if id > 0 {
		if err := s.db(tenantID).First(&row, id).Error; err != nil {
			return nil, mapNotFound(err)
		}
	} else {
		row.TenantID = repo.NormalizeTenantID(tenantID)
	}
	if code != "" {
		var dup model.Category
		err := s.db(tenantID).Where("code = ? AND id <> ?", code, row.ID).First(&dup).Error
		if err == nil {
			return nil, fmt.Errorf("%w: 品类编码已存在", ErrConflict)
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}
	sort := req.Sort
	if id == 0 && sort == 0 {
		var maxSort int
		s.db(tenantID).Model(&model.Category{}).Select("COALESCE(MAX(sort), 0)").Scan(&maxSort)
		sort = maxSort + 10
	}
	row.Name = name
	row.NameEn = strings.TrimSpace(req.NameEn)
	row.Code = code
	row.Sort = sort
	row.Status = status
	var err error
	if id > 0 {
		err = s.repos.DB.Save(&row).Error
	} else {
		err = s.repos.DB.Create(&row).Error
	}
	if err != nil {
		return nil, err
	}
	d := dto.CategoryDTO{
		ID: row.ID, Name: row.Name, NameEn: row.NameEn, Code: row.Code,
		Sort: row.Sort, Status: row.Status,
	}
	return &d, nil
}

func (s *CatalogService) DeleteCategory(tenantID, id uint64) error {
	var cnt int64
	s.db(tenantID).Model(&model.Series{}).Where("category_id = ?", id).Count(&cnt)
	if cnt > 0 {
		return fmt.Errorf("%w: 品类下仍有系列", ErrConflict)
	}
	res := s.db(tenantID).Delete(&model.Category{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
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
