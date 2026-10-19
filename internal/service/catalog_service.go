package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"catalogcore/internal/crawl"
	"catalogcore/internal/dto"
	"catalogcore/internal/model"
	"catalogcore/internal/repo"
	"catalogcore/internal/storage"

	"gorm.io/gorm"
)

type CatalogService struct {
	repos   *repo.Repos
	store   storage.Storage
	fetch   *crawl.Fetcher
	mu      sync.Mutex
	running map[uint64]context.CancelFunc
}

func NewCatalogService(repos *repo.Repos, store storage.Storage) *CatalogService {
	return &CatalogService{
		repos:   repos,
		store:   store,
		fetch:   crawl.NewFetcher(),
		running: map[uint64]context.CancelFunc{},
	}
}

func (s *CatalogService) db(tenantID uint64) *gorm.DB {
	return s.repos.ScopeTenant(tenantID)
}

func (s *CatalogService) resolveURL(u string) string {
	if s.store == nil {
		return u
	}
	return s.store.ResolvePublicURL(u)
}

// ── Brand ──

func (s *CatalogService) ListBrands(tenantID uint64, keyword string) ([]dto.BrandDTO, error) {
	q := s.db(tenantID).Model(&model.Brand{}).Order("sort desc, id desc")
	if kw := strings.TrimSpace(keyword); kw != "" {
		like := "%" + kw + "%"
		q = q.Where("name ILIKE ? OR name_en ILIKE ? OR code ILIKE ?", like, like, like)
	}
	var list []model.Brand
	if err := q.Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]dto.BrandDTO, 0, len(list))
	for i := range list {
		out = append(out, s.toBrandDTO(&list[i]))
	}
	return out, nil
}

func (s *CatalogService) SaveBrand(tenantID, id uint64, req dto.BrandSaveReq) (*dto.BrandDTO, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name required", ErrBadRequest)
	}
	status := int8(1)
	if req.Status != nil {
		status = *req.Status
	}
	var row model.Brand
	if id > 0 {
		if err := s.db(tenantID).First(&row, id).Error; err != nil {
			return nil, mapNotFound(err)
		}
	} else {
		row.TenantID = repo.NormalizeTenantID(tenantID)
	}
	row.Name = name
	row.NameEn = strings.TrimSpace(req.NameEn)
	row.Code = strings.TrimSpace(req.Code)
	row.LogoURL = strings.TrimSpace(req.LogoURL)
	row.Website = strings.TrimSpace(req.Website)
	row.Country = strings.TrimSpace(req.Country)
	row.Description = strings.TrimSpace(req.Description)
	row.Sort = req.Sort
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
	d := s.toBrandDTO(&row)
	return &d, nil
}

func (s *CatalogService) DeleteBrand(tenantID, id uint64) error {
	var cnt int64
	s.db(tenantID).Model(&model.Product{}).Where("brand_id = ?", id).Count(&cnt)
	if cnt > 0 {
		return fmt.Errorf("%w: 品牌下仍有产品", ErrConflict)
	}
	res := s.db(tenantID).Delete(&model.Brand{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *CatalogService) toBrandDTO(b *model.Brand) dto.BrandDTO {
	return dto.BrandDTO{
		ID: b.ID, Name: b.Name, NameEn: b.NameEn, Code: b.Code, LogoURL: s.resolveURL(b.LogoURL),
		Website: b.Website, Country: b.Country, Description: b.Description,
		Sort: b.Sort, Status: b.Status,
		CreatedAt: b.CreatedAt.Format(time.RFC3339),
		UpdatedAt: b.UpdatedAt.Format(time.RFC3339),
	}
}

// ── Series ──

func (s *CatalogService) ListSeries(tenantID, brandID uint64) ([]dto.SeriesDTO, error) {
	if err := SeedStandardCategories(s.repos.DB, tenantID); err != nil {
		return nil, err
	}
	cats, err := s.categoryByID(tenantID)
	if err != nil {
		return nil, err
	}
	q := s.db(tenantID).Model(&model.Series{}).Order("sort asc, id asc")
	if brandID > 0 {
		q = q.Where("brand_id = ?", brandID)
	}
	var list []model.Series
	if err := q.Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]dto.SeriesDTO, 0, len(list))
	for i := range list {
		out = append(out, s.toSeriesDTO(&list[i], cats))
	}
	return out, nil
}

