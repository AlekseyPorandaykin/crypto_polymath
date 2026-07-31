/*
    Comparing the price of one symbol across exchanges (the "One symbol on all
    exchanges" function).

    A plain table here would answer the question "what is the price on each
    exchange", but not the one this function is opened for: where it is cheaper,
    where it is dearer, and whether picking an exchange by price is worth it at
    all. Hence the summary on top (lowest, median, highest and spread), the
    extremes highlighted in the table, and a chart at the bottom showing how the
    exchanges sit relative to the median.

    The rows are sorted by ascending price: the minimum ends up first, the maximum
    last, and on the chart the exchanges line up from cheap to expensive. The
    sorting can be changed — the columns are sortable.

    As in the other modules of this page, the layout is described through h(): the
    PrimeVue components are built against the Vue runtime build without a template
    compiler.
*/

import { h } from 'vue';
import Chart from 'primevue/chart';
import Column from 'primevue/column';
import DataTable from 'primevue/datatable';

import { DataView } from './data-view.js';
import {
    buildPricesChartData,
    formatNumber,
    formatPercent,
    PRICE_CHART_OPTIONS,
    priceStats,
} from './format.js';

// The exchanges holding an extreme price: there may be several of them when the
// prices match down to the last digit.
function exchangesAt(rows, value) {
    return rows.filter(r => r.value === value).map(r => r.exchange).join(', ');
}

function statCard(kind, label, value, note) {
    return h('div', { class: ['price-stat', 'price-stat-' + kind] }, [
        h('span', { class: 'price-stat-label' }, label),
        h('strong', { class: 'price-stat-value' }, value),
        note ? h('span', { class: 'price-stat-note' }, note) : null,
    ]);
}

export const PricesView = {
    name: 'PricesView',
    props: ['rows'],
    render() {
        const stats = priceStats(this.rows);
        // A response without numeric prices (say, no exchange returned a value)
        // has nothing to compare — we show it as a plain table.
        if (!stats) return h(DataView, { value: this.rows });

        const rows = this.rows.slice().sort((a, b) => a.value - b.value);
        const extremesDiffer = stats.max > stats.min;

        return h('div', { class: 'prices-view' }, [
            h('div', { class: 'price-stats' }, [
                statCard('min', 'Lowest', formatNumber(stats.min),
                    extremesDiffer ? exchangesAt(rows, stats.min) : 'same on all exchanges'),
                statCard('median', 'Median', formatNumber(stats.median),
                    'middle across exchanges'),
                statCard('max', 'Highest', formatNumber(stats.max),
                    extremesDiffer ? exchangesAt(rows, stats.max) : 'same on all exchanges'),
                statCard('spread', 'Spread',
                    stats.spreadPercent === null ? '—' : formatPercent(stats.spreadPercent),
                    'of the lowest price'),
            ]),

            h(DataTable, {
                value: rows,
                size: 'small',
                stripedRows: true,
                sortField: 'value',
                sortOrder: 1,
                rowClass: row => {
                    if (!extremesDiffer) return '';
                    if (row.value === stats.min) return 'row-min';
                    if (row.value === stats.max) return 'row-max';
                    return '';
                },
            }, {
                default: () => [
                    h(Column, { field: 'exchange', header: 'Exchange', sortable: true }),
                    h(Column, { field: 'value', header: 'Price', sortable: true }, {
                        body: slotProps => h('span', { class: 'price-cell' }, [
                            formatNumber(slotProps.data.value),
                            extremesDiffer && slotProps.data.value === stats.min
                                ? h('span', { class: 'price-badge min' }, 'lowest')
                                : null,
                            extremesDiffer && slotProps.data.value === stats.max
                                ? h('span', { class: 'price-badge max' }, 'highest')
                                : null,
                        ]),
                    }),
                ],
            }),

            h('div', { class: 'result-chart' }, [
                h('h4', null, 'Exchange prices against the median'),
                h('div', { class: 'chart-legend' }, [
                    h('span', { class: 'legend-min' }, 'lowest'),
                    h('span', { class: 'legend-median' }, 'median'),
                    h('span', { class: 'legend-max' }, 'highest'),
                ]),
                h('div', { class: 'chart-box' }, [
                    h(Chart, {
                        type: 'line',
                        data: buildPricesChartData(rows, stats),
                        options: PRICE_CHART_OPTIONS,
                    }),
                ]),
            ]),
        ]);
    },
};
