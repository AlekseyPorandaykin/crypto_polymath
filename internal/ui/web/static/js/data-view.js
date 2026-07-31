/*
    Recursive presentable rendering of an arbitrary JSON response from the API.

    The PrimeVue components are built against the Vue runtime build, which carries
    no template compiler, so here and in tools.js the whole layout is described
    through render()/h() rather than string templates.

    Display rules:
      - array of objects        -> DataTable with a column per field encountered;
      - array of plain values   -> DataTable with a single column (also paginated);
      - object                  -> a "key - value" list;
      - scalar                  -> formatted text; numbers are tinted by sign,
                                   booleans by value.
*/

import { h } from 'vue';
import DataTable from 'primevue/datatable';
import Column from 'primevue/column';

import { formatScalar, prettyKey, tableProps } from './format.js';

export const DataView = {
    name: 'DataView',
    props: ['value'],
    render() {
        const v = this.value;
        const isArray = Array.isArray(v);
        const isArrayOfObjects = isArray
            && v.length > 0
            && v.every(x => x && typeof x === 'object' && !Array.isArray(x));

        if (isArrayOfObjects) {
            const columns = [];
            v.forEach(row => Object.keys(row).forEach(k => {
                if (!columns.includes(k)) columns.push(k);
            }));
            return h(DataTable, tableProps(v), {
                default: () => columns.map(k => h(Column, { key: k, field: k, header: prettyKey(k) }, {
                    body: (slotProps) => h(DataView, { value: slotProps.data[k] }),
                })),
            });
        }

        if (isArray) {
            const wrapped = v.map(item => ({ value: item }));
            return h(DataTable, tableProps(wrapped), {
                default: () => h(Column, { field: 'value', header: 'Value' }, {
                    body: (slotProps) => h(DataView, { value: slotProps.data.value }),
                }),
            });
        }

        if (v !== null && typeof v === 'object') {
            const rows = [];
            Object.keys(v).forEach(k => {
                rows.push(h('dt', { key: 'dt-' + k }, prettyKey(k)));
                rows.push(h('dd', { key: 'dd-' + k }, [h(DataView, { value: v[k] })]));
            });
            return h('dl', { class: 'dv-kv' }, rows);
        }

        let cls = '';
        if (typeof v === 'number') cls = v > 0 ? 'dv-pos' : v < 0 ? 'dv-neg' : '';
        if (typeof v === 'boolean') cls = v ? 'dv-bool-true' : 'dv-bool-false';
        return h('span', { class: cls }, formatScalar(v));
    },
};