func (s *CatalogService) SaveSeries(tenantID, id uint64, req dto.SeriesSaveReq) (*dto.SeriesDTO, error) {
	if strings.TrimSpace(req.Name) == "" || req.BrandID == 0 {
		return nil, ErrBadRequest
	}
	var brand model.Brand
	if err := s.db(tenantID).First(&brand, req.BrandID).Error; err != nil {
		return nil, mapNotFound(err)
	}
	if err := SeedStandardCategories(s.repos.DB, tenantID); err != nil {
		return nil, err
	}
	if req.CategoryID > 0 {
		var cat model.Category
		if err := s.db(tenantID).First(&cat, req.CategoryID).Error; err != nil {
			return nil, fmt.Errorf("%w: 品类不存在", ErrBadRequest)
		}
	}
	status := int8(1)
	if req.Status != nil {
		status = *req.Status
	}
	var row model.Series
	if id > 0 {
		if err := s.db(tenantID).First(&row, id).Error; err != nil {
			return nil, mapNotFound(err)
		}
	} else {
		row.TenantID = repo.NormalizeTenantID(tenantID)
	}
	row.BrandID = req.BrandID
	row.CategoryID = req.CategoryID
	row.Name = strings.TrimSpace(req.Name)
	row.Code = strings.TrimSpace(req.Code)
	row.Description = strings.TrimSpace(req.Description)
	row.CoverURL = strings.TrimSpace(req.CoverURL)
	row.Sort = req.Sort
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
	cats, err := s.categoryByID(tenantID)
	if err != nil {
		return nil, err
	}
	d := s.toSeriesDTO(&row, cats)
	return &d, nil
}

func (s *CatalogService) toSeriesDTO(row *model.Series, cats map[uint64]model.Category) dto.SeriesDTO {
	d := dto.SeriesDTO{
		ID: row.ID, BrandID: row.BrandID, CategoryID: row.CategoryID,
		Name: row.Name, Code: row.Code,
		Description: row.Description, CoverURL: s.resolveURL(row.CoverURL),
		Sort: row.Sort, Status: row.Status,
	}
	if cat, ok := cats[row.CategoryID]; ok {
		d.CategoryName = cat.Name
		d.CategoryNameEn = cat.NameEn
		d.CategoryCode = cat.Code
	}
	return d
}

