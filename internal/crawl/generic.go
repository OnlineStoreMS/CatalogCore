package crawl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// ParsedProduct 公开页面解析结果（可扩展站点适配器）
type ParsedProduct struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Image       string            `json:"image"`
	Brand       string            `json:"brand,omitempty"`
	SKU         string            `json:"sku,omitempty"`
	Offers      string            `json:"offers,omitempty"`
	Specs       map[string]string `json:"specs,omitempty"`
	Variants    []ParsedVariant   `json:"variants,omitempty"`
	RawJSONLD   json.RawMessage   `json:"rawJsonLd,omitempty"`
}

type ParsedVariant struct {
	Name  string            `json:"name"`
	SKU   string            `json:"sku,omitempty"`
	Specs map[string]string `json:"specs,omitempty"`
	Price string            `json:"price,omitempty"`
	Image string            `json:"image,omitempty"`
}

type Adapter interface {
	Name() string
	Parse(ctx context.Context, pageURL string, body []byte) (*ParsedProduct, error)
}

type Fetcher struct {
	client *http.Client
	ua     string
}

func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{Timeout: 25 * time.Second},
		ua:     "OSMS-CatalogCore/1.0 (+product-catalog research; polite fetch)",
	}
}

func (f *Fetcher) Get(ctx context.Context, pageURL string) ([]byte, error) {
	u, err := url.Parse(pageURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("invalid url")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("only http(s) supported")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", f.ua)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/json;q=0.9,*/*;q=0.8")
	res, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("http %d", res.StatusCode)
	}
	// 限制体积，避免超大页拖垮内存
	limited := io.LimitReader(res.Body, 4<<20)
	return io.ReadAll(limited)
}

var (
	reJSONLD = regexp.MustCompile(`(?is)<script[^>]*type=["']application/ld\+json["'][^>]*>(.*?)</script>`)
	reOG     = regexp.MustCompile(`(?is)<meta[^>]+property=["']og:([^"']+)["'][^>]+content=["']([^"']*)["']`)
	reTitle  = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
)

// GenericProductAdapter 解析常见产品页：JSON-LD Product / OpenGraph
type GenericProductAdapter struct{}

func (a *GenericProductAdapter) Name() string { return "generic_product" }

func (a *GenericProductAdapter) Parse(_ context.Context, pageURL string, body []byte) (*ParsedProduct, error) {
	html := string(body)
	out := &ParsedProduct{Specs: map[string]string{}}

	for _, m := range reJSONLD.FindAllStringSubmatch(html, -1) {
		raw := strings.TrimSpace(m[1])
		if raw == "" {
			continue
		}
		if p := parseJSONLDProduct([]byte(raw)); p != nil {
			mergeParsed(out, p)
			out.RawJSONLD = json.RawMessage(raw)
		}
	}

	for _, m := range reOG.FindAllStringSubmatch(html, -1) {
		prop := strings.ToLower(m[1])
		val := strings.TrimSpace(m[2])
		switch prop {
		case "title":
			if out.Title == "" {
				out.Title = val
			}
		case "description":
			if out.Description == "" {
				out.Description = val
			}
		case "image":
			if out.Image == "" {
				out.Image = absURL(pageURL, val)
			}
		}
	}

	if out.Title == "" {
		if m := reTitle.FindStringSubmatch(html); len(m) > 1 {
			out.Title = strings.TrimSpace(stripTags(m[1]))
		}
	}
	if out.Title == "" {
		return nil, fmt.Errorf("no product title found")
	}
	if out.Image != "" {
		out.Image = absURL(pageURL, out.Image)
	}
	return out, nil
}

func parseJSONLDProduct(raw []byte) *ParsedProduct {
	var node map[string]any
	if err := json.Unmarshal(raw, &node); err != nil {
		// @graph
		var wrap struct {
			Graph []map[string]any `json:"@graph"`
		}
		if err2 := json.Unmarshal(raw, &wrap); err2 != nil {
			return nil
		}
		for _, n := range wrap.Graph {
			if p := fromJSONLDNode(n); p != nil {
				return p
			}
		}
		return nil
	}
	if t, ok := node["@type"]; ok {
		if isProductType(t) {
			return fromJSONLDNode(node)
		}
	}
	// sometimes array root
	var arr []map[string]any
	if err := json.Unmarshal(raw, &arr); err == nil {
		for _, n := range arr {
			if p := fromJSONLDNode(n); p != nil {
				return p
			}
		}
	}
	return fromJSONLDNode(node)
}

func isProductType(t any) bool {
	switch v := t.(type) {
	case string:
		return strings.EqualFold(v, "Product") || strings.EqualFold(v, "ProductGroup")
	case []any:
		for _, x := range v {
			if s, ok := x.(string); ok && (strings.EqualFold(s, "Product") || strings.EqualFold(s, "ProductGroup")) {
				return true
			}
		}
	}
	return false
}

func fromJSONLDNode(n map[string]any) *ParsedProduct {
	if !isProductType(n["@type"]) && n["name"] == nil {
		return nil
	}
	p := &ParsedProduct{Specs: map[string]string{}}
	p.Title = asString(n["name"])
	p.Description = asString(n["description"])
	p.SKU = asString(n["sku"])
	if p.SKU == "" {
		p.SKU = asString(n["mpn"])
	}
	p.Brand = brandName(n["brand"])
	p.Image = firstImage(n["image"])
	if offers := n["offers"]; offers != nil {
		b, _ := json.Marshal(offers)
		p.Offers = string(b)
	}
	if p.Title == "" {
		return nil
	}
	return p
}

func brandName(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case map[string]any:
		return asString(t["name"])
	}
	return ""
}

func firstImage(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []any:
		if len(t) > 0 {
			return firstImage(t[0])
		}
	case map[string]any:
		return asString(t["url"])
	}
	return ""
}

func asString(v any) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case float64:
		return strings.TrimSpace(fmt.Sprintf("%v", t))
	default:
		return ""
	}
}

func mergeParsed(dst, src *ParsedProduct) {
	if dst.Title == "" {
		dst.Title = src.Title
	}
	if dst.Description == "" {
		dst.Description = src.Description
	}
	if dst.Image == "" {
		dst.Image = src.Image
	}
	if dst.Brand == "" {
		dst.Brand = src.Brand
	}
	if dst.SKU == "" {
		dst.SKU = src.SKU
	}
	if dst.Offers == "" {
		dst.Offers = src.Offers
	}
	for k, v := range src.Specs {
		if dst.Specs[k] == "" {
			dst.Specs[k] = v
		}
	}
	if len(dst.Variants) == 0 {
		dst.Variants = src.Variants
	}
}

func absURL(base, ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return ""
	}
	u, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	if u.IsAbs() {
		return u.String()
	}
	b, err := url.Parse(base)
	if err != nil {
		return ref
	}
	return b.ResolveReference(u).String()
}

func stripTags(s string) string {
	return regexp.MustCompile(`<[^>]+>`).ReplaceAllString(s, "")
}

func LookupAdapter(name string) Adapter {
	switch strings.TrimSpace(name) {
	case "", "generic_product":
		return &GenericProductAdapter{}
	default:
		return &GenericProductAdapter{}
	}
}

func SplitSeedURLs(raw string) []string {
	raw = strings.ReplaceAll(raw, ",", "\n")
	raw = strings.ReplaceAll(raw, ";", "\n")
	parts := strings.Split(raw, "\n")
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}
