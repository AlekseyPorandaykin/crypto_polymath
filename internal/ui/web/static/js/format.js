/*
    Preparing an API response for display: numbers and dates in an international
    format, booleans as words, internal snake_case keys spelled with spaces, plus
    chart data and the price summary (lowest, median, highest, spread).

    The locale is deliberately not the same for everything. The audience is
    international, so numbers follow the shape market data is usually written in
    (comma separates thousands, dot separates the fraction), while the month in a
    date is spelled out. A numeric date would force a choice between the American
    07/31 and the rest of the world's 31/07, and either choice would be misread by
    part of the readers.
*/

const ISO_DATE = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/;

// Numbers follow en-US, dates follow en-GB: the "day month year" order reads
// correctly outside the United States too, and the month name removes the last
// bit of ambiguity.
const NUMBER_LOCALE = 'en-US';
const DATE_LOCALE = 'en-GB';

const DATE_TIME_OPTIONS = {
    day: '2-digit', month: 'short', year: 'numeric',
    hour: '2-digit', minute: '2-digit', second: '2-digit',
};

export function isIsoDate(v) {
    return typeof v === 'string' && ISO_DATE.test(v);
}

export function formatNumber(n) {
    const opts = Number.isInteger(n) ? {} : { maximumFractionDigits: 8 };
    return n.toLocaleString(NUMBER_LOCALE, opts);
}

export function formatScalar(v) {
    if (v === null || v === undefined || v === '') return '—';
    if (typeof v === 'boolean') return v ? 'Yes' : 'No';
    if (typeof v === 'number') return formatNumber(v);
    if (isIsoDate(v)) {
        const d = new Date(v);
        return isNaN(d.getTime()) ? v : d.toLocaleString(DATE_LOCALE, DATE_TIME_OPTIONS);
    }
    return String(v);
}

export function prettyKey(k) {
    return String(k).replace(/_/g, ' ');
}

// Lays arbitrary JSON out into PrimeVue TreeTable nodes. Needed for the
// /api/v1/server response, whose structure is naturally nested: every exchange has
// its own symbols, every indicator its own list of allowed depths, and so on.
export function buildTreeNodes(value, keyPrefix) {
    if (Array.isArray(value)) {
        return value.map((item, i) => {
            const key = keyPrefix + '-' + i;
            if (item !== null && typeof item === 'object' && !Array.isArray(item)) {
                const label = item.name ?? item.unit ?? ('#' + (i + 1));
                return { key, data: { field: String(label), value: '' }, children: buildTreeNodes(item, key) };
            }
            return { key, data: { field: '#' + (i + 1), value: formatScalar(item) } };
        });
    }
    if (value !== null && typeof value === 'object') {
        return Object.keys(value).map(k => {
            const key = keyPrefix + '-' + k;
            const v = value[k];
            if (v !== null && typeof v === 'object') {
                const count = Array.isArray(v) ? v.length : Object.keys(v).length;
                return { key, data: { field: prettyKey(k), value: '[' + count + ']' }, children: buildTreeNodes(v, key) };
            }
            return { key, data: { field: prettyKey(k), value: formatScalar(v) } };
        });
    }
    return [];
}

// The colours come from the site palette (css/app.css) so that charts do not
// stand out from the rest of the design. Chart.js takes colour values, not CSS
// variables.
const ACCENT = '#2f5bea';
const ACCENT_FILL = 'rgba(47,91,234,0.1)';
const OK = '#17915a';
const ERR = '#d93636';
const MUTED = '#98a2b3';

// A label on the time axis: no year, no seconds. The full date is long, and
// several dozen such labels do not fit on the axis, while the exact time is always
// available in the table above the chart.
function chartLabel(v) {
    if (isIsoDate(v)) {
        const d = new Date(v);
        if (!isNaN(d.getTime())) {
            return d.toLocaleString(DATE_LOCALE, {
                day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit',
            });
        }
    }
    return formatScalar(v);
}

// A time series chart. Which fields to take for the axes is known not by this
// module but by the function description in endpoints.js: candlesticks use
// start_time and close_price, indicators and analytics use datetime and value.
export function buildSeriesChartData(rows, spec) {
    return {
        labels: rows.map(r => chartLabel(r[spec.x])),
        datasets: [{
            label: spec.label,
            data: rows.map(r => r[spec.y]),
            borderColor: ACCENT,
            backgroundColor: ACCENT_FILL,
            fill: true,
            tension: 0.25,
            pointRadius: 0,
        }],
    };
}

