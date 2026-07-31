/*
    Descriptions of every function on the tools page — data, not logic. Each entry
    mirrors one endpoint from api/rest/v1/openapi.yaml:

      key      - a stable identifier, also the value of the ?fn=<key> deep link;
      tag      - the group in the function list;
      label    - a short name for the list;
      desc     - what the endpoint does;
      method   - GET substitutes the values into the path, POST sends them as a
                 JSON body;
      fields   - the parameters; nesting inside the body is expressed by a dot in
                 the name (for example position.side -> {"position":{"side":...}});
      chart    - when set, a chart is drawn below the result table: {x, y} name the
                 response fields for the axes, label captions the series;
      view     - a special way of showing the result instead of the usual table
                 ('prices' compares the price across exchanges, 'tree' is set by
                 the tree field).

    The method and path fields are only needed to send the request: neither the URL
    nor the HTTP method is shown to a visitor.

    Lists of allowed values (exchanges, symbols, timeframes, depths) are not
    hardcoded here but derived from the GET /api/v1/server response by the opt*
    functions below, so the frontend cannot drift from the server configuration.
*/

function field(name, opts) {
    return Object.assign({ name, default: '' }, opts);
}

function withValues(fields) {
    const values = {};
    for (const f of fields) values[f.name] = f.default;
    return values;
}

function endpoint(e) {
    return Object.assign({ values: withValues(e.fields) }, e);
}

const optExchanges = (values, srv) => (srv ? srv.exchanges : []);
const optSymbols = (values, srv) => (srv ? srv.symbols : []);
const optUnits = (values, srv) => (srv ? srv.units : []);
const optIntervalsForUnit = (values, srv) => {
    if (!srv) return [];
    const entry = srv.intervals.find(iv => iv.unit === values.unit);
    return entry ? entry.values : [];
};
const optIndicatorNames = (values, srv) => (srv ? srv.indicators.map(i => i.name) : []);
const optIndicatorDepth = (values, srv) => {
    if (!srv) return [];
    const info = srv.indicators.find(i => i.name === values.name);
    return info ? info.depth : srv.depths;
};
const optAnalysisNames = (values, srv) => (srv ? srv.analysis.map(a => a.name) : []);
const optAnalysisDepth = (values, srv) => {
    if (!srv) return [];
    const info = srv.analysis.find(a => a.name === values.name);
    return info ? info.depth : srv.depths;
};
const optAnalysisIndicatorDepth = (values, srv) => {
    if (!srv) return [];
    const info = srv.analysis.find(a => a.name === values.name);
    return info ? info.indicator_depth : srv.indicator_depth;
};

// The position side field — rendered as a SelectButton rather than a dropdown.
const sideField = (name) => field(name, { default: 'long', options: ['long', 'short'], control: 'buttons' });

// The chart below the table: which response fields to take for the axes and how to
// caption the series. Candlesticks and Heiken Ashi return the time in start_time,
// indicators and analytics in datetime, which is why the field names live here
// instead of being hardcoded in the chart module.
const closePriceChart = { x: 'start_time', y: 'close_price', label: 'Closing price' };
const indicatorChart = { x: 'datetime', y: 'value', label: 'Indicator value' };
const analysisChart = { x: 'datetime', y: 'value', label: 'Metric value' };