func (s *CatalogService) DeleteSeries(tenantID, id uint64) error {
	res := s.db(tenantID).Delete(&model.Series{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ── Product ──

func (s *CatalogService) ListProducts(tenantID uint64, q dto.ProductListQuery) ([]dto.ProductDTO, int64, error) {
	page, pageSize := q.Page, q.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	dbq := s.db(tenantID).Model(&model.Product{})
	if q.BrandID > 0 {
		dbq = dbq.Where("brand_id = ?", q.BrandID)
	}
	if q.SeriesID > 0 {
		dbq = dbq.Where("series_id = ?", q.SeriesID)
	}
	if q.Status != nil {
		dbq = dbq.Where("status = ?", *q.Status)
	}
	if kw := strings.TrimSpace(q.Keyword); kw != "" {
		like := "%" + kw + "%"
		dbq = dbq.Where("name ILIKE ? OR model_no ILIKE ? OR subtitle ILIKE ?", like, like, like)
	}
	var total int64
	if err := dbq.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Product
	if err := dbq.Preload("Variants").
		Order("CASE WHEN parent_id = 0 THEN id ELSE parent_id END, parent_id, id").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	brandNames := s.brandNameMap(tenantID)
	seriesNames := s.seriesNameMap(tenantID)
	parentNames := s.productNameMap(tenantID, parentIDs(list))
	out := make([]dto.ProductDTO, 0, len(list))
	for i := range list {
		d := s.toProductDTO(&list[i], false)
		d.BrandName = brandNames[list[i].BrandID]
		d.SeriesName = seriesNames[list[i].SeriesID]
		d.ParentName = parentNames[list[i].ParentID]
		out = append(out, d)
	}
	return out, total, nil
}

func (s *CatalogService) GetProduct(tenantID, id uint64) (*dto.ProductDTO, error) {
	var row model.Product
	if err := s.db(tenantID).Preload("Variants").Preload("Attachments").First(&row, id).Error; err != nil {
		return nil, mapNotFound(err)
	}
	d := s.toProductDTO(&row, true)
	brandNames := s.brandNameMap(tenantID)
	seriesNames := s.seriesNameMap(tenantID)
	d.BrandName = brandNames[row.BrandID]
	d.SeriesName = seriesNames[row.SeriesID]
	if row.ParentID > 0 {
		names := s.productNameMap(tenantID, []uint64{row.ParentID})
		d.ParentName = names[row.ParentID]
	}
	return &d, nil
}

func (s *CatalogService) SaveProduct(tenantID, userID, id uint64, req dto.ProductSaveReq) (*dto.ProductDTO, error) {
	if strings.TrimSpace(req.Name) == "" || req.BrandID == 0 {
		return nil, ErrBadRequest
	}
	var brand model.Brand
	if err := s.db(tenantID).First(&brand, req.BrandID).Error; err != nil {
		return nil, mapNotFound(err)
	}
	if req.SeriesID > 0 {
		var series model.Series
		if err := s.db(tenantID).Where("id = ? AND brand_id = ?", req.SeriesID, req.BrandID).First(&series).Error; err != nil {
			return nil, fmt.Errorf("%w: 系列不存在或不属于该品牌", ErrBadRequest)
		}
	}
	if req.ParentID > 0 {
		if req.ParentID == id {
			return nil, fmt.Errorf("%w: 不能把产品挂到自己下面", ErrBadRequest)
		}
		var parent model.Product
		if err := s.db(tenantID).Where("id = ? AND brand_id = ?", req.ParentID, req.BrandID).First(&parent).Error; err != nil {
			return nil, fmt.Errorf("%w: 大套不存在或不属于该品牌", ErrBadRequest)
		}
		if parent.ParentID > 0 {
			return nil, fmt.Errorf("%w: 只能挂在大套下面", ErrBadRequest)
		}
	}
	status := int8(1)
	if req.Status != nil {
		status = *req.Status
	}
	err := s.repos.DB.Transaction(func(tx *gorm.DB) error {
		var row model.Product
		if id > 0 {
			if err := tx.Scopes(func(db *gorm.DB) *gorm.DB {
				return db.Where("tenant_id = ?", repo.NormalizeTenantID(tenantID))
			}).First(&row, id).Error; err != nil {
				return mapNotFound(err)
			}
		} else {
			row.TenantID = repo.NormalizeTenantID(tenantID)
			row.CreatedBy = userID
		}
		row.BrandID = req.BrandID
		row.SeriesID = req.SeriesID
		row.ParentID = req.ParentID
		row.Name = strings.TrimSpace(req.Name)
		row.ModelNo = strings.TrimSpace(req.ModelNo)
		row.Subtitle = strings.TrimSpace(req.Subtitle)
		row.CoverURL = strings.TrimSpace(req.CoverURL)
		row.Intro = req.Intro
		row.DetailHTML = req.DetailHTML
		row.SpecsJSON = req.SpecsJSON
		row.TagsJSON = req.TagsJSON
		row.SourceURL = strings.TrimSpace(req.SourceURL)
		row.SourceSite = strings.TrimSpace(req.SourceSite)
		row.Status = status
		if id > 0 {
			if err := tx.Save(&row).Error; err != nil {
				return err
			}
			if err := tx.Where("product_id = ?", row.ID).Delete(&model.Variant{}).Error; err != nil {
				return err
			}
			if err := tx.Where("product_id = ?", row.ID).Delete(&model.Attachment{}).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		id = row.ID
		for i, v := range req.Variants {
			name := strings.TrimSpace(v.Name)
			if name == "" {
				continue
			}
			st := v.Status
			if st == 0 {
				st = 1
			}
			item := model.Variant{
				TenantID:  repo.NormalizeTenantID(tenantID),
				ProductID: row.ID,
				Name:      name,
				SKUCode:   strings.TrimSpace(v.SKUCode),
				SpecJSON:  v.SpecJSON,
				PriceHint: strings.TrimSpace(v.PriceHint),
				PicURL:    strings.TrimSpace(v.PicURL),
				Sort:      v.Sort,
				Status:    st,
			}
			if item.Sort == 0 {
				item.Sort = i
			}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}
		for i, a := range req.Attachments {
			if strings.TrimSpace(a.URL) == "" {
				continue
			}
			kind := strings.TrimSpace(a.Kind)
			if kind == "" {
				kind = "other"
			}
			item := model.Attachment{
				TenantID:  repo.NormalizeTenantID(tenantID),
				ProductID: row.ID,
				Kind:      kind,
				Name:      strings.TrimSpace(a.Name),
				URL:       strings.TrimSpace(a.URL),
				Mime:      strings.TrimSpace(a.Mime),
				SizeBytes: a.SizeBytes,
				Sort:      a.Sort,
			}
			if item.Sort == 0 {
				item.Sort = i
			}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.GetProduct(tenantID, id)
}

func (s *CatalogService) DeleteProduct(tenantID, id uint64) error {
	var cnt int64
	s.db(tenantID).Model(&model.Product{}).Where("parent_id = ?", id).Count(&cnt)
	if cnt > 0 {
		return fmt.Errorf("%w: 大套下仍有产品", ErrConflict)
	}
	return s.repos.DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Where("tenant_id = ?", repo.NormalizeTenantID(tenantID)).Delete(&model.Product{}, id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}
		_ = tx.Where("product_id = ?", id).Delete(&model.Variant{}).Error
		_ = tx.Where("product_id = ?", id).Delete(&model.Attachment{}).Error
		return nil
	})
}

func (s *CatalogService) toProductDTO(p *model.Product, full bool) dto.ProductDTO {
	d := dto.ProductDTO{
		ID: p.ID, BrandID: p.BrandID, SeriesID: p.SeriesID, ParentID: p.ParentID,
		Name: p.Name, ModelNo: p.ModelNo, Subtitle: p.Subtitle,
		CoverURL: s.resolveURL(p.CoverURL), Intro: p.Intro,
		SpecsJSON: p.SpecsJSON, TagsJSON: p.TagsJSON,
		SourceURL: p.SourceURL, SourceSite: p.SourceSite,
		CrawlJobID: p.CrawlJobID, Status: p.Status,
		CreatedAt: p.CreatedAt.Format(time.RFC3339),
		UpdatedAt: p.UpdatedAt.Format(time.RFC3339),
	}
	if full {
		d.DetailHTML = p.DetailHTML
	}
	for _, v := range p.Variants {
		d.Variants = append(d.Variants, dto.VariantDTO{
			ID: v.ID, Name: v.Name, SKUCode: v.SKUCode, SpecJSON: v.SpecJSON,
			PriceHint: v.PriceHint, PicURL: s.resolveURL(v.PicURL), Sort: v.Sort, Status: v.Status,
		})
	}
	for _, a := range p.Attachments {
		d.Attachments = append(d.Attachments, dto.AttachmentDTO{
			ID: a.ID, Kind: a.Kind, Name: a.Name, URL: s.resolveURL(a.URL),
			Mime: a.Mime, SizeBytes: a.SizeBytes, Sort: a.Sort,
		})
	}
	return d
}

func parentIDs(list []model.Product) []uint64 {
	seen := map[uint64]struct{}{}
	out := make([]uint64, 0)
	for _, p := range list {
		if p.ParentID == 0 {
			continue
		}
		if _, ok := seen[p.ParentID]; ok {
			continue
		}
		seen[p.ParentID] = struct{}{}
		out = append(out, p.ParentID)
	}
	return out
}

func (s *CatalogService) productNameMap(tenantID uint64, ids []uint64) map[uint64]string {
	m := map[uint64]string{}
	if len(ids) == 0 {
		return m
	}
	var list []model.Product
	_ = s.db(tenantID).Select("id, name").Where("id IN ?", ids).Find(&list).Error
	for _, p := range list {
		m[p.ID] = p.Name
	}
	return m
}

func (s *CatalogService) brandNameMap(tenantID uint64) map[uint64]string {
	var list []model.Brand
	_ = s.db(tenantID).Select("id, name").Find(&list).Error
	m := map[uint64]string{}
	for _, b := range list {
		m[b.ID] = b.Name
	}
	return m
}

func (s *CatalogService) seriesNameMap(tenantID uint64) map[uint64]string {
	var list []model.Series
	_ = s.db(tenantID).Select("id, name").Find(&list).Error
	m := map[uint64]string{}
	for _, b := range list {
		m[b.ID] = b.Name
	}
	return m
}

func mapNotFound(err error) error {
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}
	return err
}

// ── Crawl ──

func (s *CatalogService) ListCrawlJobs(tenantID uint64, page, pageSize int) ([]dto.CrawlJobDTO, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	q := s.db(tenantID).Model(&model.CrawlJob{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.CrawlJob
	if err := q.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	out := make([]dto.CrawlJobDTO, 0, len(list))
	for i := range list {
		out = append(out, toCrawlJobDTO(&list[i]))
	}
	return out, total, nil
}

func (s *CatalogService) GetCrawlJob(tenantID, id uint64) (*dto.CrawlJobDTO, []dto.CrawlItemDTO, error) {
	var job model.CrawlJob
	if err := s.db(tenantID).First(&job, id).Error; err != nil {
		return nil, nil, mapNotFound(err)
	}
	var items []model.CrawlItem
	_ = s.db(tenantID).Where("job_id = ?", id).Order("id asc").Find(&items).Error
	itemDTOs := make([]dto.CrawlItemDTO, 0, len(items))
	for _, it := range items {
		itemDTOs = append(itemDTOs, dto.CrawlItemDTO{
			ID: it.ID, JobID: it.JobID, SourceURL: it.SourceURL, Status: it.Status,
			Title: it.Title, ProductID: it.ProductID, ErrorMsg: it.ErrorMsg,
		})
	}
	d := toCrawlJobDTO(&job)
	return &d, itemDTOs, nil
}

func (s *CatalogService) CreateCrawlJob(tenantID, userID uint64, req dto.CrawlJobCreateReq) (*dto.CrawlJobDTO, error) {
	urls := crawl.SplitSeedURLs(req.SeedURLs)
	if len(urls) == 0 {
		return nil, fmt.Errorf("%w: 至少需要一个 URL", ErrBadRequest)
	}
	adapter := strings.TrimSpace(req.Adapter)
	if adapter == "" {
		adapter = "generic_product"
	}
	job := model.CrawlJob{
		TenantID:  repo.NormalizeTenantID(tenantID),
		Name:      strings.TrimSpace(req.Name),
		Adapter:   adapter,
		BrandID:   req.BrandID,
		SeriesID:  req.SeriesID,
		SeedURLs:  strings.Join(urls, "\n"),
		Status:    model.CrawlJobPending,
		Total:     len(urls),
		CreatedBy: userID,
	}
	if err := s.repos.DB.Create(&job).Error; err != nil {
		return nil, err
	}
	items := make([]model.CrawlItem, 0, len(urls))
	for _, u := range urls {
		items = append(items, model.CrawlItem{
			TenantID:  repo.NormalizeTenantID(tenantID),
			JobID:     job.ID,
			SourceURL: u,
			Status:    "pending",
		})
	}
	if err := s.repos.DB.Create(&items).Error; err != nil {
		return nil, err
	}
	d := toCrawlJobDTO(&job)
	return &d, nil
}

func (s *CatalogService) StartCrawlJob(tenantID, id uint64) error {
	var job model.CrawlJob
	if err := s.db(tenantID).First(&job, id).Error; err != nil {
		return mapNotFound(err)
	}
	if job.Status == model.CrawlJobRunning {
		return fmt.Errorf("%w: 任务已在运行", ErrConflict)
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.mu.Lock()
	if old, ok := s.running[id]; ok {
		old()
	}
	s.running[id] = cancel
	s.mu.Unlock()

	now := time.Now()
	_ = s.repos.DB.Model(&job).Updates(map[string]any{
		"status": model.CrawlJobRunning, "started_at": now, "message": "",
		"done": 0, "success": 0, "failed": 0,
	}).Error

	go s.runCrawlJob(ctx, tenantID, id)
	return nil
}

func (s *CatalogService) runCrawlJob(ctx context.Context, tenantID, jobID uint64) {
	defer func() {
		s.mu.Lock()
		delete(s.running, jobID)
		s.mu.Unlock()
	}()

	var job model.CrawlJob
	if err := s.repos.DB.First(&job, jobID).Error; err != nil {
		return
	}
	adapter := crawl.LookupAdapter(job.Adapter)
	var items []model.CrawlItem
	_ = s.repos.DB.Where("job_id = ?", jobID).Order("id asc").Find(&items).Error

	success, failed := 0, 0
	for i, it := range items {
		select {
		case <-ctx.Done():
			fin := time.Now()
			_ = s.repos.DB.Model(&model.CrawlJob{}).Where("id = ?", jobID).Updates(map[string]any{
				"status": model.CrawlJobCancelled, "finished_at": fin, "message": "cancelled",
				"done": i, "success": success, "failed": failed,
			}).Error
			return
		default:
		}
		// 礼貌间隔，避免打爆目标站
		if i > 0 {
			time.Sleep(800 * time.Millisecond)
		}
		body, err := s.fetch.Get(ctx, it.SourceURL)
		if err != nil {
			failed++
			_ = s.repos.DB.Model(&it).Updates(map[string]any{"status": "failed", "error_msg": err.Error()}).Error
			_ = s.repos.DB.Model(&model.CrawlJob{}).Where("id = ?", jobID).Updates(map[string]any{
				"done": i + 1, "failed": failed,
			}).Error
			continue
		}
		parsed, err := adapter.Parse(ctx, it.SourceURL, body)
		if err != nil {
			failed++
			_ = s.repos.DB.Model(&it).Updates(map[string]any{"status": "failed", "error_msg": err.Error()}).Error
			_ = s.repos.DB.Model(&model.CrawlJob{}).Where("id = ?", jobID).Updates(map[string]any{
				"done": i + 1, "failed": failed,
			}).Error
			continue
		}
		payload, _ := json.Marshal(parsed)
		productID, impErr := s.importParsedProduct(tenantID, &job, it.SourceURL, parsed)
		st := "ok"
		errMsg := ""
		if impErr != nil {
			st = "failed"
			errMsg = impErr.Error()
			failed++
			log.Printf("catalog crawl import %s: %v", it.SourceURL, impErr)
		} else {
			success++
		}
		_ = s.repos.DB.Model(&it).Updates(map[string]any{
			"status": st, "title": parsed.Title, "payload": string(payload),
			"product_id": productID, "error_msg": errMsg,
		}).Error
		_ = s.repos.DB.Model(&model.CrawlJob{}).Where("id = ?", jobID).Updates(map[string]any{
			"done": i + 1, "success": success, "failed": failed,
		}).Error
	}
	fin := time.Now()
	status := model.CrawlJobSucceeded
	msg := ""
	if failed > 0 && success == 0 {
		status = model.CrawlJobFailed
		msg = "all items failed"
	} else if failed > 0 {
		msg = fmt.Sprintf("%d failed", failed)
	}
	_ = s.repos.DB.Model(&model.CrawlJob{}).Where("id = ?", jobID).Updates(map[string]any{
		"status": status, "finished_at": fin, "message": msg,
		"done": len(items), "success": success, "failed": failed,
	}).Error
}

func (s *CatalogService) importParsedProduct(tenantID uint64, job *model.CrawlJob, sourceURL string, p *crawl.ParsedProduct) (uint64, error) {
	brandID := job.BrandID
	if brandID == 0 {
		return 0, fmt.Errorf("请先为爬取任务指定品牌")
	}
	specsJSON := ""
	if len(p.Specs) > 0 {
		b, _ := json.Marshal(p.Specs)
		specsJSON = string(b)
	}
	host := ""
	if u := strings.TrimSpace(sourceURL); u != "" {
		if i := strings.Index(u, "://"); i >= 0 {
			rest := u[i+3:]
			if j := strings.Index(rest, "/"); j >= 0 {
				host = rest[:j]
			} else {
				host = rest
			}
		}
	}
	row := model.Product{
		TenantID:   repo.NormalizeTenantID(tenantID),
		BrandID:    brandID,
		SeriesID:   job.SeriesID,
		Name:       strings.TrimSpace(p.Title),
		ModelNo:    strings.TrimSpace(p.SKU),
		CoverURL:   strings.TrimSpace(p.Image),
		Intro:      strings.TrimSpace(p.Description),
		SpecsJSON:  specsJSON,
		SourceURL:  sourceURL,
		SourceSite: host,
		CrawlJobID: job.ID,
		Status:     1,
	}
	if err := s.repos.DB.Create(&row).Error; err != nil {
		return 0, err
	}
	for i, v := range p.Variants {
		name := strings.TrimSpace(v.Name)
		if name == "" {
			continue
		}
		sj := ""
		if len(v.Specs) > 0 {
			b, _ := json.Marshal(v.Specs)
			sj = string(b)
		}
		_ = s.repos.DB.Create(&model.Variant{
			TenantID:  repo.NormalizeTenantID(tenantID),
			ProductID: row.ID,
			Name:      name,
			SKUCode:   strings.TrimSpace(v.SKU),
			SpecJSON:  sj,
			PriceHint: strings.TrimSpace(v.Price),
			PicURL:    strings.TrimSpace(v.Image),
			Sort:      i,
			Status:    1,
		}).Error
	}
	// 若无 variants 但有 SKU，建一条默认子规格
	if len(p.Variants) == 0 && (p.SKU != "" || p.Offers != "") {
		_ = s.repos.DB.Create(&model.Variant{
			TenantID:  repo.NormalizeTenantID(tenantID),
			ProductID: row.ID,
			Name:      "默认规格",
			SKUCode:   strings.TrimSpace(p.SKU),
			PriceHint: strings.TrimSpace(p.Offers),
			PicURL:    strings.TrimSpace(p.Image),
			Status:    1,
		}).Error
	}
	return row.ID, nil
}

func toCrawlJobDTO(j *model.CrawlJob) dto.CrawlJobDTO {
	d := dto.CrawlJobDTO{
		ID: j.ID, Name: j.Name, Adapter: j.Adapter, BrandID: j.BrandID, SeriesID: j.SeriesID,
		SeedURLs: j.SeedURLs, Status: j.Status, Total: j.Total, Done: j.Done,
		Success: j.Success, Failed: j.Failed, Message: j.Message,
		CreatedAt: j.CreatedAt.Format(time.RFC3339),
	}
	if j.StartedAt != nil {
		d.StartedAt = j.StartedAt.Format(time.RFC3339)
	}
	if j.FinishedAt != nil {
		d.FinishedAt = j.FinishedAt.Format(time.RFC3339)
	}
	return d
}

func (s *CatalogService) DashboardStats(tenantID uint64) (*dto.DashboardStats, error) {
	stats := &dto.DashboardStats{}
	s.db(tenantID).Model(&model.Brand{}).Count(&stats.BrandCount)
	s.db(tenantID).Model(&model.Series{}).Count(&stats.SeriesCount)
	s.db(tenantID).Model(&model.Product{}).Count(&stats.ProductCount)
	s.db(tenantID).Model(&model.Variant{}).Count(&stats.VariantCount)
	s.db(tenantID).Model(&model.CrawlJob{}).Where("status = ?", model.CrawlJobRunning).Count(&stats.CrawlRunning)
	return stats, nil
}
