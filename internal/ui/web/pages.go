// Package web отдаёт веб-интерфейс приложения, встроенный в бинарник.
//
// Страницы разделены сознательно:
//
//   - "/" (IndexPage) — презентационный лендинг: рассказывает о возможностях
//     платформы. Ни одного скрипта и ни одного запроса к API: всё содержимое
//     лежит в разметке, поэтому страница открывается мгновенно и целиком видна
//     поисковым и AI-краулерам;
//   - "/tools" — рабочая страница: получение рыночных данных и калькуляторы
//     трейдера на PrimeVue;
//   - "/docs/api" — документация REST API, отрисованная из контракта, плюс сам
//     контракт по "/docs/api/openapi.yaml".
//
// Отдельно раздаются иконка приложения по "/favicon.ico" и файлы для поисковых
// систем и AI-ассистентов: robots.txt, sitemap.xml и llms.txt.
//
// Лендинг по-прежнему отдаётся сервером как index-ответ (Server.WithIndexHtmlResponse),
// а остальные страницы и вся статика (CSS, JS, картинки) регистрируются через
// Server.RegistrationPage.
package web

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	openapi "github.com/AlekseyPorandaykin/crypto_polymath/api/rest/v1"
)

//go:embed static/index.html
var IndexPage []byte

//go:embed static/tools.html
var toolsPage []byte

//go:embed static/docs.html
var docsPage []byte

//go:embed static/llms.txt
var llmsFile []byte

//go:embed static/css static/js static/img
var assets embed.FS

// SiteURL — канонический origin без завершающего слеша.
//
// Нужен там, где относительная ссылка не работает: sitemap.xml, canonical,
// Open Graph, JSON-LD. Публичный хост в учебном проекте не публикуем — только
// localhost. Значение продублировано в статике (иначе пришлось бы рендерить её
// шаблоном); TestCanonicalHostIsConsistent следит, чтобы копии не разъехались.
const SiteURL = "http://localhost"

const (
	robotsPath  = "/robots.txt"
	sitemapPath = "/sitemap.xml"
	// llmsPath — соглашение llmstxt.org: короткое описание сервиса в Markdown
	// для ассистентов и AI-поисковиков. В отличие от лендинга его не нужно
	// разбирать из HTML, поэтому ответы ассистентов получаются точнее.
	llmsPath = "/llms.txt"
)

// Страницы, которые имеет смысл индексировать. Страница инструментов сюда не
// входит: её содержимое собирается в браузере, поисковику досталась бы пустая
// оболочка, поэтому в tools.html стоит noindex.
var indexablePages = []string{"/", DocsPath}

// StaticPrefix — префикс, по которому раздаётся встроенная статика. Страницы
// ссылаются на ресурсы абсолютными путями вида /static/css/app.css.
const StaticPrefix = "/static"

// ToolsPath — адрес страницы с данными и калькуляторами. Лендинг ведёт сюда
// кнопкой в шапке, а параметр ?fn=<key> открывает конкретную функцию сразу.
const ToolsPath = "/tools"

const (
	// DocsPath — страница документации REST API.
	DocsPath = "/docs/api"
	// SpecPath — сам OpenAPI-контракт. Его читает не только страница
	// документации: этот же адрес удобно отдавать генераторам клиентов и
	// импортировать в Postman или Insomnia.
	SpecPath = DocsPath + "/openapi.yaml"
)

// Тип для YAML закреплён RFC 9512. Пишем его явно: определение типа по
// расширению .yaml в net/http отсутствует, а без заголовка клиенты получили бы
// application/octet-stream.
const mimeYAML = "application/yaml; charset=UTF-8"

const (
	// faviconPath — исторический адрес иконки. Страницы указывают иконку
	// тегами link, но браузеры и внешние сервисы (превью ссылок, читалки)
	// запрашивают этот путь и без разметки, поэтому он остаётся.
	faviconPath = "/favicon.ico"
	// faviconAsset — иконка внутри встроенной статики. Отдаём PNG: он же лежит
	// в /static/img, а формат ICO ради одного размера держать не нужно —
	// браузеры давно принимают PNG по этому адресу.
	faviconAsset = "img/icon-32.png"
)