// Market data: prices, candlesticks, indicators, analytics and reference data.
export const dataEndpoints = [
    endpoint({
        key: 'price', tag: 'Prices', label: 'Price by exchange and symbol',
        method: 'GET', path: '/api/v1/price/{exchange}/{symbol}',
        desc: 'The latest price of a specific pair on a specific exchange.',
        fields: [field('exchange', { optionsFn: optExchanges }), field('symbol', { optionsFn: optSymbols })],
    }),
    endpoint({
        key: 'prices-exchange', tag: 'Prices', label: 'All prices on an exchange',
        method: 'GET', path: '/api/v1/prices/exchange/{exchange}',
        desc: 'The latest prices for every symbol on the selected exchange.',
        fields: [field('exchange', { optionsFn: optExchanges })],
    }),
    endpoint({
        key: 'prices-symbol', tag: 'Prices', label: 'One symbol on all exchanges',
        method: 'GET', path: '/api/v1/prices/symbol/{symbol}',
        desc: 'The latest price of one symbol across every connected exchange — the quickest way to see the spread.',
        view: 'prices',
        fields: [field('symbol', { optionsFn: optSymbols })],
    }),
    endpoint({
        key: 'candlestick', tag: 'Candlesticks', label: 'Candlesticks',
        method: 'GET', path: '/api/v1/candlestick/{exchange}/{symbol}/{unit}/{interval}',
        desc: 'Historical OHLCV candlesticks by exchange, symbol and timeframe: a table of values with a closing price chart below it.',
        chart: closePriceChart,
        fields: [
            field('exchange', { optionsFn: optExchanges }),
            field('symbol', { optionsFn: optSymbols }),
            field('unit', { optionsFn: optUnits }),
            field('interval', { optionsFn: optIntervalsForUnit }),
        ],
    }),
    endpoint({
        key: 'candle-indicator', tag: 'Candlesticks', label: 'Heiken Ashi',
        method: 'GET', path: '/api/v1/candle-indicator/{exchange}/{symbol}/{unit}/{interval}/{name}',
        desc: 'An indicator that reshapes the candlestick itself — Heiken Ashi smoothed candlesticks.',
        chart: closePriceChart,
        fields: [
            field('exchange', { optionsFn: optExchanges }),
            field('symbol', { optionsFn: optSymbols }),
            field('unit', { optionsFn: optUnits }),
            field('interval', { optionsFn: optIntervalsForUnit }),
            field('name', { default: 'HeikenAshi', options: ['HeikenAshi'] }),
        ],
    }),
    endpoint({
        key: 'indicator', tag: 'Indicators', label: 'Technical indicator',
        method: 'GET', path: '/api/v1/indicator/{exchange}/{symbol}/{unit}/{interval}/{name}/{depth}',
        desc: 'Technical indicator values calculated from candlesticks: a table with a chart of the values below it.',
        chart: indicatorChart,
        fields: [
            field('exchange', { optionsFn: optExchanges }),
            field('symbol', { optionsFn: optSymbols }),
            field('unit', { optionsFn: optUnits }),
            field('interval', { optionsFn: optIntervalsForUnit }),
            field('name', { optionsFn: optIndicatorNames }),
            field('depth', { optionsFn: optIndicatorDepth }),
        ],
    }),
    endpoint({
        key: 'analysis', tag: 'Indicators', label: 'Indicator analytics',
        method: 'GET', path: '/api/v1/analysis/{exchange}/{symbol}/{unit}/{interval}/{name}/{indicator_depth}/{depth}',
        desc: 'Derived analytical metrics: trend, RSI, MACD, stochastic and others.',
        chart: analysisChart,
        fields: [
            field('exchange', { optionsFn: optExchanges }),
            field('symbol', { optionsFn: optSymbols }),
            field('unit', { optionsFn: optUnits }),
            field('interval', { optionsFn: optIntervalsForUnit }),
            field('name', { optionsFn: optAnalysisNames }),
            field('indicator_depth', { optionsFn: optAnalysisIndicatorDepth }),
            field('depth', { optionsFn: optAnalysisDepth }),
        ],
    }),
    endpoint({
        key: 'exchange', tag: 'Reference data', label: 'Symbol details',
        method: 'GET', path: '/api/v1/exchange/{exchange}/{symbol}',
        desc: 'Symbol details on an exchange: base and quote asset, funding rate and next funding time.',
        fields: [field('exchange', { optionsFn: optExchanges }), field('symbol', { optionsFn: optSymbols })],
    }),
    endpoint({
        key: 'symbols', tag: 'Reference data', label: 'Exchange symbols',
        method: 'GET', path: '/api/v1/symbols/{exchange}/{category}',
        desc: 'The symbol list of an exchange in the selected category — spot or futures.',
        fields: [
            field('exchange', { optionsFn: optExchanges }),
            field('category', { default: 'spot', options: ['spot', 'future'] }),
        ],
    }),
    endpoint({
        key: 'server', tag: 'Reference data', label: 'Server capabilities',
        method: 'GET', path: '/api/v1/server',
        desc: 'What the server supports: exchanges, symbols, timeframes, indicators and analysis types.',
        tree: true,
        fields: [],
    }),
];

