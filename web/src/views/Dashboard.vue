<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Collection, Goods, Download } from '@element-plus/icons-vue'
import { fetchDashboardStats, type DashboardStats } from '../api/catalog'

const router = useRouter()
const stats = ref<DashboardStats>({
  brandCount: 0,
  seriesCount: 0,
  productCount: 0,
  variantCount: 0,
  crawlRunning: 0,
})
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    stats.value = await fetchDashboardStats()
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div v-loading="loading" class="dashboard">
    <h2 class="page-title">产品中心</h2>
    <p class="desc">搜集整理各品牌产品知识：系列、规格、介绍与说明书。资料主要来自公开页面爬取，暂不与商品中心联动。</p>

    <div class="stat-row">
      <el-card shadow="never" class="stat"><div class="n">{{ stats.brandCount }}</div><div class="l">品牌</div></el-card>
      <el-card shadow="never" class="stat"><div class="n">{{ stats.seriesCount }}</div><div class="l">系列</div></el-card>
      <el-card shadow="never" class="stat"><div class="n">{{ stats.productCount }}</div><div class="l">产品</div></el-card>
      <el-card shadow="never" class="stat"><div class="n">{{ stats.variantCount }}</div><div class="l">子规格</div></el-card>
      <el-card shadow="never" class="stat"><div class="n">{{ stats.crawlRunning }}</div><div class="l">爬取中</div></el-card>
    </div>

    <div class="card-grid">
      <el-card shadow="hover" class="action-card" @click="router.push('/brands')">
        <el-icon :size="32" color="#409eff"><Collection /></el-icon>
        <h3>品牌与系列</h3>
        <p>维护品牌、产品线</p>
      </el-card>
      <el-card shadow="hover" class="action-card" @click="router.push('/products')">
        <el-icon :size="32" color="#67c23a"><Goods /></el-icon>
        <h3>产品库</h3>
        <p>产品条目与子规格</p>
      </el-card>
      <el-card shadow="hover" class="action-card" @click="router.push('/crawl')">
        <el-icon :size="32" color="#e6a23c"><Download /></el-icon>
        <h3>爬取任务</h3>
        <p>从公开页面采集产品信息</p>
      </el-card>
    </div>
  </div>
</template>

<style scoped>
.page-title { margin: 0 0 8px; font-size: 20px; }
.desc { color: #909399; margin: 0 0 20px; }
.stat-row { display: flex; gap: 12px; flex-wrap: wrap; margin-bottom: 20px; }
.stat { min-width: 120px; text-align: center; }
.stat .n { font-size: 28px; font-weight: 600; color: #303133; }
.stat .l { color: #909399; font-size: 13px; }
.card-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 16px; }
.action-card { cursor: pointer; text-align: center; padding: 12px 0; }
.action-card h3 { margin: 12px 0 6px; font-size: 16px; }
.action-card p { margin: 0; color: #909399; font-size: 13px; }
</style>
