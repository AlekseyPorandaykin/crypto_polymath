# Crypto polymath 

![] <img src="./img/logo.png"  width="40%">

Учебный некоммерческий проект: проверка теории и знаний из курсов на больших объёмах
данных, а также опробование разных технологий и подходов. Использует только публичную
информацию (открытые рыночные данные бирж). Не предназначен для коммерческой торговли
и не работает с приватными ключами пользователей.

Рассчитываем основные криптовалютные показатели, такие как:
- цена
- объем торгов
- рыночная капитализация
- индикаторы



## Техническое описание
- Вся логика описана в директории `core/`. Там описана только логика без привязки к реализации.
- В `internal/` находятся логика запуска разных доменов и при каких условиях
- Подробный обзор архитектуры: [docs/overview.md](./docs/overview.md)
- Контекст и правила разработки для AI-ассистентов (Cursor, Codex, Claude): [.ai/](./.ai/README.md).
  Там же журнал изменений — [.ai/journal.md](./.ai/journal.md)

## Тесты и benchmarks

### Обзор тестовой стратегии

Проект использует многоуровневую стратегию тестирования для обеспечения корректности, стабильности и производительности кода.

| Тип тестов | Команда | Что проверяет |
|------------|---------|---------------|
| Unit | `make test-unit` | Корректность отдельных функций |
| BDD (Cucumber) | `make test-bdd` | Бизнес-сценарии на языке Gherkin |
| Fuzz | `make test-fuzz` | Генеративные тесты на случайных данных |
| Golden | `make test-golden-update` | Стабильность вывода (защита от регрессий) |
| Race | `make test-race` | Конкурентность (data races) |
| Bench | `make bench-core` | Производительность и аллокации |
| Acceptance | `make test-acceptance` | End-to-end через HTTP API |
| Smoke | `make test-smoke` | Быстрая проверка «живости» сервера |
| Torture | `make test-torture` | Стресс-тесты (10k горутин, memory leaks) |

---

### Unit-тесты (`make test-unit`)

**Проблема:** ошибка в формуле расчёта средней цены или ликвидации ведёт к неправильному отображению позиции и потенциальным финансовым потерям.

**Решение:** table-driven тесты проверяют каждую функцию на известных входах/выходах. Покрывают:
- `core/trading` — средняя цена, ликвидация, PnL, SimulateAddOn, RiskAtPrice, stop-loss
- `core/indicator/calculator` — MA, EMA, Trend, Stochastic, PriceChanges, Volatility
- `core/analysis/calculators` — MACD, RSI, TrendByMA, helper-функции (calcTrend, lenBatch)
- `core/candle_indicator` — Heiken Ashi, методы Indicator (SizeBody, Direction)
- `core/exchange` — загрузка символов, поиск, категории
- `domain` — EMA, ATR, Consolidation, последовательности свечей

```bash
make test-unit
make test-core       # только core/
```

---

### BDD-тесты — Cucumber/Godog (`make test-bdd`)

**Проблема:** разработчики думают о коде, а бизнес — о поведении. Сложно убедиться, что код реализует именно то, что ожидает трейдер.

**Решение:** сценарии написаны на русском языке в формате Gherkin (`.feature` файлы). Читаемы без знания Go. Документируют бизнес-правила и служат «живой спецификацией».

**76 сценариев** в 6 feature-файлах:

| Feature | Пакет | Сценарии | Что проверяет |
|---------|-------|----------|---------------|
| `trading.feature` | `core/trading` | 16 | Средняя цена, ликвидация, PnL, объём, дистанция |
| `simulate_addon.feature` | `core/trading` | 6 | Усреднение позиции, оценка риска |
| `stop_loss.feature` | `core/trading` | 17 | Динамический SL, trailing, hit, exit price |
| `indicators.feature` | `core/indicator` | 12 | MA, EMA, Trend, Stochastic, Spot PnL |
| `heiken_ashi.feature` | `core/candle_indicator` | 11 | Формулы HA, свойства свечи, доджи |
| `analysis.feature` | `core/analysis` | 20 | Тренд, MACD, RSI, размер батча |

