// Тесты веб-интерфейса.
//
// Разметка и статика встроены в бинарник через go:embed, поэтому опечатка в пути
// к CSS, скрипту или картинке компилируется без ошибок и ломается только в
// браузере у пользователя. Тесты закрывают именно этот класс проблем:
//
//   - все ссылки /static/... со всех страниц ведут на реально встроенные файлы;
//   - роуты /tools, /docs/api и /docs/api/openapi.yaml отвечают, статика
//     отдаётся по /static;
//   - разделение страниц не разъехалось: лендинг не тянет PrimeVue и не содержит
//     форм запросов, а остальные страницы ведут обратно на лендинг;
//   - версии Vue на странице инструментов совпадают у "vue" и у PrimeVue —
//     при расхождении provide/inject внутри PrimeVue молча перестаёт работать;
//   - по /docs/api/openapi.yaml отдаётся ровно тот контракт, что лежит в
//     api/rest/v1/openapi.yaml, и с правильным типом содержимого;
//   - иконка приложения встроена во всех вариантах, объявлена на каждой странице
//     и отдаётся по /favicon.ico.
package web

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	openapi "github.com/AlekseyPorandaykin/crypto_polymath/api/rest/v1"
)

// Ссылки ищем только внутри кавычек атрибутов: в комментариях к разметке пути
// тоже упоминаются, но там они часть текста, а не ресурс.
var staticRefRe = regexp.MustCompile(`["'](/static/[^"']+)["']`)

func readAsset(t *testing.T, name string) string {
	t.Helper()
	pages, err := NewPages()
	if err != nil {
		t.Fatalf("NewPages: %v", err)
	}
	content, err := fs.ReadFile(pages.assets, name)
	if err != nil {
		t.Fatalf("чтение %s: %v", name, err)
	}
	return string(content)
}

func newTestServer(t *testing.T) *echo.Echo {
	t.Helper()
	pages, err := NewPages()
	if err != nil {
		t.Fatalf("NewPages: %v", err)
	}
	e := echo.New()
	pages.RegistrationPageRoute(e.Group(""))
	return e
}

func do(t *testing.T, e *echo.Echo, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// Каждая ссылка на статику из HTML должна существовать во встроенной файловой
// системе и отдаваться сервером — иначе страница откроется без стилей.
func TestStaticReferencesExist(t *testing.T) {
	pages, err := NewPages()
	if err != nil {
		t.Fatalf("NewPages: %v", err)
	}
	e := newTestServer(t)

	documents := map[string][]byte{
		"index.html": IndexPage,
		"tools.html": toolsPage,
		"docs.html":  docsPage,
	}
	for name, page := range documents {
		refs := staticRefRe.FindAllStringSubmatch(string(page), -1)
		if len(refs) == 0 {
			t.Errorf("%s: не найдено ни одной ссылки на /static", name)
		}
		for _, match := range refs {
			ref := match[1]
			assetPath := strings.TrimPrefix(ref, StaticPrefix+"/")
			if _, err := fs.Stat(pages.assets, assetPath); err != nil {
				t.Errorf("%s ссылается на %s, но файл не встроен: %v", name, ref, err)
				continue
			}
			if rec := do(t, e, ref); rec.Code != http.StatusOK {
				t.Errorf("%s ссылается на %s, сервер ответил %d", name, ref, rec.Code)
			}
		}
	}
}

// Страница инструментов должна отдаваться как HTML по своему адресу.
func TestToolsPageRoute(t *testing.T) {
	rec := do(t, newTestServer(t), ToolsPath)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s: код %d, ожидался %d", ToolsPath, rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get(echo.HeaderContentType); !strings.Contains(ct, echo.MIMETextHTML) {
		t.Errorf("GET %s: Content-Type %q, ожидался HTML", ToolsPath, ct)
	}
	body := rec.Body.String()
	for _, want := range []string{`id="tools-app"`, "/static/js/tools.js", `href="/"`} {
		if !strings.Contains(body, want) {
			t.Errorf("страница инструментов не содержит %q", want)
		}
	}
}

// Страница документации должна отдаваться как HTML и ссылаться на контракт.
func TestDocsPageRoute(t *testing.T) {
	rec := do(t, newTestServer(t), DocsPath)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s: код %d, ожидался %d", DocsPath, rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get(echo.HeaderContentType); !strings.Contains(ct, echo.MIMETextHTML) {
		t.Errorf("GET %s: Content-Type %q, ожидался HTML", DocsPath, ct)
	}
	body := rec.Body.String()
	for _, want := range []string{`id="redoc"`, SpecPath, "/static/js/docs.js"} {
		if !strings.Contains(body, want) {
			t.Errorf("страница документации не содержит %q", want)
		}
	}
}

// По SpecPath отдаётся контракт целиком, как YAML.
func TestSpecRoute(t *testing.T) {
	rec := do(t, newTestServer(t), SpecPath)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s: код %d, ожидался %d", SpecPath, rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get(echo.HeaderContentType); ct != mimeYAML {
		t.Errorf("GET %s: Content-Type %q, ожидался %q", SpecPath, ct, mimeYAML)
	}
	if !bytes.Equal(rec.Body.Bytes(), openapi.Specification) {
		t.Errorf("GET %s: отдано %d байт, во встроенном контракте %d",
			SpecPath, rec.Body.Len(), len(openapi.Specification))
	}
	body := rec.Body.String()
	for _, want := range []string{"openapi: 3.0.3", "paths:", "/calculator/trading/risk-at-price:"} {
		if !strings.Contains(body, want) {
			t.Errorf("в отданном контракте нет %q", want)
		}
	}
}

