<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listProducts, deleteProduct, listBrands, type Product, type Brand } from '../api/catalog'

const router = useRouter()
const loading = ref(false)
const list = ref<Product[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')
const brandId = ref<number | undefined>()
const brands = ref<Brand[]>([])

async function load() {
  loading.value = true
  try {
    const res = await listProducts({
      page: page.value,
      pageSize: pageSize.value,
      keyword: keyword.value || undefined,
      brandId: brandId.value || undefined,
    })
    list.value = res.list || []
    total.value = res.total || 0
  } catch (e) {
    ElMessage.error((e as Error).message)
  } finally {
    loading.value = false
  }
}

async function remove(row: Product) {
  await ElMessageBox.confirm(`删除产品「${row.name}」？`, '确认')
  await deleteProduct(row.id)
  ElMessage.success('已删除')
  await load()
}

onMounted(async () => {
  brands.value = await listBrands()
  await load()
})
</script>

<template>
  <div>
    <div class="page-head">
      <h2>产品库</h2>
      <div class="actions">
        <el-select v-model="brandId" clearable placeholder="品牌" style="width: 160px" @change="load">
          <el-option v-for="b in brands" :key="b.id" :label="b.name" :value="b.id" />
        </el-select>
        <el-input v-model="keyword" clearable placeholder="名称/型号" style="width: 200px" @keyup.enter="load" />
        <el-button @click="load">搜索</el-button>
        <el-button type="primary" @click="router.push('/products/new')">新建产品</el-button>
      </div>
    </div>

    <el-table v-loading="loading" :data="list" stripe>
      <el-table-column prop="name" label="产品名称" min-width="180" />
      <el-table-column prop="modelNo" label="型号" width="140" />
      <el-table-column prop="brandName" label="品牌" width="120" />
      <el-table-column prop="seriesName" label="系列" width="120" />
      <el-table-column label="大套" min-width="180">
        <template #default="{ row }">{{ row.parentName || '—' }}</template>
      </el-table-column>
      <el-table-column label="子规格" width="80">
        <template #default="{ row }">{{ (row.variants || []).length }}</template>
      </el-table-column>
      <el-table-column prop="sourceSite" label="来源站" width="120" />
      <el-table-column label="来源" min-width="160">
        <template #default="{ row }">
          <a v-if="row.sourceUrl" :href="row.sourceUrl" target="_blank" rel="noopener">链接</a>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="160" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="router.push(`/products/${row.id}`)">编辑</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pager">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next"
        @current-change="load"
      />
    </div>
  </div>
</template>

<style scoped>
.page-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-head h2 { margin: 0; font-size: 18px; }
.actions { display: flex; gap: 8px; flex-wrap: wrap; }
.pager { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
