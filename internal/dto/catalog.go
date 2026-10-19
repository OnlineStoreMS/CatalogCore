package dto

type PageQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"pageSize"`
}

type BrandDTO struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	NameEn      string `json:"nameEn"`
	Code        string `json:"code"`
	LogoURL     string `json:"logoUrl"`
	Website     string `json:"website"`
	Country     string `json:"country"`
	Description string `json:"description"`
	Sort        int    `json:"sort"`
	Status      int8   `json:"status"`
	CreatedAt   string `json:"createdAt,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
}

type BrandSaveReq struct {
	Name        string `json:"name" binding:"required"`
	NameEn      string `json:"nameEn"`
	Code        string `json:"code"`
	LogoURL     string `json:"logoUrl"`
	Website     string `json:"website"`
	Country     string `json:"country"`
	Description string `json:"description"`
	Sort        int    `json:"sort"`
	Status      *int8  `json:"status"`
}

type CategoryDTO struct {
	ID     uint64 `json:"id"`
	Name   string `json:"name"`
	NameEn string `json:"nameEn"`
	Code   string `json:"code"`
	Sort   int    `json:"sort"`
	Status int8   `json:"status"`
}

type SeriesDTO struct {
	ID             uint64 `json:"id"`
	BrandID        uint64 `json:"brandId"`
	CategoryID     uint64 `json:"categoryId"`
	CategoryName   string `json:"categoryName,omitempty"`
	CategoryNameEn string `json:"categoryNameEn,omitempty"`
	CategoryCode   string `json:"categoryCode,omitempty"`
	Name           string `json:"name"`
	Code           string `json:"code"`
	Description    string `json:"description"`
	CoverURL       string `json:"coverUrl"`
	Sort           int    `json:"sort"`
	Status         int8   `json:"status"`
}

type SeriesSaveReq struct {
	BrandID     uint64 `json:"brandId" binding:"required"`
	CategoryID  uint64 `json:"categoryId"`
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code"`
	Description string `json:"description"`
	CoverURL    string `json:"coverUrl"`
	Sort        int    `json:"sort"`
	Status      *int8  `json:"status"`
}

type VariantDTO struct {
	ID        uint64 `json:"id,omitempty"`
	Name      string `json:"name"`
	SKUCode   string `json:"skuCode"`
	SpecJSON  string `json:"specJson"`
	PriceHint string `json:"priceHint"`
	PicURL    string `json:"picUrl"`
	Sort      int    `json:"sort"`
	Status    int8   `json:"status"`
}

type AttachmentDTO struct {
	ID        uint64 `json:"id,omitempty"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	Mime      string `json:"mime"`
	SizeBytes int64  `json:"sizeBytes"`
	Sort      int    `json:"sort"`
}

type ProductDTO struct {
	ID          uint64          `json:"id"`
	BrandID     uint64          `json:"brandId"`
	SeriesID    uint64          `json:"seriesId"`
	BrandName   string          `json:"brandName,omitempty"`
	SeriesName  string          `json:"seriesName,omitempty"`
	Name        string          `json:"name"`
	ModelNo     string          `json:"modelNo"`
	Subtitle    string          `json:"subtitle"`
	CoverURL    string          `json:"coverUrl"`
	Intro       string          `json:"intro"`
	DetailHTML  string          `json:"detailHtml"`
	SpecsJSON   string          `json:"specsJson"`
	TagsJSON    string          `json:"tagsJson"`
	SourceURL   string          `json:"sourceUrl"`
	SourceSite  string          `json:"sourceSite"`
	CrawlJobID  uint64          `json:"crawlJobId"`
	Status      int8            `json:"status"`
	Variants    []VariantDTO    `json:"variants,omitempty"`
	Attachments []AttachmentDTO `json:"attachments,omitempty"`
	CreatedAt   string          `json:"createdAt,omitempty"`
	UpdatedAt   string          `json:"updatedAt,omitempty"`
}

type ProductSaveReq struct {
	BrandID     uint64          `json:"brandId" binding:"required"`
	SeriesID    uint64          `json:"seriesId"`
	Name        string          `json:"name" binding:"required"`
	ModelNo     string          `json:"modelNo"`
	Subtitle    string          `json:"subtitle"`
	CoverURL    string          `json:"coverUrl"`
	Intro       string          `json:"intro"`
	DetailHTML  string          `json:"detailHtml"`
	SpecsJSON   string          `json:"specsJson"`
	TagsJSON    string          `json:"tagsJson"`
	SourceURL   string          `json:"sourceUrl"`
	SourceSite  string          `json:"sourceSite"`
	Status      *int8           `json:"status"`
	Variants    []VariantDTO    `json:"variants"`
	Attachments []AttachmentDTO `json:"attachments"`
}

type ProductListQuery struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Keyword  string `form:"keyword"`
	BrandID  uint64 `form:"brandId"`
	SeriesID uint64 `form:"seriesId"`
	Status   *int8  `form:"status"`
}

type CrawlJobCreateReq struct {
	Name     string `json:"name" binding:"required"`
	Adapter  string `json:"adapter"`
	BrandID  uint64 `json:"brandId"`
	SeriesID uint64 `json:"seriesId"`
	SeedURLs string `json:"seedUrls" binding:"required"` // 换行或逗号分隔
}

type CrawlJobDTO struct {
	ID         uint64 `json:"id"`
	Name       string `json:"name"`
	Adapter    string `json:"adapter"`
	BrandID    uint64 `json:"brandId"`
	SeriesID   uint64 `json:"seriesId"`
	SeedURLs   string `json:"seedUrls"`
	Status     string `json:"status"`
	Total      int    `json:"total"`
	Done       int    `json:"done"`
	Success    int    `json:"success"`
	Failed     int    `json:"failed"`
	Message    string `json:"message"`
	StartedAt  string `json:"startedAt,omitempty"`
	FinishedAt string `json:"finishedAt,omitempty"`
	CreatedAt  string `json:"createdAt,omitempty"`
}

type CrawlItemDTO struct {
	ID        uint64 `json:"id"`
	JobID     uint64 `json:"jobId"`
	SourceURL string `json:"sourceUrl"`
	Status    string `json:"status"`
	Title     string `json:"title"`
	ProductID uint64 `json:"productId"`
	ErrorMsg  string `json:"errorMsg"`
}

type DashboardStats struct {
	BrandCount   int64 `json:"brandCount"`
	SeriesCount  int64 `json:"seriesCount"`
	ProductCount int64 `json:"productCount"`
	VariantCount int64 `json:"variantCount"`
	CrawlRunning int64 `json:"crawlRunning"`
}