// Встроенный контракт должен совпадать с файлом в репозитории. Копия внутри
// бинарника обновляется только пересборкой, поэтому расхождение означает, что
// спецификацию поправили, а документацию и сгенерированный код — нет.
func TestSpecMatchesRepositoryFile(t *testing.T) {
	path := filepath.Join("..", "..", "..", "api", "rest", "v1", "openapi.yaml")
	onDisk, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("чтение %s: %v", path, err)
	}
	if !bytes.Equal(onDisk, openapi.Specification) {
		t.Errorf("встроенный контракт (%d байт) отличается от %s (%d байт)",
			len(openapi.Specification), path, len(onDisk))
	}
}

// Ссылки «API» на страницах должны вести на локальную документацию: раньше они
// уходили на GitHub, где лежит версия из ветки main, а не та, что развёрнута.
func TestApiLinksPointToLocalDocs(t *testing.T) {
	for name, page := range map[string][]byte{"index.html": IndexPage, "tools.html": toolsPage} {
		body := string(page)
		if !strings.Contains(body, `href="`+DocsPath+`"`) {
			t.Errorf("%s: нет ссылки на %s", name, DocsPath)
		}
		if strings.Contains(body, "github.com/AlekseyPorandaykin/crypto_polymath/blob/main/api/rest/v1/openapi.yaml") {
			t.Errorf("%s: ссылка на контракт ведёт на GitHub вместо %s", name, SpecPath)
		}
	}
}

// Лендинг остаётся презентационным: никаких фронтенд-фреймворков и форм запроса,
// зато есть переход на страницу инструментов.
func TestLandingHasNoConsole(t *testing.T) {
	body := strings.ToLower(string(IndexPage))
	// Ищем именно подключение библиотеки, а не упоминание её названия в комментарии.
	for _, unwanted := range []string{"npm/primevue", "esm.sh/primevue", "primeicons", "esm.sh/vue", "importmap"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("лендинг подключает %q: страница должна оставаться статической, "+
				"а всё интерактивное жить на %s", unwanted, ToolsPath)
		}
	}
	if !strings.Contains(body, `href="`+ToolsPath+`"`) {
		t.Errorf("на лендинге нет перехода на %s", ToolsPath)
	}
}

// Самое важное свойство лендинга для поиска: содержимое лежит в разметке, а не
// собирается скриптом. Поисковые роботы JavaScript исполняют не всегда, а
// AI-краулеры — почти никогда, поэтому клиентский рендеринг сделал бы страницу
// пустой для них. Тест фиксирует, что ключевые смыслы есть в самом HTML.
func TestLandingContentIsServerRendered(t *testing.T) {
	body := string(IndexPage)

	if strings.Contains(body, "v-cloak") {
		t.Error("на лендинге остался v-cloak: до загрузки скрипта страница будет пустой для краулеров")
	}
	if !strings.Contains(body, "<h1") {
		t.Error("на лендинге нет заголовка h1")
	}
	for _, exchange := range landingExchanges {
		if !strings.Contains(body, exchange) {
			t.Errorf("в разметке лендинга нет биржи %q", exchange)
		}
	}
	// Каждый ответ FAQ должен быть в разметке целиком, иначе структурированные
	// данные окажутся богаче видимого текста — Google это считает нарушением.
	for _, fragment := range []string{
		"Do I need to sign up",
		"Which exchanges are supported",
		"trader calculators",
	} {
		if !strings.Contains(body, fragment) {
			t.Errorf("в разметке лендинга нет текста %q", fragment)
		}
	}
}

var landingExchanges = []string{"Binance", "Bybit", "OKX", "Kraken", "KuCoin", "Gate.io"}

// На главной заявлено шесть популярных бирж — ровно столько и должно быть
// перечислено. Подключено их больше, но остальные на лендинг не выводим.
func TestLandingListsSixExchanges(t *testing.T) {
	body := string(IndexPage)

	if want := "<b>6</b> major exchanges"; !strings.Contains(body, want) {
		t.Errorf("на лендинге нет формулировки %q", want)
	}

	chips := regexp.MustCompile(`(?s)<ul class="chips".*?</ul>`).FindString(body)
	if chips == "" {
		t.Fatal("на лендинге не найден список бирж")
	}
	items := regexp.MustCompile(`<li>([^<]+)</li>`).FindAllStringSubmatch(chips, -1)
	if len(items) != len(landingExchanges) {
		t.Errorf("в списке %d бирж, ожидалось %d", len(items), len(landingExchanges))
	}
	for i, item := range items {
		if i < len(landingExchanges) && item[1] != landingExchanges[i] {
			t.Errorf("биржа %d: %q, ожидалась %q", i, item[1], landingExchanges[i])
		}
	}
}

// Лендинг не должен светить внутренности: ссылки на исходный код, на панель
// мониторинга и на сырой контракт — это не то, что помогает пользователю
// разобраться в сервисе.
func TestLandingHidesTechnicalDetails(t *testing.T) {
	body := string(IndexPage)
	forbidden := map[string]string{
		"github.com":    "ссылка на исходный код",
		"sentry.io":     "ссылка на внутренний мониторинг",
		SpecPath:        "ссылка на сырой OpenAPI-контракт",
		"/api/v1/price": "прямая ссылка на эндпоинт API",
	}
	for fragment, what := range forbidden {
		if strings.Contains(body, fragment) {
			t.Errorf("на лендинге осталась %s (%q)", what, fragment)
		}
	}
}

