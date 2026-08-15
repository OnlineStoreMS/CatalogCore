import { createRouter, createWebHistory } from 'vue-router'
import AdminLayout from '../layouts/AdminLayout.vue'
import { redirectToPortal, ensureSession, clearToken } from '../utils/auth'

const APP_TITLE = 'CatalogCore - 产品中心'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/auth/callback',
      name: 'AuthCallback',
      component: () => import('../views/AuthCallback.vue'),
      meta: { public: true },
    },
    {
      path: '/auth/logout',
      name: 'AuthLogout',
      component: () => import('../views/AuthLogout.vue'),
      meta: { public: true },
    },
    {
      path: '/',
      component: AdminLayout,
      redirect: '/dashboard',
      children: [
        { path: 'dashboard', name: 'Dashboard', component: () => import('../views/Dashboard.vue'), meta: { title: '工作台' } },
        { path: 'brands', name: 'Brands', component: () => import('../views/Brands.vue'), meta: { title: '品牌与系列' } },
        { path: 'products', name: 'Products', component: () => import('../views/Products.vue'), meta: { title: '产品库' } },
        { path: 'products/new', name: 'ProductCreate', component: () => import('../views/ProductEdit.vue'), meta: { title: '新建产品' } },
        { path: 'products/:id', name: 'ProductEdit', component: () => import('../views/ProductEdit.vue'), meta: { title: '编辑产品' } },
        { path: 'crawl', name: 'CrawlJobs', component: () => import('../views/CrawlJobs.vue'), meta: { title: '爬取任务' } },
      ],
    },
  ],
})

router.beforeEach(async (to) => {
  if (to.meta.public) return true
  const ok = await ensureSession()
  if (!ok) {
    clearToken()
    redirectToPortal()
    return false
  }
  return true
})

router.afterEach((to) => {
  const page = to.meta.title as string | undefined
  document.title = page ? `${page} - ${APP_TITLE}` : APP_TITLE
})

export default router
