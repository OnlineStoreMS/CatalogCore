<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  getProduct, saveProduct, listBrands, listSeries, listProducts, setProductParent, seriesOptionLabel,
  type Brand, type Series, type Product, type Variant,
} from '../api/catalog'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const saving = ref(false)
const brands = ref<Brand[]>([])
const series = ref<Series[]>([])
const children = ref<Product['children']>([])
const parentName = ref('')
const picker = ref(false)
const candidates = ref<Product[]>([])

const id = computed(() => {
  const raw = route.params.id
  if (!raw || raw === 'new') return 0
  return Number(raw) || 0
})

const form = ref({
  brandId: 0,
  seriesId: 0,
  parentId: 0,
  name: '',
  modelNo: '',
  subtitle: '',
  coverUrl: '',
  intro: '',
  detailHtml: '',
  specsJson: '',
  tagsJson: '',
  sourceUrl: '',
  sourceSite: '',
  status: 1,
  variants: [] as Variant[],
  attachments: [] as { kind: string; name: string; url: string }[],
})

async function loadSeriesForBrand() {
  if (!form.value.brandId) {
    series.value = []
    return
  }
  series.value = await listSeries(form.value.brandId)
}

function addVariant() {
  form.value.variants.push({ name: '', skuCode: '', priceHint: '', sort: form.value.variants.length })
}

function removeVariant(i: number) {
  form.value.variants.splice(i, 1)
}

async function load() {
  brands.value = await listBrands()
  if (!id.value) return
  loading.value = true
  try {
    const p = await getProduct(id.value)
    form.value = {
      brandId: p.brandId,
      seriesId: p.seriesId,
      parentId: p.parentId || 0,
      name: p.name,
      modelNo: p.modelNo,
      subtitle: p.subtitle,
      coverUrl: p.coverUrl,
      intro: p.intro,
      detailHtml: p.detailHtml || '',
      specsJson: p.specsJson,
      tagsJson: p.tagsJson,
      sourceUrl: p.sourceUrl,
      sourceSite: p.sourceSite,
      status: p.status,
      variants: p.variants || [],
      attachments: (p.attachments || []).map((a) => ({ kind: a.kind, name: a.name, url: a.url })),
    }
    children.value = p.children || []
    parentName.value = p.parentName || ''
    await loadSeriesForBrand()
  } catch (e) {
    ElMessage.error((e as Error).message)
  } finally {
    loading.value = false
  }
}

async function submit() {
  if (!form.value.name.trim()) {
    ElMessage.warning('请填写产品名称')
    return
  }
  if (!form.value.brandId) {
    ElMessage.warning('请选择品牌')
    return
  }
  saving.value = true
  try {
    const saved = await saveProduct({ ...form.value, parentId: form.value.parentId || 0 }, id.value || undefined)
    ElMessage.success('已保存')
    if (!id.value) router.replace(`/products/${saved.id}`)
    else await load()
  } catch (e) {
    ElMessage.error((e as Error).message)
  } finally {
    saving.value = false
  }
}

async function openPicker() {
  if (!form.value.brandId) {
    ElMessage.warning('请先选择品牌')
    return
  }
  const res = await listProducts({ brandId: form.value.brandId, page: 1, pageSize: 200 })
  const taken = new Set((children.value || []).map((c) => c.id))
  candidates.value = (res.list || []).filter((p) => p.id !== id.value && !p.parentId && !p.childCount && !taken.has(p.id))
  picker.value = true
}

async function addComponent(row: Product) {
  await setProductParent(row.id, id.value)
  picker.value = false
  ElMessage.success('已加入组成品')
  await load()
}

async function removeComponent(componentId: number) {
  await setProductParent(componentId, 0)
  ElMessage.success('已移出')
  await load()
}

onMounted(load)
</script>