// Набор мета-тегов, без которых страница теряет позиции в поиске и выглядит
// сломанной при пересылке ссылки в мессенджер.
func TestLandingSeoMeta(t *testing.T) {
	body := string(IndexPage)
	required := map[string]string{
		`<link rel="canonical" href="` + SiteURL + `/">`:      "canonical",
		`<meta name="description"`:                            "описание страницы",
		`<meta name="robots"`:                                 "директива для роботов",
		`<meta property="og:title"`:                           "og:title",
		`<meta property="og:description"`:                     "og:description",
		`<meta property="og:image"`:                           "og:image",
		`<meta property="og:url" content="` + SiteURL + `/">`: "og:url",
		`<meta name="twitter:card"`:                           "twitter:card",
	}
	for fragment, what := range required {
		if !strings.Contains(body, fragment) {
			t.Errorf("на лендинге нет %s (%q)", what, fragment)
		}
	}

	// Слишком длинный title поисковик обрежет, слишком длинное описание — тоже.
	title := regexp.MustCompile(`<title>([^<]+)</title>`).FindStringSubmatch(body)
	if title == nil {
		t.Fatal("на лендинге нет title")
	}
	if n := len([]rune(title[1])); n > 65 {
		t.Errorf("title из %d символов будет обрезан в выдаче: %q", n, title[1])
	}
	desc := regexp.MustCompile(`<meta name="description" content="([^"]+)"`).FindStringSubmatch(body)
	if desc == nil {
		t.Fatal("на лендинге нет description")
	}
	if n := len([]rune(desc[1])); n > 170 {
		t.Errorf("description из %d символов будет обрезан в выдаче", n)
	}
}

// Сервис рассчитан на международную аудиторию, и язык объявлен в трёх местах:
// атрибутом lang, для соцсетей — og:locale, для поиска и ассистентов —
// inLanguage в структурированных данных. Если копии разъедутся, поисковик может
// решить, что страница не на том языке, и подменить сниппет автопереводом.
func TestPagesDeclareEnglish(t *testing.T) {
	for name, page := range map[string][]byte{
		"index.html": IndexPage,
		"tools.html": toolsPage,
		"docs.html":  docsPage,
	} {
		if !strings.Contains(string(page), `<html lang="en">`) {
			t.Errorf(`%s: язык страницы не объявлен как lang="en"`, name)
		}
	}

	body := string(IndexPage)
	if !strings.Contains(body, `<meta property="og:locale" content="en_US">`) {
		t.Error("на лендинге нет og:locale для англоязычной аудитории")
	}
	if strings.Contains(body, `"inLanguage": "ru-RU"`) {
		t.Error("в структурированных данных остался русский inLanguage")
	}
	if !strings.Contains(body, `"inLanguage": "en"`) {
		t.Error("в структурированных данных не объявлен английский язык")
	}
}

// Весь фронтенд ведётся на английском — и то, что читает посетитель, и
// комментарии для разработчика. Смешанный язык здесь дороже, чем кажется:
// правку в этих файлах делает то же, кто читает разметку и стили, и переключаться
// между языками внутри одного файла приходится на каждом экране.
//
// Проверяем всё дерево целиком, а не отдельные файлы: новый скрипт или картинка
// попадут под правило сами, без правки теста.
func TestStaticFilesAreEnglish(t *testing.T) {
	cyrillic := regexp.MustCompile(`\p{Cyrillic}+[^\n]*`)

	documents := map[string]string{
		"index.html": string(IndexPage),
		"tools.html": string(toolsPage),
		"docs.html":  string(docsPage),
		"llms.txt":   string(llmsFile),
	}

	pages, err := NewPages()
	if err != nil {
		t.Fatalf("NewPages: %v", err)
	}
	// Двоичные файлы пропускаем: осмысленного текста в них нет, а случайная
	// последовательность байт способна совпасть с кириллической буквой.
	binary := map[string]bool{".png": true, ".webp": true, ".ico": true}
	err = fs.WalkDir(pages.assets, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || binary[strings.ToLower(filepath.Ext(path))] {
			return err
		}
		content, err := fs.ReadFile(pages.assets, path)
		if err != nil {
			return err
		}
		documents[path] = string(content)
		return nil
	})
	if err != nil {
		t.Fatalf("обход встроенной статики: %v", err)
	}
	if len(documents) < 10 {
		t.Fatalf("проверено всего %d файлов — похоже, обход статики сломался", len(documents))
	}

	for name, content := range documents {
		for _, found := range cyrillic.FindAllString(content, -1) {
			t.Errorf("%s: непереведённый текст %q", name, found)
		}
	}
}

// Картинка для превью в соцсетях указана абсолютной ссылкой, поэтому опечатку в
// пути обычная проверка /static-ссылок не поймает — проверяем отдельно.
func TestLandingOpenGraphImageExists(t *testing.T) {
	pages, err := NewPages()
	if err != nil {
		t.Fatalf("NewPages: %v", err)
	}
	m := regexp.MustCompile(`<meta property="og:image" content="([^"]+)"`).FindStringSubmatch(string(IndexPage))
	if m == nil {
		t.Fatal("на лендинге нет og:image")
	}
	if !strings.HasPrefix(m[1], SiteURL+StaticPrefix+"/") {
		t.Fatalf("og:image %q должен быть абсолютной ссылкой на %s%s", m[1], SiteURL, StaticPrefix)
	}
	assetPath := strings.TrimPrefix(m[1], SiteURL+StaticPrefix+"/")
	if _, err := fs.Stat(pages.assets, assetPath); err != nil {
		t.Errorf("og:image ссылается на %s, но файл не встроен: %v", m[1], err)
	}
}

