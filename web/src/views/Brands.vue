<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  listBrands, saveBrand, deleteBrand,
  listSeries, saveSeries, deleteSeries,
  type Brand, type Series,
} from '../api/catalog'

const loading = ref(false)
const brands = ref<Brand[]>([])
const series = ref<Series[]>([])
const keyword = ref('')
const activeBrandId = ref<number | null>(null)

const brandDialog = ref(false)
const brandForm = ref({ id: 0, name: '', code: '', website: '', country: '', description: '' })
const seriesDialog = ref(false)
const seriesForm = ref({ id: 0, brandId: 0, name: '', code: '', description: '' })

async function loadBrands() {
  loading.value = true
  try {
    brands.value = await listBrands(keyword.value)
    if (activeBrandId.value == null && brands.value.length) {
      activeBrandId.value = brands.value[0].id
    }
  } catch (e) {
    ElMessage.error((e as Error).message)
  } finally {
    loading.value = false
  }
}

async function loadSeries() {
  if (!activeBrandId.value) {
    series.value = []
    return
  }
  series.value = await listSeries(activeBrandId.value)
}

watch(activeBrandId, () => { void loadSeries() })

function openBrand(row?: Brand) {
  brandForm.value = row
    ? { id: row.id, name: row.name, code: row.code, website: row.website, country: row.country, description: row.description }
    : { id: 0, name: '', code: '', website: '', country: '', description: '' }
  brandDialog.value = true
}

async function submitBrand() {
  try {
    await saveBrand(brandForm.value, brandForm.value.id || undefined)
    brandDialog.value = false
    ElMessage.success('已保存')
    await loadBrands()
  } catch (e) {
    ElMessage.error((e as Error).message)
  }
}

async function removeBrand(row: Brand) {
  await ElMessageBox.confirm(`删除品牌「${row.name}」？`, '确认')
  await deleteBrand(row.id)
  if (activeBrandId.value === row.id) activeBrandId.value = null
  ElMessage.success('已删除')
  await loadBrands()
}

function openSeries(row?: Series) {
  if (!activeBrandId.value) {
    ElMessage.warning('请先选择品牌')
    return
  }
  seriesForm.value = row
    ? { id: row.id, brandId: row.brandId, name: row.name, code: row.code, description: row.description }
    : { id: 0, brandId: activeBrandId.value, name: '', code: '', description: '' }
  seriesDialog.value = true
}

async function submitSeries() {
  try {
    await saveSeries(seriesForm.value, seriesForm.value.id || undefined)
    seriesDialog.value = false
    ElMessage.success('已保存')
    await loadSeries()
  } catch (e) {
    ElMessage.error((e as Error).message)
  }
}

async function removeSeries(row: Series) {
  await ElMessageBox.confirm(`删除系列「${row.name}」？`, '确认')
  await deleteSeries(row.id)
  ElMessage.success('已删除')
  await loadSeries()
}

onMounted(async () => {
  await loadBrands()
  await loadSeries()
})
</script>

<template>
  <div v-loading="loading">
    <div class="page-head">
      <h2>品牌与系列</h2>
      <div class="actions">
        <el-input v-model="keyword" clearable placeholder="搜索品牌" style="width: 200px" @keyup.enter="loadBrands" />
        <el-button @click="loadBrands">搜索</el-button>
        <el-button type="primary" @click="openBrand()">新建品牌</el-button>
      </div>
    </div>

    <div class="split">
      <el-card shadow="never" class="pane">
        <template #header>品牌</template>
        <el-table :data="brands" highlight-current-row height="520" @current-change="(r: Brand | undefined) => { if (r) activeBrandId = r.id }">
          <el-table-column prop="name" label="名称" min-width="120" />
          <el-table-column prop="code" label="编码" width="100" />
          <el-table-column label="操作" width="140">
            <template #default="{ row }">
              <el-button link type="primary" @click.stop="openBrand(row)">编辑</el-button>
              <el-button link type="danger" @click.stop="removeBrand(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <el-card shadow="never" class="pane">
        <template #header>
          <div class="pane-head">
            <span>系列</span>
            <el-button size="small" type="primary" @click="openSeries()">新建系列</el-button>
          </div>
        </template>
        <el-table :data="series" height="520">
          <el-table-column prop="name" label="名称" min-width="140" />
          <el-table-column prop="code" label="编码" width="100" />
          <el-table-column label="操作" width="140">
            <template #default="{ row }">
              <el-button link type="primary" @click="openSeries(row)">编辑</el-button>
              <el-button link type="danger" @click="removeSeries(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>

    <el-dialog v-model="brandDialog" :title="brandForm.id ? '编辑品牌' : '新建品牌'" width="520px">
      <el-form label-width="80px">
        <el-form-item label="名称" required><el-input v-model="brandForm.name" /></el-form-item>
        <el-form-item label="编码"><el-input v-model="brandForm.code" /></el-form-item>
        <el-form-item label="官网"><el-input v-model="brandForm.website" /></el-form-item>
        <el-form-item label="国家"><el-input v-model="brandForm.country" /></el-form-item>
        <el-form-item label="简介"><el-input v-model="brandForm.description" type="textarea" :rows="3" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="brandDialog = false">取消</el-button>
        <el-button type="primary" @click="submitBrand">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="seriesDialog" :title="seriesForm.id ? '编辑系列' : '新建系列'" width="520px">
      <el-form label-width="80px">
        <el-form-item label="名称" required><el-input v-model="seriesForm.name" /></el-form-item>
        <el-form-item label="编码"><el-input v-model="seriesForm.code" /></el-form-item>
        <el-form-item label="简介"><el-input v-model="seriesForm.description" type="textarea" :rows="3" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="seriesDialog = false">取消</el-button>
        <el-button type="primary" @click="submitSeries">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-head h2 { margin: 0; font-size: 18px; }
.actions { display: flex; gap: 8px; }
.split { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.pane-head { display: flex; justify-content: space-between; align-items: center; }
@media (max-width: 960px) { .split { grid-template-columns: 1fr; } }
</style>
