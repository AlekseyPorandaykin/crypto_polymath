/*
    Initialising Redoc on the /docs/api page.

    The theme is set by hand so that the documentation looks like part of the site
    rather than a third-party widget: the same accent colour, fonts and header
    gradient as on the landing page (see css/app.css).

    This is a classic script, not a module: redoc.standalone.js publishes itself as
    the global Redoc object. Both scripts are attached with defer, so by the time
    this file runs the bundle is already parsed and the DOM is built.
*/

(function () {
    'use strict';

    var SPEC_URL = '/docs/api/openapi.yaml';
    var CONTAINER = '#redoc';

    var SANS = '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Arial, sans-serif';
    var MONO = 'ui-monospace, SFMono-Regular, Menlo, monospace';

    var options = {
        // The right column with samples repeats the dark hero gradient of the
        // landing page.
        theme: {
            colors: {
                primary: { main: '#2f5bea' },
                success: { main: '#17915a' },
                error: { main: '#d93636' },
                text: { primary: '#131a2a', secondary: '#667085' },
                http: {
                    get: '#2f5bea',
                    post: '#17915a',
                    put: '#c47f17',
                    delete: '#d93636',
                },
            },
            typography: {
                fontSize: '14px',
                lineHeight: '1.5',
                fontFamily: SANS,
                headings: { fontFamily: SANS, fontWeight: '700' },
                code: { fontFamily: MONO, fontSize: '12.5px' },
            },
            sidebar: {
                backgroundColor: '#f5f7fa',
                textColor: '#131a2a',
                activeTextColor: '#2f5bea',
            },
            rightPanel: {
                backgroundColor: '#101a33',
                textColor: '#e8ecfa',
            },
        },
        // The site header is sticky, so Redoc has to account for its height —
        // otherwise jumping between sections hides the heading behind the header.
        scrollYOffset: 'header',
        expandResponses: '200',
        hideDownloadButton: false,
        nativeScrollbars: false,
    };

    function showError(message) {
        var container = document.querySelector(CONTAINER);
        if (!container) return;
        container.innerHTML = ''
            + '<div class="docs-fallback">'
            + '<p>Could not render the documentation: ' + message + '</p>'
            + '<p class="muted">The raw contract: '
            + '<a href="' + SPEC_URL + '">' + SPEC_URL + '</a></p>'
            + '</div>';
    }

    if (typeof Redoc === 'undefined') {
        showError('the documentation renderer failed to load');
        return;
    }

    Redoc.init(SPEC_URL, options, document.querySelector(CONTAINER), function (err) {
        if (err) showError(String(err));
    });
})();