// Структурированные данные не имеют визуального представления: синтаксическая
// ошибка не ломает страницу, а просто молча лишает её расширенных сниппетов.
// Поэтому разбираем JSON-LD и сверяем вопросы FAQ с видимыми на странице:
// Google требует, чтобы разметка совпадала с содержимым.
func TestLandingStructuredData(t *testing.T) {
	body := string(IndexPage)

	raw := regexp.MustCompile(`(?s)<script type="application/ld\+json">(.*?)</script>`).FindStringSubmatch(body)
	if raw == nil {
		t.Fatal("на лендинге нет структурированных данных JSON-LD")
	}

	var doc struct {
		Context string `json:"@context"`
		Graph   []struct {
			Type       string `json:"@type"`
			MainEntity []struct {
				Name           string `json:"name"`
				AcceptedAnswer struct {
					Text string `json:"text"`
				} `json:"acceptedAnswer"`
			} `json:"mainEntity"`
		} `json:"@graph"`
	}
	if err := json.Unmarshal([]byte(raw[1]), &doc); err != nil {
		t.Fatalf("JSON-LD не разбирается: %v", err)
	}
	if doc.Context != "https://schema.org" {
		t.Errorf("@context = %q, ожидался https://schema.org", doc.Context)
	}

	types := make(map[string]bool, len(doc.Graph))
	for _, node := range doc.Graph {
		types[node.Type] = true
	}
	for _, want := range []string{"WebSite", "WebApplication", "FAQPage"} {
		if !types[want] {
			t.Errorf("в структурированных данных нет типа %s", want)
		}
	}

	visible := regexp.MustCompile(`<summary>([^<]+)</summary>`).FindAllStringSubmatch(body, -1)
	if len(visible) == 0 {
		t.Fatal("на странице нет видимого блока вопросов")
	}
	for _, node := range doc.Graph {
		if node.Type != "FAQPage" {
			continue
		}
		if len(node.MainEntity) != len(visible) {
			t.Errorf("в разметке %d вопросов, на странице видно %d — Google требует совпадения",
				len(node.MainEntity), len(visible))
		}
		for i, question := range node.MainEntity {
			if question.AcceptedAnswer.Text == "" {
				t.Errorf("у вопроса %q нет ответа в разметке", question.Name)
			}
			if i < len(visible) && question.Name != visible[i][1] {
				t.Errorf("вопрос %d в разметке — %q, а на странице — %q", i, question.Name, visible[i][1])
			}
		}
	}
}

// robots.txt должен разрешать обход и указывать на карту сайта, а краулеры
// ассистентов — быть перечислены явно: часть из них не читает правило для "*".
func TestRobotsTxt(t *testing.T) {
	rec := do(t, newTestServer(t), robotsPath)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s: код %d", robotsPath, rec.Code)
	}
	if ct := rec.Header().Get(echo.HeaderContentType); !strings.Contains(ct, "text/plain") {
		t.Errorf("GET %s: Content-Type %q, ожидался text/plain", robotsPath, ct)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Sitemap: "+SiteURL+sitemapPath) {
		t.Error("в robots.txt нет ссылки на sitemap.xml")
	}
	if strings.Contains(body, "Disallow: /\n") {
		t.Error("robots.txt запрещает обход всего сайта")
	}
	// Список сознательно международный: помимо ассистентов здесь и глобальные
	// поисковики, и краулер Common Crawl, из корпуса которого учатся модели.
	for _, agent := range []string{
		"GPTBot", "ClaudeBot", "PerplexityBot", "OAI-SearchBot",
		"Googlebot", "Bingbot", "Applebot", "CCBot",
	} {
		if !strings.Contains(body, "User-agent: "+agent) {
			t.Errorf("в robots.txt не разрешён краулер %s", agent)
		}
	}
}

// Карта сайта должна быть валидным XML и содержать только индексируемые адреса.
func TestSitemapXml(t *testing.T) {
	rec := do(t, newTestServer(t), sitemapPath)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s: код %d", sitemapPath, rec.Code)
	}
	if ct := rec.Header().Get(echo.HeaderContentType); !strings.Contains(ct, "xml") {
		t.Errorf("GET %s: Content-Type %q, ожидался XML", sitemapPath, ct)
	}

	var sitemap struct {
		URLs []struct {
			Loc string `xml:"loc"`
		} `xml:"url"`
	}
	if err := xml.Unmarshal(rec.Body.Bytes(), &sitemap); err != nil {
		t.Fatalf("sitemap.xml не разбирается: %v", err)
	}
	if len(sitemap.URLs) != len(indexablePages) {
		t.Fatalf("в карте сайта %d адресов, ожидалось %d", len(sitemap.URLs), len(indexablePages))
	}
	for i, u := range sitemap.URLs {
		if want := SiteURL + indexablePages[i]; u.Loc != want {
			t.Errorf("адрес %d: %q, ожидался %q", i, u.Loc, want)
		}
	}
	// Страница инструментов собирается в браузере: роботу досталась бы пустая
	// оболочка, поэтому в карте сайта её быть не должно.
	if strings.Contains(rec.Body.String(), SiteURL+ToolsPath) {
		t.Errorf("в карте сайта есть %s, хотя её содержимое строится скриптом", ToolsPath)
	}
}

// llms.txt — описание сервиса для ассистентов и AI-поиска: оно должно отдаваться
// как текст и вести на существующие страницы.
func TestLlmsTxt(t *testing.T) {
	rec := do(t, newTestServer(t), llmsPath)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s: код %d", llmsPath, rec.Code)
	}
	if ct := rec.Header().Get(echo.HeaderContentType); !strings.Contains(ct, "text/plain") {
		t.Errorf("GET %s: Content-Type %q, ожидался text/plain", llmsPath, ct)
	}
	body := rec.Body.String()
	if !strings.HasPrefix(body, "# Crypto Polymath") {
		t.Error("llms.txt должен начинаться с заголовка сервиса")
	}
	for _, page := range []string{SiteURL + "/", SiteURL + ToolsPath, SiteURL + DocsPath} {
		if !strings.Contains(body, page) {
			t.Errorf("в llms.txt нет ссылки на %s", page)
		}
	}
	for _, exchange := range landingExchanges {
		if !strings.Contains(body, exchange) {
			t.Errorf("в llms.txt не упомянута биржа %q", exchange)
		}
	}
	// Ассистенты подсказывают запросы к API, поэтому лимиты должны быть названы
	// здесь же: без них совет «дёргайте /api/v1» приведёт пользователя к 429.
	for fragment, what := range map[string]string{
		"10 requests per minute":      "анонимная квота",
		"X-Token":                     "заголовок с токеном",
		"429":                         "код ответа при превышении лимита",
		"Interface language: English": "язык интерфейса",
	} {
		if !strings.Contains(body, fragment) {
			t.Errorf("в llms.txt не указан %s (%q)", what, fragment)
		}
	}
}

