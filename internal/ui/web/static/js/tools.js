/*
    The tools page (/tools): the only place where the application talks to the API —
    market data and trader calculators. The landing page stayed presentational and
    knows nothing about this code.

    Layout: a TabView with two tabs ("Market data" and "Calculators"), each holding
    a Listbox of functions on the left and the parameters with the result on the
    right. There is no modal window any more: on a dedicated page the result fits
    right where it belongs.

    The tabs keep independent state, so a calculation in a calculator does not wipe
    an already fetched candlestick table, and the other way round.

    The result is shown the human way: first a table with the exact values, then a
    chart below it when the function describes one in endpoints.js. The endpoint URL
    and the HTTP method are not displayed — they explain nothing to a visitor, and
    whoever needs the contract is served by the documentation at /docs/api.

    A 429 response is handled separately: the API gives an anonymous address 10
    requests per minute, and with a token the limit follows the pricing plan. This
    page calls the API without a token and falls under the same anonymous limit, so
    the rule is stated in the page header — before the rejection, not after. Showing
    "API unavailable" or a bare status code on rejection would misstate the reason,
    which is why the limit has its own badge and its own explanation with the
    waiting time.

    As in data-view.js, the layout is described through render()/h(): the PrimeVue
    components use the Vue runtime build without a template compiler.
*/

import { createApp, reactive, ref, onMounted, h } from 'vue';
import PrimeVue from 'primevue/config';
import Button from 'primevue/button';
import Card from 'primevue/card';
import Chart from 'primevue/chart';
import Column from 'primevue/column';
import Dropdown from 'primevue/dropdown';
import InputNumber from 'primevue/inputnumber';
import Listbox from 'primevue/listbox';
import Message from 'primevue/message';
import SelectButton from 'primevue/selectbutton';
import Skeleton from 'primevue/skeleton';
import TabPanel from 'primevue/tabpanel';
import TabView from 'primevue/tabview';
import Tag from 'primevue/tag';
import TreeTable from 'primevue/treetable';

import { DataView } from './data-view.js';
import { PricesView } from './prices-view.js';
import { buildSeriesChartData, buildTreeNodes, CHART_OPTIONS } from './format.js';
import {
    buildBody,
    buildUrl,
    calculatorEndpoints,
    dataEndpoints,
    groupByTag,
    refreshEndpoint,
    resolveOptions,
} from './endpoints.js';

const DATA_TAB = 0;
const CALCULATOR_TAB = 1;

function createSection(endpoints) {
    return reactive({
        endpoints,
        groups: groupByTag(endpoints),
        selectedKey: endpoints[0].key,
        result: {
            loading: false, error: false, status: '', data: null, raw: '', showRaw: false, network: '',
            // Seconds to wait after a rate limit rejection. Empty string means there was none.
            retryAfter: '',
        },
    });
}

