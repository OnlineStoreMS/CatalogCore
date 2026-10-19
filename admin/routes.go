package admin

import "github.com/gin-gonic/gin"

func RegisterRoutes(g *gin.RouterGroup, h *Handlers) {
	g.GET("/dashboard/stats", h.DashboardStats)

	g.GET("/brands", h.ListBrands)
	g.POST("/brands", h.CreateBrand)
	g.PUT("/brands/:id", h.UpdateBrand)
	g.DELETE("/brands/:id", h.DeleteBrand)

	g.GET("/categories", h.ListCategories)

	g.GET("/series", h.ListSeries)
	g.POST("/series", h.CreateSeries)
	g.PUT("/series/:id", h.UpdateSeries)
	g.DELETE("/series/:id", h.DeleteSeries)

	g.GET("/products", h.ListProducts)
	g.POST("/products", h.CreateProduct)
	g.GET("/products/:id", h.GetProduct)
	g.PUT("/products/:id", h.UpdateProduct)
	g.DELETE("/products/:id", h.DeleteProduct)

	g.GET("/crawl-jobs", h.ListCrawlJobs)
	g.POST("/crawl-jobs", h.CreateCrawlJob)
	g.GET("/crawl-jobs/:id", h.GetCrawlJob)
	g.POST("/crawl-jobs/:id/start", h.StartCrawlJob)
}