// Канонический адрес продублирован в Go и в разметке — расхождение увело бы
// краулер на несуществующий хост, поэтому сверяем копии. Публичный домен в
// учебных материалах не держим: в SiteURL только локальный origin.
func TestCanonicalHostIsConsistent(t *testing.T) {
	hostRe := regexp.MustCompile(`https?://[a-z0-9.-]+`)

	sources := map[string]string{
		"index.html": string(IndexPage),
		"docs.html":  string(docsPage),
		"llms.txt":   string(llmsFile),
	}
	for name, content := range sources {
		if strings.Contains(content, "cryptopolymath.org") {
			t.Errorf("%s всё ещё содержит публичный хост cryptopolymath.org", name)
		}
		for _, found := range hostRe.FindAllString(content, -1) {
			// Внешние адреса (CDN, schema.org) под это правило не попадают.
			if strings.HasPrefix(found, SiteURL) && found != SiteURL {
				t.Errorf("%s ссылается на %q, а канонический адрес — %q", name, found, SiteURL)
			}
		}
	}
}

// Страница инструментов не должна попадать в индекс: без исполнения скриптов
// поисковик увидит пустую оболочку и посчитает страницу малополезной.
func TestToolsPageIsNotIndexable(t *testing.T) {
	if !strings.Contains(string(toolsPage), `content="noindex"`) {
		t.Errorf("на %s нет noindex", ToolsPath)
	}
}

// Иконка приложения. Основной формат — вектор: во вкладке браузера знаку
// достаётся 16 пикселей, где растр мылится. PNG нужны как запас для браузеров
// без SVG и для закладки на домашний экран iOS.
var iconAssets = []string{"icon.svg", "icon-32.png", "apple-touch-icon.png"}

// Знаки — исключение из правила WebP ниже: это векторы и мелкий растр под них, а
// не иллюстрации, для которых WebP и заведён.
var markAssets = append([]string{"almamaro.svg"}, iconAssets...)

// Иллюстрации лендинга дают основной вес страницы, поэтому держим их в WebP и
// следим, чтобы в бинарник не попали тяжёлые исходники.
func TestLandingImagesAreOptimized(t *testing.T) {
	const maxAssetSize = 200 << 10

	pages, err := NewPages()
	if err != nil {
		t.Fatalf("NewPages: %v", err)
	}
	entries, err := fs.ReadDir(pages.assets, "img")
	if err != nil {
		t.Fatalf("чтение каталога картинок: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("каталог картинок пуст")
	}
	marks := make(map[string]bool, len(markAssets))
	for _, name := range markAssets {
		marks[name] = true
	}
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			t.Fatalf("%s: %v", entry.Name(), err)
		}
		if !marks[entry.Name()] && !strings.HasSuffix(entry.Name(), ".webp") {
			t.Errorf("%s: ожидался формат WebP", entry.Name())
		}
		if info.Size() > maxAssetSize {
			t.Errorf("%s: %d байт, больше лимита %d", entry.Name(), info.Size(), maxAssetSize)
		}
	}
}

// Иконка приложения должна быть встроена во всех трёх вариантах: браузер сам
// выбирает подходящий, а отсутствие файла даёт битую картинку во вкладке.
func TestIconAssetsExist(t *testing.T) {
	pages, err := NewPages()
	if err != nil {
		t.Fatalf("NewPages: %v", err)
	}
	for _, name := range iconAssets {
		info, err := fs.Stat(pages.assets, filepath.Join("img", name))
		if err != nil {
			t.Errorf("иконка %s не встроена: %v", name, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("иконка %s пустая", name)
		}
	}
}

// Подпись разработчика в подвале: название компании, чем она занимается и её
// знак. Проверяем целиком, потому что выходные данные легко потерять при правке
// подвала, а заметить это некому — на них никто не нажимает.
func TestLandingCreditsDeveloper(t *testing.T) {
	body := string(IndexPage)
	required := map[string]string{
		"Built by <strong>AlmaMaro</strong>": "упоминание разработчика",
		">tech<":                             "вторая часть названия",
		"API bots and AI bots":               "чем занимается компания",
		StaticPrefix + "/img/almamaro.svg":   "знак разработчика",
		"fintech and crypto solutions":       "область работы компании",
	}
	for fragment, what := range required {
		if !strings.Contains(body, fragment) {
			t.Errorf("в подвале лендинга нет %s (%q)", what, fragment)
		}
	}
}

// Каждая страница должна объявлять иконку сама: без тега link браузер берёт
// /favicon.ico, а это лишний запрос и, в отличие от SVG, всего один размер.
func TestPagesDeclareIcon(t *testing.T) {
	declarations := []string{
		`<link rel="icon" type="image/svg+xml" href="` + StaticPrefix + `/img/icon.svg">`,
		`<link rel="icon" type="image/png" sizes="32x32" href="` + StaticPrefix + `/img/icon-32.png">`,
		`<link rel="apple-touch-icon" href="` + StaticPrefix + `/img/apple-touch-icon.png">`,
	}
	for name, page := range map[string][]byte{
		"index.html": IndexPage,
		"tools.html": toolsPage,
		"docs.html":  docsPage,
	} {
		body := string(page)
		for _, want := range declarations {
			if !strings.Contains(body, want) {
				t.Errorf("%s не объявляет иконку: нет %s", name, want)
			}
		}
		// Знак рядом с названием в шапке — тот же файл, что и во вкладке.
		if !strings.Contains(body, `<img src="`+StaticPrefix+`/img/icon.svg"`) {
			t.Errorf("%s: в шапке рядом с названием нет знака приложения", name)
		}
	}
}

// Название сервиса на виду у посетителя всегда идёт со знаком. Мест таких три:
// шапка (её проверяет TestPagesDeclareIcon), строка в подвале лендинга и
// заголовок документации, который Redoc рисует из контракта по info.x-logo.
func TestBrandNameAlwaysHasIcon(t *testing.T) {
	footer := regexp.MustCompile(`(?s)<footer>.*</footer>`).FindString(string(IndexPage))
	if footer == "" {
		t.Fatal("на лендинге не найден подвал")
	}
	brand := regexp.MustCompile(`(?s)<span class="footer-brand">(.*?)</span>`).FindStringSubmatch(footer)
	if brand == nil {
		t.Fatal("в подвале нет строки с названием сервиса")
	}
	if !strings.Contains(brand[1], StaticPrefix+"/img/icon.svg") {
		t.Error("в подвале название сервиса идёт без знака")
	}
	if !strings.Contains(brand[1], "Crypto Polymath") {
		t.Error("строка с классом footer-brand не содержит названия сервиса")
	}

	// Redoc берёт знак из расширения x-logo в info. Ссылка относительная,
	// поэтому проверяем, что файл действительно раздаётся.
	logoRe := regexp.MustCompile(`x-logo:\s*\n\s+url:\s*(\S+)`)
	logo := logoRe.FindSubmatch(openapi.Specification)
	if logo == nil {
		t.Fatal("в контракте нет x-logo: на странице документации название останется без знака")
	}
	if rec := do(t, newTestServer(t), string(logo[1])); rec.Code != http.StatusOK {
		t.Errorf("x-logo ссылается на %s, сервер ответил %d", logo[1], rec.Code)
	}
}

// Иконка отдаётся из встроенной статики. Раньше этот путь вёл на файл на диске
// относительно рабочего каталога, поэтому вне корня репозитория отдавалась 404.
func TestFaviconRoute(t *testing.T) {
	pages, err := NewPages()
	if err != nil {
		t.Fatalf("NewPages: %v", err)
	}
	rec := do(t, newTestServer(t), faviconPath)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s: код %d, ожидался %d", faviconPath, rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get(echo.HeaderContentType); !strings.Contains(ct, "image/") {
		t.Errorf("GET %s: Content-Type %q, ожидалась картинка", faviconPath, ct)
	}
	want, err := fs.ReadFile(pages.assets, faviconAsset)
	if err != nil {
		t.Fatalf("чтение %s: %v", faviconAsset, err)
	}
	if !bytes.Equal(rec.Body.Bytes(), want) {
		t.Errorf("GET %s: отдано %d байт, в статике %d", faviconPath, rec.Body.Len(), len(want))
	}
}