const ToolsRoot = {
    setup() {
        const serverData = ref(null);
        const online = ref(null);
        // A rate limit rejection is not the same as an unavailable API: the service
        // works, it simply refused to hand over the reference data. Showing "API
        // unavailable" here would mislead the user and hide the real reason.
        const limited = ref(false);
        const activeTab = ref(DATA_TAB);
        const sections = [createSection(dataEndpoints), createSection(calculatorEndpoints)];

        function activeEndpoint(section) {
            return section.endpoints.find(e => e.key === section.selectedKey) || null;
        }

        function resetResult(section) {
            Object.assign(section.result, {
                loading: false, error: false, status: '', data: null, raw: '', showRaw: false, network: '',
                retryAfter: '',
            });
        }

        // Keep ?fn= in the address bar up to date so that a link to a particular
        // function can be copied and shared — the landing page buttons use the very
        // same parameter.
        function syncLocation(key) {
            const url = new URL(window.location.href);
            url.searchParams.set('fn', key);
            window.history.replaceState(null, '', url);
        }

        function select(section, key) {
            if (!key || key === section.selectedKey) return;
            section.selectedKey = key;
            resetResult(section);
            const ep = activeEndpoint(section);
            if (ep) refreshEndpoint(ep, serverData.value);
            syncLocation(key);
        }

        async function run(section) {
            const ep = activeEndpoint(section);
            if (!ep) return;
            Object.assign(section.result, {
                loading: true, error: false, data: null, raw: '', showRaw: false, network: '', retryAfter: '',
            });
            try {
                const resp = ep.method === 'POST'
                    ? await fetch(ep.path, {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify(buildBody(ep)),
                    })
                    : await fetch(buildUrl(ep));
                const text = await resp.text();
                let data = null;
                try { data = JSON.parse(text); } catch (e) { /* not JSON — show it as is */ }
                section.result.status = resp.status + ' ' + resp.statusText;
                section.result.error = !resp.ok;
                // The API limits how often anonymous requests may be sent and names
                // the waiting time in the rejection. The message itself sits inside
                // the json — without a separate hint the user only sees a status code.
                if (resp.status === 429) {
                    section.result.retryAfter = resp.headers.get('Retry-After') || '10';
                }
                section.result.data = data;
                section.result.raw = data !== null ? JSON.stringify(data, null, 2) : text;
            } catch (e) {
                section.result.status = 'Request failed';
                section.result.error = true;
                section.result.network = String(e);
            } finally {
                section.result.loading = false;
            }
        }

        onMounted(async () => {
            try {
                const resp = await fetch('/api/v1/server');
                online.value = resp.ok;
                limited.value = resp.status === 429;
                if (resp.ok) {
                    serverData.value = await resp.json();
                    // We walk the functions through the reactive tab state rather
                    // than the source arrays: Vue does not track edits to raw
                    // objects, and the filled-in values would never reach the form.
                    sections.forEach(s => s.endpoints.forEach(ep => refreshEndpoint(ep, serverData.value)));
                }
            } catch (e) {
                online.value = false;
            }

            const requested = new URLSearchParams(window.location.search).get('fn');
            if (!requested) return;
            if (dataEndpoints.some(e => e.key === requested)) {
                activeTab.value = DATA_TAB;
                select(sections[DATA_TAB], requested);
            } else if (calculatorEndpoints.some(e => e.key === requested)) {
                activeTab.value = CALCULATOR_TAB;
                select(sections[CALCULATOR_TAB], requested);
            }
        });

        function renderField(ep, f) {
            if (f.control === 'buttons') {
                return h(SelectButton, {
                    modelValue: ep.values[f.name],
                    'onUpdate:modelValue': v => { if (v !== null) ep.values[f.name] = v; },
                    options: f.options,
                });
            }
            if (f.options || f.optionsFn) {
                return h(Dropdown, {
                    modelValue: ep.values[f.name],
                    'onUpdate:modelValue': v => { ep.values[f.name] = v; },
                    options: resolveOptions(ep, f, serverData.value),
                    onChange: () => refreshEndpoint(ep, serverData.value),
                });
            }
            return h(InputNumber, {
                modelValue: ep.values[f.name],
                'onUpdate:modelValue': v => { ep.values[f.name] = v; },
                useGrouping: false,
                minFractionDigits: 0,
                maxFractionDigits: 8,
            });
        }

        function renderResultBody(ep, result) {
            if (result.loading) {
                return h('div', { class: 'result-skeleton' }, [
                    h(Skeleton, { height: '1.4rem' }),
                    h(Skeleton, { height: '1.4rem', width: '80%' }),
                    h(Skeleton, { height: '1.4rem', width: '60%' }),
                ]);
            }
            if (result.network) {
                return h(Message, { severity: 'error', closable: false }, {
                    default: () => 'Could not reach the API: ' + result.network,
                });
            }
            if (result.retryAfter) {
                // The number from Retry-After is the time until the next request,
                // not the quota itself: calling it the limit would misstate the rule.
                return h(Message, { severity: 'warn', closable: false }, {
                    default: () => 'Too many requests. The rate limit for calls without a token is used up — wait '
                        + result.retryAfter + ' s and try again. With a token the limit follows your pricing plan.',
                });
            }
            if (result.data === null || result.showRaw) {
                return h('pre', { class: 'raw' }, result.raw);
            }

            const rows = Array.isArray(result.data) ? result.data : null;

            if (ep.view === 'prices' && rows && rows.length > 0) {
                return h(PricesView, { rows });
            }
            if (ep.tree) {
                return h(TreeTable, { value: buildTreeNodes(result.data, 'root'), size: 'small' }, {
                    default: () => [
                        h(Column, { field: 'field', header: 'Field', expander: true }),
                        h(Column, { field: 'value', header: 'Value' }),
                    ],
                });
            }

            // Exact values in a table first, the chart below it. Numbers are needed
            // more often: they are what a particular candlestick or indicator value
            // is checked against, while the chart answers a different question —
            // where the series is heading as a whole.
            return [
                h(DataView, { value: result.data }),
                (ep.chart && rows && rows.length > 0)
                    ? h('div', { class: 'result-chart' }, [
                        h('h4', null, ep.chart.label),
                        h('div', { class: 'chart-box' }, [
                            h(Chart, {
                                type: 'line',
                                data: buildSeriesChartData(rows, ep.chart),
                                options: CHART_OPTIONS,
                            }),
                        ]),
                    ])
                    : null,
            ];
        }

        function renderResult(ep, result) {
            if (!result.loading && !result.status) return null;
            return h('div', { class: 'result-panel' }, [
                h('div', { class: 'result-head' }, [
                    h('span', { class: ['status-line', result.error ? 'err' : 'ok'] },
                        result.loading ? 'Running…' : result.status),
                    (!result.loading && result.data !== null)
                        ? h(Button, {
                            class: 'p-button-text p-button-sm',
                            label: result.showRaw ? 'Formatted' : 'Raw JSON',
                            onClick: () => { result.showRaw = !result.showRaw; },
                        })
                        : null,
                ]),
                h('div', { class: 'result-body' }, [renderResultBody(ep, result)]),
            ]);
        }

        function renderWorkspace(section) {
            const ep = activeEndpoint(section);
            if (!ep) {
                return h('div', { class: 'fn-empty' }, 'Pick a function from the list on the left.');
            }
            return h(Card, null, {
                content: () => [
                    h('h2', { class: 'fn-title' }, ep.label),
                    h('p', { class: 'fn-desc' }, ep.desc),
                    ep.fields.length
                        ? h('form', { class: 'params', onSubmit: e => e.preventDefault() },
                            ep.fields.map(f => h('label', { key: f.name }, [f.name, renderField(ep, f)])))
                        : null,
                    h('div', { class: 'run-row' }, [
                        h(Button, {
                            label: 'Run',
                            icon: 'pi pi-play',
                            loading: section.result.loading,
                            onClick: () => run(section),
                        }),
                    ]),
                    renderResult(ep, section.result),
                ],
            });
        }

        function renderSection(section) {
            return h('div', { class: 'tools-grid' }, [
                h(Listbox, {
                    modelValue: section.selectedKey,
                    'onUpdate:modelValue': v => select(section, v),
                    options: section.groups,
                    optionLabel: 'label',
                    optionValue: 'key',
                    optionGroupLabel: 'tag',
                    optionGroupChildren: 'items',
                    filter: true,
                    filterPlaceholder: 'Search functions',
                }),
                renderWorkspace(section),
            ]);
        }

        function renderStatus() {
            if (online.value === null) return h(Tag, { value: 'Checking API…' });
            if (limited.value) {
                return h(Tag, { value: 'Rate limited', severity: 'warning', icon: 'pi pi-clock' });
            }
            return online.value
                ? h(Tag, { value: 'API online', severity: 'success', icon: 'pi pi-check' })
                : h(Tag, { value: 'API unavailable', severity: 'danger', icon: 'pi pi-times' });
        }

        return () => h('div', { class: 'tools-shell' }, [
            h('div', { class: 'tools-intro' }, [
                renderStatus(),
                h('span', { style: 'margin-left:10px;' },
                    'Pick a function, fill in the parameters and run the request — the result appears below.'),
                // The limit rule is stated here, not only in the rejection message:
                // this page calls the API without a token and runs into the same
                // limit, so it is better known before the rejection than after.
                h('p', { class: 'tools-limits' }, [
                    'This page uses the same API: 10 requests per minute without a token, ',
                    'or the quota of the pricing plan your token was issued with. ',
                    h('a', { href: '/docs/api' }, 'See the documentation'),
                    '.',
                ]),
            ]),
            h(TabView, {
                activeIndex: activeTab.value,
                'onUpdate:activeIndex': v => { activeTab.value = v; },
            }, {
                default: () => [
                    h(TabPanel, { header: 'Market data' }, { default: () => renderSection(sections[DATA_TAB]) }),
                    h(TabPanel, { header: 'Calculators' }, { default: () => renderSection(sections[CALCULATOR_TAB]) }),
                ],
            }),
        ]);
    },
};

const app = createApp(ToolsRoot);
app.use(PrimeVue);
app.mount('#tools-app');
