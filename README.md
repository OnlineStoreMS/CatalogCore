# CatalogCore — OSMS 产品中心

搜集整理各品牌产品知识（系列、子规格、介绍与说明书）。**暂不与 ProductCore 商品中心联动**；资料来源以**公开页面爬取**为主（礼貌抓取 JSON-LD/OG，不绕过登录/验证码）。

| 项 | 值 |
|----|-----|
| Go module | `catalogcore` |
| API | `:8104` |
| Web | `:5188` |
| 路径 | `/apps/catalog/` |
| Docker 镜像 | `catalogcore-api`、`catalogcore-web` |
| UserCore app | `catalogcore`（`catalog:read` / `catalog:write`） |
| 存储 | 本地 `./data/uploads` 或平台 MinIO bucket `catalogcore` |
| 端口约定 | [deploy/docs/PORTS.md](../deploy/docs/PORTS.md) |
| 平台编排 | `/home/asialeaf/projects/deploy` |

## 本地开发

```bash
cp configs/config.example.yaml configs/config.yaml
# 编辑 postgres_dsn / jwt_secret（与平台其他 Core 一样使用 PostgreSQL）

go run ./cmd/api -config configs/config.yaml

cd web && npm i && npm run dev
```

首次需在 PostgreSQL 创建库/用户（或走平台）：

```bash
cd ~/projects/deploy
./scripts/init-external-db.sh   # 含 catalogcore
./scripts/init-external-minio.sh  # 含 catalogcore bucket
```

## 功能概览

- **品牌与系列**：维护品牌、产品线
- **产品库**：产品条目 + **子规格** + 来源 URL/站点
- **爬取任务**：种子 URL → `generic_product` 适配器（JSON-LD Product / OpenGraph）→ 写入产品库

## API

- `GET /health`
- Admin（JWT）：`/api/v1/admin/dashboard/stats`、`/brands`、`/series`、`/products`、`/crawl-jobs`、`/upload`