// Скрипт на лендинге ровно один — контактная форма, и это единственная причина,
// по которой он там допустим: окно с формой без JavaScript не открыть.
//
// Всё остальное на странице остаётся статической разметкой. Ограничение не
// формальное: поисковые роботы и особенно AI-краулеры JavaScript обычно не
// исполняют, поэтому содержимое, собранное скриптом, для них не существует.
// Индикатор доступности сервиса убрали именно поэтому — заодно страница перестала
// зависеть от ответа API.
func TestLandingScriptsAreLimitedToContactForm(t *testing.T) {
	body := string(IndexPage)

	// Структурированные данные лежат в script с типом ld+json — это данные, а не
	// исполняемый код, поэтому их из проверки исключаем.
	scripts := regexp.MustCompile(`<script(?:\s[^>]*)?>`).FindAllString(body, -1)
	executable := make([]string, 0, len(scripts))
	for _, tag := range scripts {
		if strings.Contains(tag, "application/ld+json") {
			continue
		}
		executable = append(executable, tag)
	}
	if len(executable) != 1 {
		t.Fatalf("на лендинге %d исполняемых скриптов, ожидается один (форма): %v", len(executable), executable)
	}
	if !strings.Contains(executable[0], "/static/js/contact.js") {
		t.Errorf("единственный скрипт лендинга должен быть формой обратной связи, получено %s", executable[0])
	}
	for _, gone := range []string{"api-status-dot", "api-status-text", "Сервис работает", "js/landing.js"} {
		if strings.Contains(body, gone) {
			t.Errorf("на лендинге остался индикатор доступности (%q)", gone)
		}
	}
}

// Форма обратной связи: кнопка на странице, окно с тремя полями и двумя
// действиями. Проверяем разметку, а не скрипт: если поле или кнопка потеряются,
// форму нельзя будет ни заполнить, ни отменить, а скрипт при этом продолжит
// работать молча.
func TestLandingContactForm(t *testing.T) {
	body := string(IndexPage)

	for fragment, what := range map[string]string{
		`data-contact-open`:           "кнопка открытия формы",
		`<dialog id="contact-dialog"`: "окно формы",
		`action="/api/v1/contact"`:    "адрес отправки",
		`name="email"`:                "поле почты",
		`name="subject"`:              "поле темы",
		`name="message"`:              "поле сообщения",
		`type="submit"`:               "кнопка отправки",
		`data-contact-close`:          "кнопка отмены",
		`data-contact-status`:         "место для ответа сервера",
	} {
		if !strings.Contains(body, fragment) {
			t.Errorf("на лендинге нет %s (%q)", what, fragment)
		}
	}

	// Обязательность полей задана разметкой: браузер проверяет их до запроса, и
	// посетитель узнаёт о пустом поле сразу, а не после отказа сервера.
	if n := strings.Count(body, "required"); n < 3 {
		t.Errorf("обязательными объявлены не все три поля формы (%d)", n)
	}
	if !strings.Contains(body, `type="email"`) {
		t.Error("поле почты не объявлено type=email: браузер не проверит адрес до отправки")
	}
}