<template>
  <div v-loading="loading" class="edit">
    <div class="page-head">
      <h2>{{ id ? '编辑产品' : '新建产品' }}</h2>
      <div>
        <el-button @click="router.push('/products')">返回</el-button>
        <el-button type="primary" :loading="saving" @click="submit">保存</el-button>
      </div>
    </div>

    <el-form label-width="96px" class="form">
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="品牌" required>
            <el-select v-model="form.brandId" filterable style="width: 100%" @change="() => { form.seriesId = 0; loadSeriesForBrand() }">
              <el-option v-for="b in brands" :key="b.id" :label="b.name" :value="b.id" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="系列">
            <el-select v-model="form.seriesId" clearable filterable style="width: 100%">
              <el-option v-for="s in series" :key="s.id" :label="seriesOptionLabel(s)" :value="s.id" />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="16">
        <el-col :span="12"><el-form-item label="名称" required><el-input v-model="form.name" /></el-form-item></el-col>
        <el-col :span="12"><el-form-item label="型号"><el-input v-model="form.modelNo" /></el-form-item></el-col>
      </el-row>

      <el-form-item label="副标题"><el-input v-model="form.subtitle" /></el-form-item>
      <el-form-item label="封面 URL"><el-input v-model="form.coverUrl" /></el-form-item>
      <el-form-item label="简介"><el-input v-model="form.intro" type="textarea" :rows="3" /></el-form-item>
      <el-form-item label="详情 HTML"><el-input v-model="form.detailHtml" type="textarea" :rows="6" /></el-form-item>
      <el-form-item label="规格 JSON"><el-input v-model="form.specsJson" type="textarea" :rows="4" placeholder='{"key":"value"}' /></el-form-item>

      <el-row :gutter="16">
        <el-col :span="12"><el-form-item label="来源 URL"><el-input v-model="form.sourceUrl" /></el-form-item></el-col>
        <el-col :span="12"><el-form-item label="来源站"><el-input v-model="form.sourceSite" /></el-form-item></el-col>
      </el-row>

      <template v-if="id && !form.parentId">
        <el-divider content-position="left">组成品</el-divider>
        <div class="components">
          <el-table v-if="children?.length" :data="children" size="small">
            <el-table-column prop="name" label="产品" min-width="180" />
            <el-table-column prop="modelNo" label="型号" width="140" />
            <el-table-column label="操作" width="140">
              <template #default="{ row }">
                <el-button link type="primary" @click="router.push(`/products/${row.id}`)">查看</el-button>
                <el-button link type="danger" @click="removeComponent(row.id)">移出</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div v-else class="hint">还没有组成品。加上之后，这个产品会显示为组合品。</div>
          <el-button class="add-component" @click="openPicker">添加组成品</el-button>
        </div>
      </template>
      <div v-else-if="id && parentName" class="hint member-hint">这是「{{ parentName }}」的组成品，本身也是一条产品。</div>

      <el-divider content-position="left">子规格</el-divider>
      <div class="variants">
        <div v-for="(v, i) in form.variants" :key="i" class="variant-row">
          <el-input v-model="v.name" placeholder="规格名" />
          <el-input v-model="v.skuCode" placeholder="SKU/编码" />
          <el-input v-model="v.priceHint" placeholder="参考价" />
          <el-input v-model="v.specJson" placeholder="规格 JSON" />
          <el-button type="danger" link @click="removeVariant(i)">删除</el-button>
        </div>
        <el-button @click="addVariant">添加子规格</el-button>
      </div>
    </el-form>

    <el-dialog v-model="picker" title="添加组成品" width="640px">
      <el-table :data="candidates" height="360">
        <el-table-column prop="name" label="产品" min-width="180" />
        <el-table-column prop="modelNo" label="型号" width="140" />
        <el-table-column label="操作" width="80">
          <template #default="{ row }">
            <el-button link type="primary" @click="addComponent(row)">加入</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-head h2 { margin: 0; font-size: 18px; }
.form { max-width: 960px; }
.variants { display: flex; flex-direction: column; gap: 8px; margin-left: 96px; }
.variant-row { display: grid; grid-template-columns: 1.2fr 1fr 0.8fr 1.4fr auto; gap: 8px; align-items: center; }
.components { margin-left: 96px; }
.add-component { margin-top: 8px; }
.hint { color: var(--el-text-color-secondary); font-size: 13px; }
.member-hint { margin: 0 0 16px 96px; }
</style>
