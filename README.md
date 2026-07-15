# Crypto polymath (cryptopolymath.org / 167.235.194.60)

![]() <img src="./img/logo.png"  width="40%">

Рассчитываем основные криптовалютные показатели, такие как:
- цена
- объем торгов
- рыночная капитализация
- индикаторы

- [github](https://github.com/AlekseyPorandaykin/crypto_polymath)
- [sentry](https://test-p43.sentry.io/projects/crypto-polymath/)



## Техническое описание
- Вся логика описана в директории `core/`. Там описана только логика без привязки к реализации.
- В `internal/` находятся логика запуска разных доменов и при каких условиях
- Подробный обзор архитектуры: [docs/overview.md](./docs/overview.md)

## Тесты и benchmarks

```bash
make test-core      # unit-тесты core/
make bench-core     # performance-тесты core/
```

## Экземпляры запуска
- loader  `sh ./bin/crypto_polemath daemon loader` - запуск загрузки данных из разных исчтоников в бд. 

## Список features
- Цены
  - /prices/exchange/{exchange} - все цены по бирже
- Свечи (/api/v1/candlestick/ХЪ/BTCUSDT/H/1)
