package model

import (
	"time"

	"gorm.io/gorm"
)

// Brand 品牌
type Brand struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	TenantID    uint64         `gorm:"index;not null;default:1" json:"tenantId"`
	Name        string         `gorm:"size:128;not null" json:"name"`
	NameEn      string         `gorm:"size:128" json:"nameEn"`
	Code        string         `gorm:"size:64;index" json:"code"`
	LogoURL     string         `gorm:"size:1024" json:"logoUrl"`
	Website     string         `gorm:"size:512" json:"website"`
	Country     string         `gorm:"size:64" json:"country"`
	Description string         `gorm:"type:text" json:"description"`
	Sort        int            `gorm:"default:0" json:"sort"`
	Status      int8           `gorm:"default:1" json:"status"` // 1启用 0停用
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Brand) TableName() string { return "catalog_brands" }

// Category 品类。平台标准的一级分类，不隶属于品牌。
// 骑行场景（公路、山地）和产品域（服饰）共用这一级，靠稳定 code 对齐。
type Category struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	TenantID  uint64         `gorm:"index;not null;default:1" json:"tenantId"`
	Name      string         `gorm:"size:128;not null" json:"name"`
	NameEn    string         `gorm:"size:128" json:"nameEn"`
	Code      string         `gorm:"size:64;index;not null" json:"code"`
	Sort      int            `gorm:"default:0" json:"sort"`
	Status    int8           `gorm:"default:1" json:"status"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Category) TableName() string { return "catalog_categories" }

// Series 产品系列 / 产品线
type Series struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	TenantID    uint64         `gorm:"index;not null;default:1" json:"tenantId"`
	BrandID     uint64         `gorm:"index;not null" json:"brandId"`
	CategoryID  uint64         `gorm:"index;default:0" json:"categoryId"`
	Name        string         `gorm:"size:128;not null" json:"name"`
	Code        string         `gorm:"size:64" json:"code"`
	Description string         `gorm:"type:text" json:"description"`
	CoverURL    string         `gorm:"size:1024" json:"coverUrl"`
	Sort        int            `gorm:"default:0" json:"sort"`
	Status      int8           `gorm:"default:1" json:"status"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Series) TableName() string { return "catalog_series" }

// Product 产品知识条目（非电商可售 SKU）
type Product struct {
	ID         uint64         `gorm:"primaryKey" json:"id"`
	TenantID   uint64         `gorm:"index;not null;default:1" json:"tenantId"`
	BrandID    uint64         `gorm:"index;not null" json:"brandId"`
	SeriesID   uint64         `gorm:"index;default:0" json:"seriesId"`
	Name       string         `gorm:"size:256;not null" json:"name"`
	ModelNo    string         `gorm:"size:128;index" json:"modelNo"` // 型号
	Subtitle   string         `gorm:"size:512" json:"subtitle"`
	CoverURL   string         `gorm:"size:1024" json:"coverUrl"`
	Intro      string         `gorm:"type:text" json:"intro"`
	DetailHTML string         `gorm:"type:text" json:"detailHtml"`
	SpecsJSON  string         `gorm:"type:text" json:"specsJson"` // 通用规格 [{k,v}]
	TagsJSON   string         `gorm:"type:text" json:"tagsJson"`
	SourceURL  string         `gorm:"size:1024" json:"sourceUrl"` // 来源页
	SourceSite string         `gorm:"size:128" json:"sourceSite"`
	CrawlJobID uint64         `gorm:"index;default:0" json:"crawlJobId"`
	Status     int8           `gorm:"default:1" json:"status"` // 1草稿 2已发布 0废弃
	CreatedBy  uint64         `json:"createdBy"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	Variants    []Variant    `gorm:"foreignKey:ProductID" json:"variants,omitempty"`
	Attachments []Attachment `gorm:"foreignKey:ProductID" json:"attachments,omitempty"`
}

func (Product) TableName() string { return "catalog_products" }

// Variant 子规格 / 变体
type Variant struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	TenantID  uint64         `gorm:"index;not null;default:1" json:"tenantId"`
	ProductID uint64         `gorm:"index;not null" json:"productId"`
	Name      string         `gorm:"size:256;not null" json:"name"`
	SKUCode   string         `gorm:"size:128;index" json:"skuCode"`
	SpecJSON  string         `gorm:"type:text" json:"specJson"` // {color,size,...}
	PriceHint string         `gorm:"size:64" json:"priceHint"`  // 参考价文本
	PicURL    string         `gorm:"size:1024" json:"picUrl"`
	Sort      int            `gorm:"default:0" json:"sort"`
	Status    int8           `gorm:"default:1" json:"status"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Variant) TableName() string { return "catalog_variants" }

