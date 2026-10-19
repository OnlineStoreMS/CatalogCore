import client, { unwrap, type PageData } from './client'

export interface DashboardStats {
  brandCount: number
  seriesCount: number
  productCount: number
  variantCount: number
  crawlRunning: number
}

export interface Brand {
  id: number
  name: string
  nameEn: string
  code: string
  logoUrl: string
  website: string
  country: string
  description: string
  sort: number
  status: number
}

export interface Category {
  id: number
  name: string
  nameEn: string
  code: string
  sort: number
  status: number
}

export interface Series {
  id: number
  brandId: number
  categoryId: number
  categoryName?: string
  categoryNameEn?: string
  categoryCode?: string
  name: string
  code: string
  description: string
  coverUrl: string
  sort: number
  status: number
}

export interface Variant {
  id?: number
  name: string
  skuCode?: string
  specJson?: string
  priceHint?: string
  picUrl?: string
  sort?: number
  status?: number
}

export interface Product {
  id: number
  brandId: number
  seriesId: number
  brandName?: string
  seriesName?: string
  name: string
  modelNo: string
  subtitle: string
  coverUrl: string
  intro: string
  detailHtml?: string
  specsJson: string
  tagsJson: string
  sourceUrl: string
  sourceSite: string
  status: number
  variants?: Variant[]
  attachments?: { id?: number; kind: string; name: string; url: string }[]
}

export interface CrawlJob {
  id: number
  name: string
  adapter: string
  brandId: number
  seriesId: number
  seedUrls: string
  status: string
  total: number
  done: number
  success: number
  failed: number
  message: string
}

export async function fetchDashboardStats() {
  return unwrap<DashboardStats>(await client.get('/dashboard/stats'))
}

export async function listBrands(keyword = '') {
  return unwrap<Brand[]>(await client.get('/brands', { params: { keyword } }))
}

export async function saveBrand(data: Partial<Brand> & { name: string }, id?: number) {
  if (id) return unwrap<Brand>(await client.put(`/brands/${id}`, data))
  return unwrap<Brand>(await client.post('/brands', data))
}

export async function deleteBrand(id: number) {
  return unwrap(await client.delete(`/brands/${id}`))
}

export async function listCategories() {
  return unwrap<Category[]>(await client.get('/categories'))
}

export function seriesOptionLabel(series: Series) {
  return series.categoryName ? `${series.categoryName} / ${series.name}` : series.name
}

export async function listSeries(brandId?: number) {
  return unwrap<Series[]>(await client.get('/series', { params: { brandId: brandId || undefined } }))
}

export async function saveSeries(data: Partial<Series> & { name: string; brandId: number }, id?: number) {
  if (id) return unwrap<Series>(await client.put(`/series/${id}`, data))
  return unwrap<Series>(await client.post('/series', data))
}

export async function deleteSeries(id: number) {
  return unwrap(await client.delete(`/series/${id}`))
}

export async function listProducts(params: Record<string, unknown>) {
  return unwrap<PageData<Product>>(await client.get('/products', { params }))
}

export async function getProduct(id: number) {
  return unwrap<Product>(await client.get(`/products/${id}`))
}

export async function saveProduct(data: Record<string, unknown>, id?: number) {
  if (id) return unwrap<Product>(await client.put(`/products/${id}`, data))
  return unwrap<Product>(await client.post('/products', data))
}

export async function deleteProduct(id: number) {
  return unwrap(await client.delete(`/products/${id}`))
}

export async function listCrawlJobs(page = 1, pageSize = 20) {
  return unwrap<PageData<CrawlJob>>(await client.get('/crawl-jobs', { params: { page, pageSize } }))
}

export async function getCrawlJob(id: number) {
  return unwrap<{ job: CrawlJob; items: { id: number; sourceUrl: string; status: string; title: string; productId: number; errorMsg: string }[] }>(
    await client.get(`/crawl-jobs/${id}`),
  )
}

export async function createCrawlJob(data: { name: string; adapter?: string; brandId: number; seriesId?: number; seedUrls: string }) {
  return unwrap<CrawlJob>(await client.post('/crawl-jobs', data))
}

export async function startCrawlJob(id: number) {
  return unwrap(await client.post(`/crawl-jobs/${id}/start`))
}