// Trader calculators — computed on the server, they need no input from exchanges.
export const calculatorEndpoints = [
    endpoint({
        key: 'avg-entry-price', tag: 'Position entry', label: 'Average entry price by volume',
        method: 'POST', path: '/api/v1/calculator/trading/avg-entry-price',
        desc: 'The average entry price after adding a given volume to the position.',
        fields: [
            field('entry_volume', { default: 1, input: 'number' }),
            field('entry_price', { default: 50000, input: 'number' }),
            field('new_volume', { default: 1, input: 'number' }),
            field('new_price', { default: 45000, input: 'number' }),
        ],
    }),
    endpoint({
        key: 'avg-entry-price-by-sum', tag: 'Position entry', label: 'Average entry price by amount',
        method: 'POST', path: '/api/v1/calculator/trading/avg-entry-price-by-sum',
        desc: 'The average entry price when you add to the position by margin amount with leverage instead of by volume.',
        fields: [
            field('entry_volume', { default: 1, input: 'number' }),
            field('entry_price', { default: 50000, input: 'number' }),
            field('sum', { default: 1000, input: 'number' }),
            field('leverage', { default: 10, input: 'number' }),
            field('new_price', { default: 45000, input: 'number' }),
        ],
    }),
    endpoint({
        key: 'volume-from-margin', tag: 'Position entry', label: 'Position size from margin',
        method: 'POST', path: '/api/v1/calculator/trading/volume-from-margin',
        desc: 'The position size your margin buys at the chosen leverage and current price.',
        fields: [
            field('margin', { default: 1000, input: 'number' }),
            field('leverage', { default: 10, input: 'number' }),
            field('price', { default: 50000, input: 'number' }),
        ],
    }),
    endpoint({
        key: 'liquidation-price', tag: 'Risk', label: 'Liquidation price',
        method: 'POST', path: '/api/v1/calculator/trading/liquidation-price',
        desc: 'The price at which an isolated position gets liquidated.',
        fields: [
            sideField('side'),
            field('volume', { default: 1, input: 'number' }),
            field('entry_price', { default: 50000, input: 'number' }),
            field('margin', { default: 1000, input: 'number' }),
            field('maintenance_margin_rate', { default: 0.005, input: 'number' }),
        ],
    }),
    endpoint({
        key: 'distance-to-liquidation', tag: 'Risk', label: 'Distance to liquidation',
        method: 'POST', path: '/api/v1/calculator/trading/distance-to-liquidation',
        desc: 'How far, in percent, the price can move against the position before liquidation.',
        fields: [
            sideField('side'),
            field('mark_price', { default: 52000, input: 'number' }),
            field('liquidation_price', { default: 45000, input: 'number' }),
        ],
    }),
    endpoint({
        key: 'risk-at-price', tag: 'Risk', label: 'Risk snapshot at a price',
        method: 'POST', path: '/api/v1/calculator/trading/risk-at-price',
        desc: 'The full picture of a position at a given market price: PnL, liquidation and safety margin.',
        fields: [
            sideField('position.side'),
            field('position.volume', { default: 1, input: 'number' }),
            field('position.entry_price', { default: 50000, input: 'number' }),
            field('position.margin', { default: 1000, input: 'number' }),
            field('position.leverage', { default: 10, input: 'number' }),
            field('mark_price', { default: 52000, input: 'number' }),
            field('maintenance_margin_rate', { default: 0.005, input: 'number' }),
        ],
    }),
    endpoint({
        key: 'simulate-add-on', tag: 'Risk', label: 'Add-on simulation',
        method: 'POST', path: '/api/v1/calculator/trading/simulate-add-on',
        desc: 'What happens to entry price, liquidation and risk if you average down the position.',
        fields: [
            sideField('position.side'),
            field('position.volume', { default: 1, input: 'number' }),
            field('position.entry_price', { default: 50000, input: 'number' }),
            field('position.margin', { default: 1000, input: 'number' }),
            field('position.leverage', { default: 10, input: 'number' }),
            field('add_on.price', { default: 45000, input: 'number' }),
            field('add_on.volume', { default: 0, input: 'number' }),
            field('add_on.margin', { default: 500, input: 'number' }),
            field('maintenance_margin_rate', { default: 0.005, input: 'number' }),
        ],
    }),
    endpoint({
        key: 'unrealized-pnl', tag: 'Result', label: 'Unrealized PnL',
        method: 'POST', path: '/api/v1/calculator/trading/unrealized-pnl',
        desc: 'The paper profit or loss on an open position at the current price.',
        fields: [
            sideField('side'),
            field('volume', { default: 1, input: 'number' }),
            field('entry_price', { default: 50000, input: 'number' }),
            field('mark_price', { default: 52000, input: 'number' }),
        ],
    }),
    endpoint({
        key: 'spot-pnl', tag: 'Result', label: 'Spot position PnL',
        method: 'POST', path: '/api/v1/calculator/trading/spot-pnl',
        desc: 'The profit or loss on a spot position, both in value and in percent.',
        fields: [
            field('volume', { default: 1, input: 'number' }),
            field('entry_price', { default: 50000, input: 'number' }),
            field('mark_price', { default: 55000, input: 'number' }),
        ],
    }),
];

