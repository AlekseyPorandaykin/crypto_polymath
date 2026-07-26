# language: ru
@trading
Функционал: Калькулятор фьючерсных позиций
  Как трейдер
  Я хочу рассчитывать метрики позиции
  Чтобы эффективно управлять рисками

  # -------------------------------------------------------------------
  # Контекст:
  # Этот feature описывает базовые расчёты для маржинальной торговли
  # фьючерсами (isolated margin, USDT-M):
  # - Средняя цена входа при докупке (усреднение позиции)
  # - Цена ликвидации и факторы, на неё влияющие
  # - Нереализованный PnL (прибыль/убыток до закрытия позиции)
  # - Расчёт объёма позиции из суммы залога
  # - Запас до ликвидации в процентах
  #
  # Бизнес-задача: трейдер должен понимать, как докупка влияет на его
  # позицию, насколько далеко ликвидация, и какой текущий PnL — всё это
  # критически важно для управления капиталом.
  # -------------------------------------------------------------------

  Контекст:
    Допустим MMR is 0.5%

  Правило: Средняя цена входа — средневзвешенная по объёму

    # Бизнес-идея: при докупке (усреднении) новая средняя цена должна быть
    # между ценами двух входов, взвешенная по объёмам.
    # Это позволяет трейдеру понять, куда сместится его точка безубытка.

    Структура сценария: Средневзвешенная двух входов
      Допустим an existing position with volume <v1> at price <p1>
      Когда I add volume <v2> at price <p2>
      Тогда the average entry price should be <avg>

      Примеры:
        | v1 | p1  | v2 | p2  | avg |
        | 1  | 100 | 1  | 200 | 150 |
        | 3  | 100 | 1  | 200 | 125 |
        | 2  | 50  | 4  | 50  | 50  |

    # Бизнес-идея: если трейдер добавляет к позиции определённую СУММУ (а не объём),
    # система должна сначала рассчитать объём через плечо, а затем среднюю цену.

    Сценарий: Средняя цена при добавлении суммой с плечом
      Допустим leverage is 2x
      Когда I calculate avg entry price by sum with volume 1 at 100, adding sum 200 at price 200
      Тогда the average entry price should be approximately 166.667

  Правило: Цена ликвидации зависит от направления позиции

    # Бизнес-идея: ликвидация — это момент, когда убыток «съедает» залог.
    # Для лонга цена ликвидации НИЖЕ входа (цена должна упасть).
    # Для шорта цена ликвидации ВЫШЕ входа (цена должна вырасти).
    # Это фундаментальное свойство маржинальной торговли.

    Сценарий: Ликвидация лонга ниже цены входа
      Допустим a long position at 100000 with volume 1 and margin 10000
      Когда I calculate the liquidation price
      Тогда the liquidation price should be below 100000

    Сценарий: Ликвидация шорта выше цены входа
      Допустим a short position at 100000 with volume 1 and margin 10000
      Когда I calculate the liquidation price
      Тогда the liquidation price should be above 100000

    # Бизнес-идея: при одинаковых параметрах ликвидация лонга всегда ниже
    # ликвидации шорта — это гарантия корректности формул.

    Сценарий: Ликвидация лонга всегда ниже ликвидации шорта
      Допустим a long position at 100000 with volume 1 and margin 10000
      Когда I calculate the liquidation price
      И I also calculate the short liquidation price at 100000 with volume 1 and margin 10000
      Тогда the long liquidation should be below the short liquidation

    # Бизнес-идея: больше залога = больше «подушка безопасности» = ликвидация дальше.
    # Трейдер может добавить маржу, чтобы отодвинуть ликвидацию.

    Сценарий: Больше залога — ликвидация дальше от входа
      Допустим a long position at 100000 with volume 1 and margin 10000
      Когда I calculate the liquidation price
      И I calculate the liquidation price with margin 20000
      Тогда the second liquidation should be further from entry

  Правило: Нереализованный PnL

    # Бизнес-идея: PnL — это текущая «бумажная» прибыль или убыток.
    # Лонг зарабатывает на росте, шорт — на падении.
    # При одинаковых параметрах PnL лонга + PnL шорта = 0 (нулевая сумма рынка).

    Сценарий: Лонг в прибыли
      Допустим a long position at 100000 with volume 1 and margin 10000
      Когда market price is 110000
      Тогда unrealized PnL should be 10000

    Сценарий: Шорт в прибыли
      Допустим a short position at 100000 with volume 1 and margin 10000
      Когда market price is 90000
      Тогда unrealized PnL should be 10000

    Сценарий: PnL лонга и шорта в сумме дают ноль
      Допустим a long position at 100000 with volume 1 and margin 10000
      Когда market price is 105000
      Тогда long PnL plus short PnL should equal zero

  Правило: Объём из залога

    # Бизнес-идея: трейдер указывает сумму залога (маржу), а система
    # рассчитывает объём позиции: Volume = Margin × Leverage / Price.
    # Это основа для расчёта размера позиции.

    Структура сценария: Volume = margin × leverage / price
      Допустим leverage is <leverage>x
      Когда I calculate volume from margin <margin> at price <price>
      Тогда the volume should be <volume>

      Примеры:
        | leverage | margin | price  | volume |
        | 10       | 1000   | 50000  | 0.2    |
        | 5        | 2000   | 100000 | 0.1    |
        | 20       | 500    | 10000  | 1.0    |

  Правило: Запас до ликвидации

    # Бизнес-идея: трейдеру важно знать, на сколько процентов цена может
    # двинуться против него до ликвидации. Это ключевой показатель риска.

    Сценарий: Запас до ликвидации лонга
      Допустим a long position at mark price 100000 with liquidation at 90000
      Тогда distance to liquidation should be 10%

    Сценарий: Запас до ликвидации шорта
      Допустим a short position at mark price 100000 with liquidation at 110000
      Тогда distance to liquidation should be 10%
