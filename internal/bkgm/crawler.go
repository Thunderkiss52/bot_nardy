package bkgm

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type Document struct {
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	Category  string    `json:"category"`
	Text      string    `json:"text"`
	Links     []string  `json:"links,omitempty"`
	FetchedAt time.Time `json:"fetched_at"`
}

type Config struct {
	Seeds      []string
	MaxPages   int
	Delay      time.Duration
	UserAgent  string
	HTTPClient *http.Client
}

type Crawler struct {
	cfg    Config
	client *http.Client
}

func NewCrawler(cfg Config) *Crawler {
	if cfg.MaxPages <= 0 {
		cfg.MaxPages = 128
	}
	if cfg.Delay <= 0 {
		cfg.Delay = 250 * time.Millisecond
	}
	if strings.TrimSpace(cfg.UserAgent) == "" {
		cfg.UserAgent = "bot-nardy-bkgm-importer/1.0 (+local research crawler)"
	}
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	return &Crawler{cfg: cfg, client: client}
}

func DefaultSeeds() []string {
	return []string{
		"https://bkgm.com/articles/index.html",
		"https://bkgm.com/matches/",
		"https://bkgm.com/gloss/",
	}
}

func (c *Crawler) CrawlToJSONL(outPath string) (int, error) {
	f, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	seen := map[string]struct{}{}
	queue := make([]string, 0, len(c.cfg.Seeds))
	for _, seed := range c.cfg.Seeds {
		if normalized, ok := normalizeURL(seed); ok {
			queue = append(queue, normalized)
		}
	}

	enc := json.NewEncoder(w)
	count := 0
	for len(queue) > 0 && count < c.cfg.MaxPages {
		raw := queue[0]
		queue = queue[1:]
		if _, ok := seen[raw]; ok {
			continue
		}
		seen[raw] = struct{}{}

		doc, links, err := c.fetchDocument(raw)
		if err != nil {
			continue
		}
		if strings.TrimSpace(doc.Text) == "" {
			continue
		}
		if err := enc.Encode(doc); err != nil {
			return count, err
		}
		count++

		for _, link := range links {
			if _, ok := seen[link]; ok {
				continue
			}
			queue = append(queue, link)
		}

		time.Sleep(c.cfg.Delay)
	}
	return count, nil
}

func (c *Crawler) fetchDocument(rawURL string) (Document, []string, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return Document{}, nil, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return Document{}, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Document{}, nil, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, rawURL)
	}
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if !strings.Contains(contentType, "text/html") && !strings.Contains(contentType, "text/plain") {
		return Document{}, nil, fmt.Errorf("unsupported content-type %q", contentType)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4*1024*1024))
	if err != nil {
		return Document{}, nil, err
	}
	doc, links, err := parseDocument(rawURL, body)
	if err != nil {
		return Document{}, nil, err
	}
	doc.FetchedAt = time.Now().UTC()
	return doc, links, nil
}

func parseDocument(rawURL string, body []byte) (Document, []string, error) {
	root, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return Document{}, nil, err
	}

	base, err := url.Parse(rawURL)
	if err != nil {
		return Document{}, nil, err
	}

	title := ""
	var titleBuilder strings.Builder
	var textParts []string
	linkSet := map[string]struct{}{}
	skipDepth := 0

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n == nil {
			return
		}
		if n.Type == html.ElementNode {
			tag := strings.ToLower(n.Data)
			if tag == "script" || tag == "style" || tag == "noscript" {
				skipDepth++
			}
			if tag == "a" {
				for _, attr := range n.Attr {
					if strings.EqualFold(attr.Key, "href") {
						if normalized, ok := resolveAllowedLink(base, attr.Val); ok {
							linkSet[normalized] = struct{}{}
						}
						break
					}
				}
			}
		}
		if skipDepth == 0 && n.Type == html.TextNode {
			text := cleanSpace(n.Data)
			if text != "" {
				if n.Parent != nil && strings.EqualFold(n.Parent.Data, "title") {
					if titleBuilder.Len() > 0 {
						titleBuilder.WriteByte(' ')
					}
					titleBuilder.WriteString(text)
				} else {
					textParts = append(textParts, text)
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
		if n.Type == html.ElementNode {
			tag := strings.ToLower(n.Data)
			if tag == "script" || tag == "style" || tag == "noscript" {
				skipDepth--
			}
		}
	}
	walk(root)

	title = cleanSpace(titleBuilder.String())
	if title == "" {
		title = rawURL
	}

	links := make([]string, 0, len(linkSet))
	for link := range linkSet {
		links = append(links, link)
	}

	return Document{
		URL:      rawURL,
		Title:    title,
		Category: categorize(rawURL),
		Text:     cleanSpace(strings.Join(textParts, " ")),
		Links:    links,
	}, links, nil
}

func normalizeURL(raw string) (string, bool) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" {
		return "", false
	}
	if !allowedHost(u.Host) {
		return "", false
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	u.Fragment = ""
	if u.Path == "" {
		u.Path = "/"
	}
	if !allowedPath(u.Path) {
		return "", false
	}
	return u.String(), true
}

func resolveAllowedLink(base *url.URL, href string) (string, bool) {
	href = strings.TrimSpace(href)
	if href == "" || strings.HasPrefix(href, "#") {
		return "", false
	}
	rel, err := url.Parse(href)
	if err != nil {
		return "", false
	}
	resolved := base.ResolveReference(rel)
	return normalizeURL(resolved.String())
}

func allowedHost(host string) bool {
	host = strings.ToLower(host)
	return host == "bkgm.com" || host == "www.bkgm.com"
}

func allowedPath(p string) bool {
	p = strings.ToLower(p)
	if strings.HasSuffix(p, ".jpg") || strings.HasSuffix(p, ".jpeg") || strings.HasSuffix(p, ".gif") || strings.HasSuffix(p, ".png") || strings.HasSuffix(p, ".svg") {
		return false
	}
	if strings.HasSuffix(p, ".css") || strings.HasSuffix(p, ".js") || strings.HasSuffix(p, ".ico") || strings.HasSuffix(p, ".pdf") || strings.HasSuffix(p, ".zip") {
		return false
	}
	return strings.HasPrefix(p, "/articles/") ||
		strings.HasPrefix(p, "/matches/") ||
		strings.HasPrefix(p, "/gloss/") ||
		p == "/"
}

func categorize(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "page"
	}
	p := strings.ToLower(u.Path)
	switch {
	case strings.HasPrefix(p, "/matches/"):
		return "match"
	case strings.HasPrefix(p, "/gloss/"):
		return "glossary"
	case strings.HasPrefix(p, "/articles/"):
		if path.Base(p) == "index.html" || p == "/articles/" {
			return "article-index"
		}
		return "article"
	default:
		return "page"
	}
}

func cleanSpace(s string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(s, "\u00a0", " ")), " ")
}
