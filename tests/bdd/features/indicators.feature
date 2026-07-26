# language: ru
@indicators
Функционал: Технические индикаторы
  Как аналитик
  Я хочу рассчитывать технические индикаторы по свечным данным
  Чтобы определять тренды и рыночные условия

  # -------------------------------------------------------------------
  # Контекст:
  # Технические индикаторы — основа алгоритмического анализа рынка.
  # Каждый индикатор принимает массив свечей (OHLCV) и возвращает числовое
  # значение, характеризующее состояние рынка.
  #
  # Бизнес-задача: обеспечить корректность расчёта индикаторов, которые
  # используются далее в аналитике (MACD, RSI) и торговых стратегиях.
  # -------------------------------------------------------------------

  Правило: Скользящая средняя (MA)

    # Бизнес-идея: MA — простое арифметическое среднее цен закрытия.
    # Используется для определения общего направления тренда и уровней поддержки.
    # Свойства: MA константного ряда = значение; MA пустых данных = nil.

    Сценарий: MA постоянных цен равна значению
      Допустим candlesticks with close prices 100, 100, 100
      Когда I calculate MA
      Тогда the indicator value should be 100
      И the indicator name should be "MA"

    Сценарий: MA [100, 200, 300] примерно 200
      Допустим candlesticks with close prices 100, 200, 300
      Когда I calculate MA
      Тогда the indicator value should be approximately 200

    Сценарий: MA пустых данных не возвращает результат
      Допустим no candlestick data
      Когда I calculate MA
      Тогда no indicator should be returned

  Правило: Экспоненциальная скользящая средняя (EMA)

    # Бизнес-идея: EMA придаёт больший вес последним ценам, быстрее
    # реагирует на изменения. Используется в MACD, стратегиях пересечения.
    # Свойства: EMA константного ряда = значение; при тренде вверх EMA
    # выше середины, но ниже последней цены (отставание).

    Сценарий: EMA постоянного ряда равна значению
      Допустим candlesticks with close prices 100, 100, 100, 100, 100
      Когда I calculate EMA
      Тогда the indicator value should be approximately 100
      И the indicator name should be "EMA"

    Сценарий: EMA восходящего ряда — выше середины, ниже последней
      Допустим candlesticks with close prices 10, 20, 30, 40, 50
      Когда I calculate EMA
      Тогда the indicator value should be above 30
      И the indicator value should be below 50

  Правило: Определение тренда

    # Бизнес-идея: Trend определяет направление движения цены:
    # +1 (восходящий), -1 (нисходящий), 0 (боковой/flat).
    # Используется для фильтрации сигналов: не входить против тренда.

    Сценарий: Растущие цены дают восходящий тренд
      Допустим 20 candlesticks with prices rising from 100 by 10
      Когда I calculate Trend
      Тогда the indicator value should be 1

    Сценарий: Падающие цены дают нисходящий тренд
      Допустим 20 candlesticks with prices falling from 300 by 10
      Когда I calculate Trend
      Тогда the indicator value should be -1

    Сценарий: Плоские цены дают боковой тренд
      Допустим 20 candlesticks with constant price 100
      Когда I calculate Trend
      Тогда the indicator value should be 0

  Правило: Стохастический осциллятор

    # Бизнес-идея: Stochastic показывает положение текущей цены относительно
    # диапазона (min–max) за период. Значения 0–100:
    # - 100 = цена на максимуме (перекупленность)
    # - 0 = цена на минимуме (перепроданность)
    # Используется для поиска разворотных точек.

    Сценарий: На максимуме стохастик = 100
      Допустим candlesticks with close prices 50, 60, 70, 80, 100
      Когда I calculate Stochastic
      Тогда the indicator value should be 100

    Сценарий: На минимуме стохастик = 0
      Допустим candlesticks with close prices 100, 90, 80, 70, 50
      Когда I calculate Stochastic
      Тогда the indicator value should be 0

    Сценарий: Все цены одинаковые — стохастик 0
      Допустим candlesticks with close prices 100, 100, 100
      Когда I calculate Stochastic
      Тогда the indicator value should be 0

  Правило: Спот PnL

    # Бизнес-идея: расчёт прибыли/убытка для спотовой позиции (без плеча).
    # PnL = Volume × (MarkPrice − EntryPrice)
    # Percent = (MarkPrice − EntryPrice) / EntryPrice × 100

    Структура сценария: Прибыль и убыток на споте
      Допустим spot volume <volume> at entry <entry>
      Когда market is at <mark>
      Тогда spot PnL value should be <pnl>
      И spot PnL percent should be <percent>%

      Примеры:
        | volume | entry | mark | pnl  | percent |
        | 2      | 100   | 150  | 100  | 50      |
        | 2      | 100   | 80   | -40  | -20     |
        | 1      | 100   | 100  | 0    | 0       |