type Pages struct {
	assets fs.FS
	// Иконка прочитана один раз при сборке страниц: путь фиксированный, а
	// запрашивают её на каждой странице.
	faviconPNG []byte
}

func NewPages() (*Pages, error) {
	// Каталог static в embed-пути присутствует, а в URL — нет, поэтому уровень срезаем.
	assetsRoot, err := fs.Sub(assets, "static")
	if err != nil {
		return nil, err
	}
	favicon, err := fs.ReadFile(assetsRoot, faviconAsset)
	if err != nil {
		return nil, err
	}
	return &Pages{assets: assetsRoot, faviconPNG: favicon}, nil
}

func (p *Pages) RegistrationPageRoute(g *echo.Group) {
	g.StaticFS(StaticPrefix, p.assets)
	g.GET(ToolsPath, p.tools)
	g.GET(DocsPath, p.docs)
	g.GET(SpecPath, p.spec)
	g.GET(faviconPath, p.favicon)
	g.GET(robotsPath, p.robots)
	g.GET(sitemapPath, p.sitemap)
	g.GET(llmsPath, p.llms)
}

func (p *Pages) favicon(c echo.Context) error {
	return c.Blob(http.StatusOK, "image/png", p.faviconPNG)
}

func (p *Pages) tools(c echo.Context) error {
	return c.HTMLBlob(http.StatusOK, toolsPage)
}

func (p *Pages) docs(c echo.Context) error {
	return c.HTMLBlob(http.StatusOK, docsPage)
}

func (p *Pages) spec(c echo.Context) error {
	return c.Blob(http.StatusOK, mimeYAML, openapi.Specification)
}

func (p *Pages) llms(c echo.Context) error {
	return c.Blob(http.StatusOK, echo.MIMETextPlainCharsetUTF8, llmsFile)
}

// Краулеры ассистентов и AI-поиска перечислены отдельными группами, потому что
// часть из них игнорирует общее правило для "*". Для нас индексация ассистентами
// — способ попасть в их ответы, поэтому доступ разрешаем явно.
//
// Список международный сознательно: аудитория сервиса не привязана к региону,
// поэтому здесь и глобальные поисковики, и локальные (Yandex, Naver, Seznam),
// и краулеры, собирающие корпусы для AI (CCBot).
var crawlers = []string{
	"*",
	"Googlebot", "Bingbot", "DuckDuckBot", "Applebot", "Applebot-Extended",
	"YandexBot", "Baiduspider", "Yeti", "SeznamBot",
	"GPTBot", "OAI-SearchBot", "ChatGPT-User",
	"ClaudeBot", "Claude-User", "Claude-SearchBot",
	"PerplexityBot", "Perplexity-User",
	"Google-Extended", "Google-CloudVertexBot",
	"MistralAI-User", "cohere-ai", "meta-externalagent",
	"Amazonbot", "Bytespider", "YouBot", "DuckAssistBot", "Kagibot", "CCBot",
}

func (p *Pages) robots(c echo.Context) error {
	var b strings.Builder
	for _, agent := range crawlers {
		fmt.Fprintf(&b, "User-agent: %s\nAllow: /\n\n", agent)
	}
	fmt.Fprintf(&b, "Sitemap: %s%s\n", SiteURL, sitemapPath)
	return c.String(http.StatusOK, b.String())
}

func (p *Pages) sitemap(c echo.Context) error {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	for _, page := range indexablePages {
		fmt.Fprintf(&b, "  <url><loc>%s%s</loc></url>\n", SiteURL, page)
	}
	b.WriteString("</urlset>\n")
	return c.Blob(http.StatusOK, echo.MIMEApplicationXMLCharsetUTF8, []byte(b.String()))
}
