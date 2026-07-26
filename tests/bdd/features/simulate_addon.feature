# language: ru
@trading @addon
Функционал: Симуляция докупки (Add-On)
  Как трейдер
  Я хочу смоделировать добавление к позиции
  Чтобы заранее увидеть влияние на среднюю цену, маржу и ликвидацию

  # -------------------------------------------------------------------
  # Контекст:
  # Докупка (усреднение) — одна из ключевых стратегий управления позицией.
  # Трейдер добавляет объём к убыточной (или прибыльной) позиции, чтобы:
  # - Сместить среднюю цену входа ближе к рынку
  # - Увеличить потенциальную прибыль при возврате цены
  # - Но при этом увеличивается риск (больше объём = больше убыток при дальнейшем движении)
  #
  # Бизнес-задача: перед реальной докупкой трейдер должен видеть полную
  # картину — как изменятся все метрики позиции.
  # -------------------------------------------------------------------

  Контекст:
    Допустим MMR is 0.5%

  Правило: Усреднение вниз снижает среднюю цену для лонга

    # Бизнес-идея: классическое «усреднение вниз» — докупка лонга по более
    # низкой цене. Средняя цена входа снижается, ликвидация отодвигается вниз,
    # точка безубытка приближается к текущей цене.

    Сценарий: Лонг — усреднение вниз
      Допустим a long position at 100000 with volume 1 and margin 10000 and leverage 10x
      Когда I simulate add-on at price 90000 with margin 10000
      Тогда the new entry price should be below 100000
      И the volume should increase
      И total margin should be 20000
      И the liquidation price should move lower
      И break-even should equal the new entry price

    # Бизнес-идея: если докупка по той же цене — средняя не меняется,
    # а PnL на момент докупки равен нулю (позиция «на месте»).

    Сценарий: Докупка по той же цене не меняет среднюю
      Допустим a long position at 100000 with volume 1 and margin 10000 and leverage 10x
      Когда I simulate add-on at price 100000 with margin 10000
      Тогда the new entry price should be approximately 100000
      И unrealized PnL at add price should be approximately 0

  Правило: Усреднение вверх увеличивает среднюю цену для шорта

    # Бизнес-идея: шорт усредняется вверх — добавляет к позиции по более
    # высокой цене. Средняя цена растёт, ликвидация отодвигается выше.

    Сценарий: Шорт — усреднение вверх
      Допустим a short position at 100000 with volume 1 and margin 10000 and leverage 10x
      Когда I simulate add-on at price 110000 with margin 10000
      Тогда the new entry price should be above 100000
      И the liquidation price should move higher

  Правило: Оценка риска на определённой цене

    # Бизнес-идея: RiskAtPrice — «что будет, если цена дойдёт до X?»
    # Трейдер может оценить PnL, % от маржи, запас до ликвидации и
    # эффективное плечо на ЛЮБОЙ гипотетической цене.

    Сценарий: Лонг в убытке на 95k
      Допустим a long position at 100000 with volume 1 and margin 10000 and leverage 10x
      Когда I calculate risk at price 95000
      Тогда risk unrealized PnL should be -5000
      И risk PnL on margin should be -50%
      И risk distance to liquidation should be positive

    Сценарий: Лонг в прибыли на 110k
      Допустим a long position at 100000 with volume 1 and margin 10000 and leverage 10x
      Когда I calculate risk at price 110000
      Тогда risk unrealized PnL should be 10000
      И risk PnL on margin should be 100%

    # Бизнес-идея: на цене входа плечо должно равняться номинальному —
    # это проверка корректности формулы эффективного левериджа.

    Сценарий: На цене входа плечо равно номинальному
      Допустим a long position at 100000 with volume 1 and margin 10000 and leverage 10x
      Когда I calculate risk at price 100000
      Тогда risk effective leverage should be approximately 10
