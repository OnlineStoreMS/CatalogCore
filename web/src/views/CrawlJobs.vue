<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  listCrawlJobs, createCrawlJob, startCrawlJob, getCrawlJob, listBrands, listSeries,
  type CrawlJob, type Brand, type Series,
} from '../api/catalog'

const loading = ref(false)
const jobs = ref<CrawlJob[]>([])
const total = ref(0)
const page = ref(1)
const brands = ref<Brand[]>([])
const series = ref<Series[]>([])

const dialog = ref(false)
const form = ref({ name: '', brandId: 0, seriesId: 0, seedUrls: '', adapter: 'generic_product' })

const detailVisible = ref(false)
const detail = ref<{ job: CrawlJob; items: { id: number; sourceUrl: string; status: string; title: string; productId: number; errorMsg: string }[] } | null>(null)

async function load() {
  loading.value = true
  try {
    const res = await listCrawlJobs(page.value, 20)
    jobs.value = res.list || []
    total.value = res.total || 0
  } catch (e) {
    ElMessage.error((e as Error).message)
  } finally {
    loading.value = false
  }
}

async function openCreate() {
  brands.value = await listBrands()
  form.value = { name: '', brandId: brands.value[0]?.id || 0, seriesId: 0, seedUrls: '', adapter: 'generic_product' }
  await onBrandChange()
  dialog.value = true
}

async function onBrandChange() {
  series.value = form.value.brandId ? await listSeries(form.value.brandId) : []
  form.value.seriesId = 0
}

async function submit() {
  if (!form.value.brandId || !form.value.seedUrls.trim()) {
    ElMessage.warning('请选择品牌并填写种子 URL（每行一个）')
    return
  }
  try {
    await createCrawlJob({
      name: form.value.name || '爬取任务',
      brandId: form.value.brandId,
      seriesId: form.value.seriesId || undefined,
      seedUrls: form.value.seedUrls,
      adapter: form.value.adapter,
    })
    dialog.value = false
    ElMessage.success('已创建')
    await load()
  } catch (e) {
    ElMessage.error((e as Error).message)
  }
}

async function start(row: CrawlJob) {
  try {
    await startCrawlJob(row.id)
    ElMessage.success('已启动')
    await load()
  } catch (e) {
    ElMessage.error((e as Error).message)
  }
}

async function showDetail(row: CrawlJob) {
  try {
    detail.value = await getCrawlJob(row.id)
    detailVisible.value = true
  } catch (e) {
    ElMessage.error((e as Error).message)
  }
}

onMounted(load)
</script>

<template>
  <div>
    <div class="page-head">
      <div>
        <h2>爬取任务</h2>
        <p class="hint">从公开产品页采集标题、简介、图片与 JSON-LD 规格；遵守站点 robots，不绕过登录/验证码。</p>
      </div>
      <el-button type="primary" @click="openCreate">新建任务</el-button>
    </div>

    <el-table v-loading="loading" :data="jobs" stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="name" label="名称" min-width="140" />
      <el-table-column prop="adapter" label="适配器" width="140" />
      <el-table-column prop="status" label="状态" width="100" />
      <el-table-column label="进度" width="160">
        <template #default="{ row }">{{ row.done }}/{{ row.total }}（成功 {{ row.success }} / 失败 {{ row.failed }}）</template>
      </el-table-column>
      <el-table-column prop="message" label="备注" min-width="160" show-overflow-tooltip />
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" :disabled="row.status === 'running'" @click="start(row)">启动</el-button>
          <el-button link @click="showDetail(row)">明细</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pager">
      <el-pagination v-model:current-page="page" :total="total" layout="total, prev, pager, next" @current-change="load" />
    </div>

    <el-dialog v-model="dialog" title="新建爬取任务" width="640px">
      <el-form label-width="90px">
        <el-form-item label="名称"><el-input v-model="form.name" placeholder="可选" /></el-form-item>
        <el-form-item label="品牌" required>
          <el-select v-model="form.brandId" filterable style="width: 100%" @change="onBrandChange">
            <el-option v-for="b in brands" :key="b.id" :label="b.name" :value="b.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="系列">
          <el-select v-model="form.seriesId" clearable filterable style="width: 100%">
            <el-option v-for="s in series" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="适配器">
          <el-select v-model="form.adapter" style="width: 100%">
            <el-option label="通用产品页 (JSON-LD/OG)" value="generic_product" />
          </el-select>
        </el-form-item>
        <el-form-item label="种子 URL" required>
          <el-input v-model="form.seedUrls" type="textarea" :rows="8" placeholder="每行一个公开产品页 URL" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog = false">取消</el-button>
        <el-button type="primary" @click="submit">创建</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="detailVisible" title="任务明细" size="50%">
      <template v-if="detail">
        <p>状态：{{ detail.job.status }} · {{ detail.job.message }}</p>
        <el-table :data="detail.items" size="small">
          <el-table-column prop="status" label="状态" width="90" />
          <el-table-column prop="title" label="标题" min-width="140" />
          <el-table-column prop="sourceUrl" label="URL" min-width="200" show-overflow-tooltip />
          <el-table-column prop="productId" label="产品ID" width="90" />
          <el-table-column prop="errorMsg" label="错误" min-width="120" show-overflow-tooltip />
        </el-table>
      </template>
    </el-drawer>
  </div>
</template>

<style scoped>
.page-head { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 16px; gap: 16px; }
.page-head h2 { margin: 0 0 6px; font-size: 18px; }
.hint { margin: 0; color: #909399; font-size: 13px; max-width: 640px; }
.pager { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