// Пределы длин в форме и в контракте — одни и те же числа. Разойдясь, они дают
// худший из возможных исходов: браузер принимает текст, сервер его отвергает, а
// посетитель видит отказ уже после нажатия «Send».
func TestContactFormLimitsMatchContract(t *testing.T) {
	body := string(IndexPage)
	spec := string(openapi.Specification)

	contract := regexp.MustCompile(`(?s)ContactMessageRequest:.*?ContactMessageResponse:`).FindString(spec)
	if contract == "" {
		t.Fatal("в контракте нет схемы ContactMessageRequest")
	}

	for _, limit := range []struct {
		attr string
		key  string
	}{
		{attr: `maxlength="254"`, key: "maxLength: 254"},
		{attr: `minlength="3"`, key: "minLength: 3"},
		{attr: `maxlength="150"`, key: "maxLength: 150"},
		{attr: `minlength="10"`, key: "minLength: 10"},
		{attr: `maxlength="5000"`, key: "maxLength: 5000"},
	} {
		if !strings.Contains(contract, limit.key) {
			t.Errorf("в контракте нет ограничения %q", limit.key)
		}
		if !strings.Contains(body, limit.attr) {
			t.Errorf("в форме нет ограничения %q, хотя контракт его требует", limit.attr)
		}
	}
}

// Без JavaScript кнопка формы ничего не делает, поэтому разметка подменяет её
// почтовой ссылкой. Подмена живёт в noscript: браузер разбирает его только при
// отключённых скриптах, поэтому обе кнопки одновременно не появляются.
func TestContactFormHasNoScriptFallback(t *testing.T) {
	body := string(IndexPage)

	noscript := regexp.MustCompile(`(?s)<noscript>.*?</noscript>`).FindString(body)
	if noscript == "" {
		t.Fatal("на лендинге нет noscript-подмены для формы обратной связи")
	}
	for _, rule := range []string{"[data-contact-open]", "display: none", ".contact-fallback"} {
		if !strings.Contains(noscript, rule) {
			t.Errorf("в noscript нет правила %q: без скриптов останется кнопка, которая ничего не делает", rule)
		}
	}
	if !strings.Contains(body, `class="btn btn-outline contact-fallback" href="mailto:`) {
		t.Error("нет почтовой ссылки на замену кнопке: без скриптов связаться будет нечем")
	}
	if !strings.Contains(readAsset(t, "css/app.css"), ".contact-fallback { display: none; }") {
		t.Error("почтовая ссылка не скрыта по умолчанию: со скриптами покажутся обе кнопки")
	}
}

// Форма обращается ровно к тому адресу, который объявлен в контракте, и ждёт 202:
// ответ 200 означал бы «обработали», а сервис обещает только «приняли».
func TestContactScriptMatchesContract(t *testing.T) {
	script := readAsset(t, "js/contact.js")

	if !strings.Contains(script, "'/api/v1/contact'") {
		t.Error("скрипт формы обращается не к /api/v1/contact")
	}
	if !strings.Contains(string(openapi.Specification), "  /contact:") {
		t.Error("в контракте нет пути /contact, к которому обращается форма")
	}
	// Отказ по лимиту объясняется отдельно: у формы тот же анонимный лимит, что у
	// остальных вызовов, и «попробуйте позже» без причины выглядит как поломка.
	if !strings.Contains(script, "429") || !strings.Contains(script, "Retry-After") {
		t.Error("скрипт формы не разбирает отказ по лимиту запросов")
	}
}

// PrimeVue общается с Vue через provide/inject, поэтому оба должны получить один
// и тот же экземпляр рантайма. Версия зафиксирована в importmap дважды — здесь
// проверяем, что при обновлении её поправили в обоих местах.
func TestToolsPagePinsSameVueVersion(t *testing.T) {
	body := string(toolsPage)

	vueRe := regexp.MustCompile(`esm\.sh/vue@(\d+\.\d+\.\d+)`)
	vueMatch := vueRe.FindStringSubmatch(body)
	if vueMatch == nil {
		t.Fatal(`в importmap не найден закреплённый адрес "vue" вида esm.sh/vue@x.y.z`)
	}

	depsRe := regexp.MustCompile(`primevue@\d+\.\d+\.\d+/[a-z]+\?deps=vue@(\d+\.\d+\.\d+)`)
	depsMatches := depsRe.FindAllStringSubmatch(body, -1)
	if len(depsMatches) == 0 {
		t.Fatal("в importmap не найдено ни одного модуля PrimeVue с закреплённой версией vue")
	}
	for _, m := range depsMatches {
		if m[1] != vueMatch[1] {
			t.Errorf("PrimeVue собран против vue@%s, а importmap отдаёт vue@%s", m[1], vueMatch[1])
		}
	}
}

// Кнопки «Попробовать» на лендинге ведут на /tools?fn=<key> и должны попадать в
// существующую функцию: при переименовании ключа в endpoints.js ссылка молча
// открывала бы страницу с функцией по умолчанию.
func TestLandingDeepLinksResolve(t *testing.T) {
	pages, err := NewPages()
	if err != nil {
		t.Fatalf("NewPages: %v", err)
	}
	definitions, err := fs.ReadFile(pages.assets, "js/endpoints.js")
	if err != nil {
		t.Fatalf("чтение описаний функций: %v", err)
	}

	linkRe := regexp.MustCompile(regexp.QuoteMeta(ToolsPath) + `\?fn=([a-z0-9-]+)`)
	links := linkRe.FindAllStringSubmatch(string(IndexPage), -1)
	if len(links) == 0 {
		t.Fatal("на лендинге нет ни одной ссылки на конкретную функцию")
	}
	for _, m := range links {
		if !strings.Contains(string(definitions), "key: '"+m[1]+"'") {
			t.Errorf("лендинг ссылается на функцию %q, которой нет в endpoints.js", m[1])
		}
	}
}

// Разбирает описания функций из endpoints.js в карту «ключ функции -> её блок».
// Полноценный парсер JS для этого не нужен: каждое описание начинается с
// endpoint({, и внутри блока лежит и ключ, и все остальные поля.
func endpointBlocks(t *testing.T) map[string]string {
	t.Helper()
	definitions := readAsset(t, "js/endpoints.js")
	keyRe := regexp.MustCompile(`key: '([a-z0-9-]+)'`)

	blocks := make(map[string]string)
	for _, chunk := range strings.Split(definitions, "endpoint({")[1:] {
		if key := keyRe.FindStringSubmatch(chunk); key != nil {
			blocks[key[1]] = chunk
		}
	}
	if len(blocks) == 0 {
		t.Fatal("в endpoints.js не найдено ни одного описания функции")
	}
	return blocks
}

