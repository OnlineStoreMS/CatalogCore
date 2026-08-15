package admin

import (
	"errors"
	"net/http"
	"strconv"

	"catalogcore/internal/dto"
	"catalogcore/internal/pkg/authcontext"
	"catalogcore/internal/pkg/httputil"
	"catalogcore/internal/pkg/response"
	"catalogcore/internal/service"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	svc *service.CatalogService
}

func NewHandlers(svc *service.CatalogService) *Handlers {
	return &Handlers{svc: svc}
}

func (h *Handlers) mapError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		response.Fail(c, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrBadRequest):
		response.Fail(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrConflict):
		response.Fail(c, http.StatusConflict, err.Error())
	default:
		response.Fail(c, http.StatusInternalServerError, err.Error())
	}
}

func (h *Handlers) DashboardStats(c *gin.Context) {
	stats, err := h.svc.DashboardStats(authcontext.TenantID(c))
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.OK(c, stats)
}

func (h *Handlers) ListBrands(c *gin.Context) {
	list, err := h.svc.ListBrands(authcontext.TenantID(c), c.Query("keyword"))
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.OK(c, list)
}

func (h *Handlers) CreateBrand(c *gin.Context) {
	var req dto.BrandSaveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.svc.SaveBrand(authcontext.TenantID(c), 0, req)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *Handlers) UpdateBrand(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var req dto.BrandSaveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.svc.SaveBrand(authcontext.TenantID(c), id, req)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *Handlers) DeleteBrand(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.DeleteBrand(authcontext.TenantID(c), id); err != nil {
		h.mapError(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *Handlers) ListSeries(c *gin.Context) {
	brandID, _ := strconv.ParseUint(c.Query("brandId"), 10, 64)
	list, err := h.svc.ListSeries(authcontext.TenantID(c), brandID)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.OK(c, list)
}

func (h *Handlers) CreateSeries(c *gin.Context) {
	var req dto.SeriesSaveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.svc.SaveSeries(authcontext.TenantID(c), 0, req)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *Handlers) UpdateSeries(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var req dto.SeriesSaveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.svc.SaveSeries(authcontext.TenantID(c), id, req)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *Handlers) DeleteSeries(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.DeleteSeries(authcontext.TenantID(c), id); err != nil {
		h.mapError(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *Handlers) ListProducts(c *gin.Context) {
	var q dto.ProductListQuery
	_ = c.ShouldBindQuery(&q)
	list, total, err := h.svc.ListProducts(authcontext.TenantID(c), q)
	if err != nil {
		h.mapError(c, err)
		return
	}
	page, pageSize := q.Page, q.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	response.OK(c, response.PageResult(list, total, page, pageSize))
}

func (h *Handlers) GetProduct(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	item, err := h.svc.GetProduct(authcontext.TenantID(c), id)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *Handlers) CreateProduct(c *gin.Context) {
	var req dto.ProductSaveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.svc.SaveProduct(authcontext.TenantID(c), authcontext.UserID(c), 0, req)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *Handlers) UpdateProduct(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var req dto.ProductSaveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.svc.SaveProduct(authcontext.TenantID(c), authcontext.UserID(c), id, req)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *Handlers) DeleteProduct(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.DeleteProduct(authcontext.TenantID(c), id); err != nil {
		h.mapError(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *Handlers) ListCrawlJobs(c *gin.Context) {
	page, pageSize := httputil.ParsePage(c)
	list, total, err := h.svc.ListCrawlJobs(authcontext.TenantID(c), page, pageSize)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.OK(c, response.PageResult(list, total, page, pageSize))
}

func (h *Handlers) GetCrawlJob(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	job, items, err := h.svc.GetCrawlJob(authcontext.TenantID(c), id)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.OK(c, gin.H{"job": job, "items": items})
}

func (h *Handlers) CreateCrawlJob(c *gin.Context) {
	var req dto.CrawlJobCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.svc.CreateCrawlJob(authcontext.TenantID(c), authcontext.UserID(c), req)
	if err != nil {
		h.mapError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *Handlers) StartCrawlJob(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.StartCrawlJob(authcontext.TenantID(c), id); err != nil {
		h.mapError(c, err)
		return
	}
	response.OK(c, gin.H{"ok": true})
}