Пример сценария:
```gherkin
Сценарий: Лонг — усреднение вниз
  Допустим a long position at 100000 with volume 1 and margin 10000 and leverage 10x
  Когда I simulate add-on at price 90000 with margin 10000
  Тогда the new entry price should be below 100000
  И the liquidation price should move lower
```

```bash
make test-bdd
```

---

### Fuzz-тесты (`make test-fuzz`)

**Проблема:** ручные тесты не покрывают экстремальные входы: очень большие числа, граничные значения, комбинации, приводящие к делению на ноль или NaN.

**Решение:** Go fuzzer генерирует тысячи случайных комбинаций параметров и проверяет инварианты:
- `NewAvgEntryPrice`: результат ∈ [min(p1,p2), max(p1,p2)]
- `LiquidationPrice`: long_liq < entry < short_liq
- `UnrealizedPnL`: long + short = 0 (нулевая сумма)
- `MA`: результат ∈ [min(closes), max(closes)]
- `Stochastic`: результат ∈ [0, 100]

```bash
make test-fuzz           # 60 секунд
make test-fuzz-short     # 10 секунд
```

---

### Golden-тесты (`make test-golden-update`)

**Проблема:** рефакторинг может незаметно изменить точность вычислений или структуру результата. Обычные assert не поймают потерю 0.001 в 10-м знаке.

**Решение:** фиксируют полный JSON-вывод ключевых функций в файлах `testdata/*.golden.json`. При изменении — тест падает с diff.

```bash
# Проверка:
go test ./core/trading/ -run=Test.*_golden
# Обновление после осознанного изменения:
make test-golden-update
```

---

### Race-тесты (`make test-race`)

**Проблема:** калькулятор может вызываться параллельно из разных горутин (API handler, WebSocket, batch processing). Data race → undefined behavior.

**Решение:** запуск с `-race` детектором. Проверяет отсутствие конкурентных обращений к shared state.

```bash
make test-race
```

---

### Benchmarks (`make bench-core`)

**Проблема:** калькулятор в hot path (API, real-time UI). Рефакторинг может добавить лишние аллокации и замедлить ответ.

**Решение:** измеряют ns/op и allocs/op. При деградации — видно в CI.

```bash
make bench-core
```

---

### Acceptance-тесты (`make test-acceptance`)

**Проблема:** unit-тесты проверяют функции изолированно, но не гарантируют, что API работает end-to-end (routing → handler → calculator → response).

**Решение:** HTTP-запросы к реальному серверу, проверка status code и структуры JSON. Требуют запущенный docker-compose.

```bash
make test-acceptance
```

---

### Smoke-тесты (`make test-smoke`)

**Проблема:** после деплоя нужно быстро (< 30 сек) убедиться, что сервис жив.

**Решение:** минимальные проверки: TCP connect, HTTP 200, наличие ключевых полей в ответе, metrics endpoint.

```bash
make test-smoke
```

---

### Torture-тесты (`make test-torture`)

**Проблема:** обычные тесты не обнаруживают: memory leaks при длительной работе, паники при 10k concurrent requests, деградацию под нагрузкой.

**Решение:** запуск с `-race`, 10 000 горутин, проверка `runtime.MemStats`, concurrent API load.

```bash
make test-torture
```

---

### Запуск всех локальных тестов

```bash
make test-all    # unit + race + cover
```

## Экземпляры запуска
- loader  `sh ./bin/crypto_polemath daemon loader` - запуск загрузки данных из разных исчтоников в бд. 

## Список features
- Цены
  - /prices/exchange/{exchange} - все цены по бирже
- Свечи (/api/v1/candlestick/ХЪ/BTCUSDT/H/1)