export const CHART_OPTIONS = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: { legend: { display: false } },
    scales: {
        // The labels stay horizontal and thinned out: dates rotated by 45 degrees
        // take away a third of the chart height, and nobody is going to read all
        // forty-eight of them anyway.
        x: { ticks: { maxRotation: 0, autoSkip: true, maxTicksLimit: 8 } },
        y: { ticks: { callback: v => formatNumber(v) } },
    },
};

// How many decimal places the most precise of the values has. Needed so that the
// median does not look more precise than the source data: with prices carrying two
// decimals, the middle between 63,120.00 and 63,145.25 is 63,132.625, and the
// third decimal would suggest a precision exchange quotes do not have.
function maxDecimals(values) {
    return values.reduce((max, v) => {
        const text = String(v);
        // Exponential notation is not parsed: such values certainly carry more
        // digits than it makes sense to show.
        if (text.includes('e')) return max;
        const dot = text.indexOf('.');
        return Math.max(max, dot < 0 ? 0 : text.length - dot - 1);
    }, 0);
}

// The lowest, median and highest price of one symbol across exchanges. Returns
// null when the response holds no numeric values — the caller then shows a plain
// table instead of the comparison.
export function priceStats(rows) {
    const values = rows
        .map(r => r.value)
        .filter(v => typeof v === 'number' && Number.isFinite(v));
    if (values.length === 0) return null;

    const sorted = values.slice().sort((a, b) => a - b);
    const middle = Math.floor(sorted.length / 2);
    const min = sorted[0];
    const max = sorted[sorted.length - 1];
    // With an even number of exchanges the set has no central value, so we take
    // the middle between the two central ones — that is the median.
    const median = sorted.length % 2
        ? sorted[middle]
        : (sorted[middle - 1] + sorted[middle]) / 2;

    return {
        min,
        max,
        median: Number(median.toFixed(maxDecimals(sorted))),
        // The spread as a percentage of the minimum: this is what tells whether
        // picking an exchange by price is worth it at all. At a zero price the
        // percentage is undefined.
        spreadPercent: min > 0 ? ((max - min) / min) * 100 : null,
    };
}

export function formatPercent(v) {
    return v.toLocaleString(NUMBER_LOCALE, { maximumFractionDigits: 3 }) + '%';
}

// Price dots per exchange plus a dashed median line.
//
// Dots rather than bars, deliberately: exchange prices differ by fractions of a
// percent, so the Y axis has to be scaled to the data range (starting from zero
// would merge every value into one line). The length of a bar drawn from a
// non-zero axis origin would exaggerate the difference several times over, while
// the position of a dot on the scale reads correctly at any scale.
export function buildPricesChartData(rows, stats) {
    const extremesDiffer = stats.max > stats.min;
    return {
        datasets: [
            {
                type: 'line',
                label: 'Exchange price',
                data: rows.map(r => r.value),
                showLine: false,
                pointRadius: 7,
                pointHoverRadius: 9,
                pointBackgroundColor: rows.map(r => {
                    if (!extremesDiffer) return ACCENT;
                    if (r.value === stats.min) return OK;
                    if (r.value === stats.max) return ERR;
                    return ACCENT;
                }),
                pointBorderColor: '#fff',
                pointBorderWidth: 2,
            },
            {
                type: 'line',
                label: 'Median',
                data: rows.map(() => stats.median),
                borderColor: MUTED,
                borderDash: [6, 4],
                borderWidth: 2,
                pointRadius: 0,
                fill: false,
            },
        ],
        labels: rows.map(r => r.exchange),
    };
}

export const PRICE_CHART_OPTIONS = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
        // The legend is drawn with our own markup next to the chart: the built-in
        // one takes the series colour from the first point, which would label
        // "Exchange price" in green — the very colour that marks the minimum.
        legend: { display: false },
        tooltip: {
            callbacks: { label: ctx => ctx.dataset.label + ': ' + formatNumber(ctx.parsed.y) },
        },
    },
    scales: {
        y: {
            // grace widens the range by 40% of the data spread while keeping the
            // axis labels round. Explicit min and max would put values like
            // "63,234.874" at the edges, which are impossible to read.
            grace: '40%',
            ticks: { callback: v => formatNumber(v) },
        },
    },
};

// From this number of rows on, pagination is switched on: a hundred candlesticks
// are easier to page through than to scroll as a single block.
const TABLE_PAGE_THRESHOLD = 10;

export function tableProps(rows) {
    const base = { value: rows, size: 'small', stripedRows: true };
    return rows.length > TABLE_PAGE_THRESHOLD
        ? Object.assign(base, {
            paginator: true,
            rows: 10,
            rowsPerPageOptions: [10, 25, 50],
            scrollable: true,
            scrollHeight: '360px',
        })
        : Object.assign(base, { scrollable: true, scrollHeight: '280px' });
}