// У свечей и индикаторов под таблицей значений рисуется график, поэтому у этих
// функций он должен быть описан. Без описания результат молча остался бы одной
// таблицей — сломанного вида на странице не появилось бы, и заметить это можно
// было бы только глазами.
func TestSeriesFunctionsHaveChart(t *testing.T) {
	blocks := endpointBlocks(t)
	for _, key := range []string{"candlestick", "candle-indicator", "indicator", "analysis"} {
		block, ok := blocks[key]
		if !ok {
			t.Errorf("в endpoints.js нет функции %q", key)
			continue
		}
		if !strings.Contains(block, "chart:") {
			t.Errorf("у функции %q не описан график под таблицей", key)
		}
	}
}

// Сравнение цен по биржам показывается особым видом: сводка с минимумом,
// медианой и максимумом, подсветка крайних значений и график.
func TestPricesFunctionUsesComparisonView(t *testing.T) {
	block, ok := endpointBlocks(t)["prices-symbol"]
	if !ok {
		t.Fatal("в endpoints.js нет функции prices-symbol")
	}
	if !strings.Contains(block, `view: 'prices'`) {
		t.Error("у функции «One symbol on all exchanges» не задан вид сравнения цен")
	}

	view := readAsset(t, "js/prices-view.js")
	for _, want := range []string{"priceStats", "buildPricesChartData", "row-min", "row-max"} {
		if !strings.Contains(view, want) {
			t.Errorf("в сравнении цен нет %q", want)
		}
	}
	// Медиана — единственный из трёх показателей, которого нет в ответе API:
	// он считается на стороне страницы, поэтому проверяем расчёт отдельно.
	if !strings.Contains(readAsset(t, "js/format.js"), "median:") {
		t.Error("медианная цена не вычисляется в format.js")
	}
}

// Поля, по которым строятся графики, берутся из ответа API. Если их
// переименуют в контракте, график молча станет пустым, поэтому сверяем имена из
// описаний функций с openapi.yaml.
func TestChartFieldsExistInSpec(t *testing.T) {
	definitions := readAsset(t, "js/endpoints.js")
	specRe := regexp.MustCompile(`\{ x: '([a-z_]+)', y: '([a-z_]+)', label:`)
	specs := specRe.FindAllStringSubmatch(definitions, -1)
	if len(specs) == 0 {
		t.Fatal("в endpoints.js не найдено ни одного описания графика")
	}
	for _, spec := range specs {
		for _, field := range spec[1:] {
			fieldRe := regexp.MustCompile(`(?m)^\s+` + field + `:$`)
			if !fieldRe.Match(openapi.Specification) {
				t.Errorf("график строится по полю %q, которого нет в openapi.yaml", field)
			}
		}
	}
}

// Адрес эндпоинта и HTTP-метод посетителю ничего не объясняют: он выбирает
// функцию по названию, а контракт живёт в документации на /docs/api. Метод и
// путь по-прежнему нужны для отправки запроса — проверяем, что их не вернули на
// страницу как элементы разметки.
func TestToolsPageHidesRequestDetails(t *testing.T) {
	displays := map[string]string{
		"'fn-path'":        "блок с адресом эндпоинта",
		"value: ep.method": "значок с HTTP-методом",
	}
	script := readAsset(t, "js/tools.js")
	for fragment, what := range displays {
		if strings.Contains(script, fragment) {
			t.Errorf("на странице инструментов снова показывается %s (%q)", what, fragment)
		}
	}
	if strings.Contains(readAsset(t, "css/tools.css"), ".fn-path") {
		t.Error("в стилях остался класс .fn-path — блок с адресом эндпоинта")
	}
}

// API ограничивает анонимные запросы, и страница инструментов живёт по тому же
// правилу: она обращается к API из браузера без токена. Значит отказ по лимиту —
// обычное состояние страницы, а не аварийное, и он должен быть разобран отдельно
// от «API недоступен»: причина у них разная, и пользователю нужно знать, что
// достаточно подождать. Само правило названо в шапке страницы, до отказа.
func TestToolsPageExplainsRateLimit(t *testing.T) {
	script := readAsset(t, "js/tools.js")

	for fragment, what := range map[string]string{
		"429":                       "проверка кода отказа по лимиту",
		"Retry-After":               "чтение времени ожидания из ответа",
		"Rate limited":              "отдельный значок состояния вместо «API недоступен»",
		"10 requests per minute":    "анонимная квота названа в шапке страницы",
		"quota of the pricing plan": "правило лимитов для владельцев токена",
	} {
		if !strings.Contains(script, fragment) {
			t.Errorf("на странице инструментов нет %s (%q)", what, fragment)
		}
	}
}

// Все модули, которые импортирует код страницы, должны быть объявлены в
// importmap: браузер не умеет разрешать голые имена сам.
func TestToolsImportMapCoversModules(t *testing.T) {
	pages, err := NewPages()
	if err != nil {
		t.Fatalf("NewPages: %v", err)
	}
	importMap := string(toolsPage)
	importRe := regexp.MustCompile(`from '((?:vue|primevue/)[a-z/]*)'`)

	for _, script := range []string{"js/tools.js", "js/data-view.js", "js/prices-view.js"} {
		f, err := pages.assets.Open(script)
		if err != nil {
			t.Fatalf("открытие %s: %v", script, err)
		}
		content, err := io.ReadAll(f)
		_ = f.Close()
		if err != nil {
			t.Fatalf("чтение %s: %v", script, err)
		}
		for _, m := range importRe.FindAllStringSubmatch(string(content), -1) {
			if !strings.Contains(importMap, `"`+m[1]+`"`) {
				t.Errorf("%s импортирует %q, но модуля нет в importmap страницы инструментов", script, m[1])
			}
		}
	}
}