// Groups the functions by tag into the model PrimeVue Listbox understands
// (optionGroupLabel="tag" / optionGroupChildren="items").
export function groupByTag(endpoints) {
    const groups = [];
    for (const ep of endpoints) {
        let group = groups.find(g => g.tag === ep.tag);
        if (!group) {
            group = { tag: ep.tag, items: [] };
            groups.push(group);
        }
        group.items.push({ key: ep.key, label: ep.label });
    }
    return groups;
}

// Brings dependent field values (interval depends on unit, depth on name) back to
// the ones allowed for the current selection, using the /api/v1/server reference.
export function resolveOptions(ep, f, serverData) {
    if (f.optionsFn) return f.optionsFn(ep.values, serverData) || [];
    return f.options || [];
}

export function refreshEndpoint(ep, serverData) {
    for (const f of ep.fields) {
        if (!f.optionsFn) continue;
        const opts = resolveOptions(ep, f, serverData);
        if (!opts.includes(ep.values[f.name])) {
            ep.values[f.name] = opts.length ? opts[0] : '';
        }
    }
}

export function buildUrl(ep) {
    let url = ep.path;
    for (const f of ep.fields) {
        url = url.split('{' + f.name + '}').join(encodeURIComponent(ep.values[f.name] ?? ''));
    }
    return url;
}

export function buildBody(ep) {
    const body = {};
    for (const f of ep.fields) {
        const raw = ep.values[f.name];
        if (raw === '' || raw === null || raw === undefined) continue;
        const val = f.input === 'number' ? Number(raw) : raw;
        const parts = f.name.split('.');
        let cur = body;
        parts.forEach((p, i) => {
            if (i === parts.length - 1) cur[p] = val;
            else cur = (cur[p] = cur[p] || {});
        });
    }
    return body;
}
