# language: ru
@trading @stop-loss
Функционал: Динамический Stop-Loss
  Как трейдер
  Я хочу, чтобы стоп-лосс адаптировался к волатильности рынка
  Чтобы избежать ложных срабатываний в волатильных условиях

  # -------------------------------------------------------------------
  # Контекст:
  # Фиксированный стоп-лосс (например, -5% всегда) плохо работает:
  # - На волатильном рынке выбивает слишком часто (ложные сработки)
  # - На спокойном рынке слишком далеко (большие потери при реальном развороте)
  #
  # Решение: динамический SL, который масштабируется от текущей волатильности.
  # Дополнительно — trailing stop, который «подтягивается» за ценой в прибыли
  # по Heiken Ashi свечам (сглаженные, фильтруют шум).
  #
  # Бизнес-задача: автоматический risk management без ручного вмешательства.
  # -------------------------------------------------------------------

  Контекст:
    Допустим leverage is 10x
    И default stop-loss coefficients

  Правило: SL и трейлинг пропорциональны волатильности

    # Бизнес-идея: при умеренной волатильности (2%) стоп-лосс и порог трейлинга
    # устанавливаются на 8% от цены (коэффициент 4×).
    # Трейдеру показываются:
    # - Цена стопа (ниже входа для лонга)
    # - Цена активации трейлинга (выше входа для лонга)
    # - Потери маржи при стопе / прибыль маржи при активации трейлинга

    Сценарий: Умеренная волатильность
      Допустим volatility RangePct is 2.0%
      Когда I calculate dynamic stop-loss for a long at 100000
      Тогда SL price percent should be 8.0%
      И trail activation percent should be 8.0%
      И initial stop price should be below 100000
      И trail activation price should be above 100000
      И SL margin loss should be -80%
      И trail margin profit should be 80%

    # Бизнес-идея: для шорта всё зеркально — стоп выше входа, трейлинг ниже.

    Сценарий: Шорт — стоп выше входа
      Допустим volatility RangePct is 2.0%
      Когда I calculate dynamic stop-loss for a short at 100000
      Тогда initial stop price should be above 100000
      И trail activation price should be below 100000

  Правило: Минимальный и максимальный пороги (floor/cap)

    # Бизнес-идея: даже на очень спокойном рынке стоп не должен быть слишком
    # близким (floor = 3%), а на хаотичном — слишком далёким (cap = 15%).
    # Это защищает от edge-кейсов и гарантирует разумный диапазон.

    Сценарий: Низкая волатильность — зажимается до floor
      Допустим volatility RangePct is 0.1%
      Когда I calculate dynamic stop-loss for a long at 100000
      Тогда SL price percent should be 3.0%
      И trail activation percent should be 2.0%

    Сценарий: Высокая волатильность — зажимается до cap
      Допустим volatility RangePct is 10.0%
      Когда I calculate dynamic stop-loss for a long at 100000
      Тогда SL price percent should be 15.0%
      И trail activation percent should be 20.0%

  Правило: Trailing stop двигается только в направлении прибыли

    # Бизнес-идея: trailing stop «подтягивается» за ценой, когда позиция
    # в прибыли. Для лонга стоп двигается вверх (за low HA-свечи),
    # для шорта — вниз (за high HA-свечи).
    #
    # Ключевые свойства:
    # 1. Стоп НИКОГДА не опускается ниже текущего значения (ratchet)
    # 2. Стоп НИКОГДА не опускается ниже цены входа (пол)
    # 3. До активации трейлинга стоп не двигается

    Сценарий: Лонг — стоп двигается вверх за HA low
      Допустим a long entry at 100 with current stop at 98
      И trailing is activated
      Когда HA candle has low 105 and high 115
      Тогда the stop should be 105

    Сценарий: Лонг — стоп не опускается ниже текущего
      Допустим a long entry at 100 with current stop at 103
      И trailing is activated
      Когда HA candle has low 101 and high 115
      Тогда the stop should be 103

    Сценарий: Лонг — цена входа как минимальный порог
      Допустим a long entry at 100 with current stop at 95
      И trailing is activated
      Когда HA candle has low 92 and high 110
      Тогда the stop should be 100

    Сценарий: Шорт — стоп двигается вниз за HA high
      Допустим a short entry at 100 with current stop at 103
      И trailing is activated
      Когда HA candle has low 90 and high 98
      Тогда the stop should be 98

    Сценарий: До активации стоп не меняется
      Допустим a long entry at 100 with current stop at 95
      И trailing is not activated
      Когда HA candle has low 105 and high 115
      Тогда the stop should be 95

  Правило: Определение срабатывания стопа

    # Бизнес-идея: стоп считается сработавшим, если свеча «прошла» через
    # уровень стопа. Для лонга — low ≤ stop, для шорта — high ≥ stop.

    Структура сценария: Определение касания стопа
      Допустим a <side> stop at <stop>
      Когда candle has low <low> and high <high>
      Тогда stop hit should be <hit>

      Примеры:
        | side  | stop | low | high | hit   |
        | long  | 100  | 99  | 110  | true  |
        | long  | 100  | 100 | 110  | true  |
        | long  | 100  | 101 | 110  | false |
        | short | 100  | 90  | 100  | true  |
        | short | 100  | 90  | 105  | true  |
        | short | 100  | 90  | 99   | false |

  Правило: Цена выхода с учётом гэпа

    # Бизнес-идея: если рынок открылся с гэпом (разрыв через стоп),
    # реальная цена выхода — не стоп, а цена открытия (хуже стопа).
    # Если гэпа нет — выходим ровно по стопу.

    Структура сценария: Цена выхода при гэпе
      Допустим a <side> stop at <stop>
      Когда candle opens at <open>
      Тогда exit price should be <exit>

      Примеры:
        | side  | stop | open | exit |
        | long  | 100  | 105  | 100  |
        | long  | 100  | 95   | 95   |
        | short | 100  | 95   | 100  |
        | short | 100  | 105  | 105  |