// Attachment 说明书 / 图片等
type Attachment struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	TenantID  uint64         `gorm:"index;not null;default:1" json:"tenantId"`
	ProductID uint64         `gorm:"index;not null" json:"productId"`
	Kind      string         `gorm:"size:32;not null" json:"kind"` // image|manual|other
	Name      string         `gorm:"size:256" json:"name"`
	URL       string         `gorm:"size:1024;not null" json:"url"`
	Mime      string         `gorm:"size:128" json:"mime"`
	SizeBytes int64          `json:"sizeBytes"`
	Sort      int            `gorm:"default:0" json:"sort"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Attachment) TableName() string { return "catalog_attachments" }

const (
	CrawlJobPending   = "pending"
	CrawlJobRunning   = "running"
	CrawlJobSucceeded = "succeeded"
	CrawlJobFailed    = "failed"
	CrawlJobCancelled = "cancelled"
)

// CrawlJob 爬取任务（公开页面采集；不用于绕过登录/验证码）
type CrawlJob struct {
	ID         uint64         `gorm:"primaryKey" json:"id"`
	TenantID   uint64         `gorm:"index;not null;default:1" json:"tenantId"`
	Name       string         `gorm:"size:128;not null" json:"name"`
	Adapter    string         `gorm:"size:64;not null;default:generic_product" json:"adapter"`
	BrandID    uint64         `gorm:"index;default:0" json:"brandId"`
	SeriesID   uint64         `gorm:"index;default:0" json:"seriesId"`
	SeedURLs   string         `gorm:"type:text" json:"seedUrls"` // 换行分隔
	Status     string         `gorm:"size:32;index;default:pending" json:"status"`
	Total      int            `gorm:"default:0" json:"total"`
	Done       int            `gorm:"default:0" json:"done"`
	Success    int            `gorm:"default:0" json:"success"`
	Failed     int            `gorm:"default:0" json:"failed"`
	Message    string         `gorm:"type:text" json:"message"`
	CreatedBy  uint64         `json:"createdBy"`
	StartedAt  *time.Time     `json:"startedAt"`
	FinishedAt *time.Time     `json:"finishedAt"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	Items []CrawlItem `gorm:"foreignKey:JobID" json:"items,omitempty"`
}

func (CrawlJob) TableName() string { return "catalog_crawl_jobs" }

// CrawlItem 单页采集结果
type CrawlItem struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	TenantID  uint64         `gorm:"index;not null;default:1" json:"tenantId"`
	JobID     uint64         `gorm:"index;not null" json:"jobId"`
	SourceURL string         `gorm:"size:1024;not null" json:"sourceUrl"`
	Status    string         `gorm:"size:32;index" json:"status"` // pending|ok|failed|imported
	Title     string         `gorm:"size:512" json:"title"`
	Payload   string         `gorm:"type:text" json:"payload"` // JSON 原始解析
	ProductID uint64         `gorm:"index;default:0" json:"productId"`
	ErrorMsg  string         `gorm:"type:text" json:"errorMsg"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CrawlItem) TableName() string { return "catalog_crawl_items" }
