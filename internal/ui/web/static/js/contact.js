/*
    The contact form on the landing page: opening the dialog and sending the message
    to POST /api/v1/contact.

    This is the only script on the landing page, and it stays that way on purpose.
    Everything a search or AI crawler needs is static markup; this module adds no
    content, it only makes the dialog work. If it fails to load, the markup swaps the
    button for a mail link (see the noscript block in index.html), so the page never
    ends up with a button that does nothing.

    A native <dialog> takes care of the backdrop, the Escape key and the focus trap,
    which is why there is no state machine here — only the request and what to say
    about its outcome.

    The API answers 202 rather than 200: the message is accepted for delivery, so the
    confirmation promises a reply, not that the message has already been read.
*/

const ENDPOINT = '/api/v1/contact';

const dialog = document.getElementById('contact-dialog');
const form = dialog?.querySelector('form');
const status = dialog?.querySelector('[data-contact-status]');
const submit = dialog?.querySelector('[data-contact-submit]');
const close = dialog?.querySelector('[data-contact-close]');

// Without the dialog there is nothing to wire up: the module is attached to the
// landing page only, and a missing form means the markup changed.
if (dialog && form && status && submit && close) {
    for (const trigger of document.querySelectorAll('[data-contact-open]')) {
        trigger.addEventListener('click', open);
    }
    close.addEventListener('click', () => dialog.close());
    form.addEventListener('submit', send);

    // A click on the backdrop closes the dialog. The backdrop belongs to the dialog
    // itself, so a click outside the form is a click on the element.
    dialog.addEventListener('click', event => {
        if (event.target === dialog) dialog.close();
    });
}

function open() {
    clearStatus();
    // showModal, not the open attribute: only the modal state gives the backdrop and
    // keeps the focus inside the form.
    dialog.showModal();
}

function clearStatus() {
    status.textContent = '';
    status.hidden = true;
    status.classList.remove('ok', 'err');
    // The dialog reopens with an empty form, so the button cancels again — the
    // "Close" wording belongs to a sent message only.
    close.textContent = 'Cancel';
}

function showStatus(text, kind) {
    status.textContent = text;
    status.hidden = false;
    status.classList.remove('ok', 'err');
    status.classList.add(kind);
}

async function send(event) {
    event.preventDefault();
    // The browser has already checked the fields by required, type and length, so
    // reporting is enough: it points at the first offending field.
    if (!form.reportValidity()) return;

    const payload = {
        email: form.email.value.trim(),
        subject: form.subject.value.trim(),
        message: form.message.value.trim(),
    };

    submit.disabled = true;
    showStatus('Sending…', 'ok');
    try {
        const resp = await fetch(ENDPOINT, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
        });
        if (resp.ok) {
            form.reset();
            showStatus('Message sent. We will reply to the address you provided.', 'ok');
            // There is nothing left to cancel once the message is gone: the button
            // now only closes the dialog, and its label should say so.
            close.textContent = 'Close';
            return;
        }
        showStatus(await failureText(resp), 'err');
    } catch (e) {
        // A network failure is not a rejection: the message is still valid, so the
        // fields keep their content and the visitor can simply try again.
        showStatus('Could not reach the server. Please try again in a moment.', 'err');
    } finally {
        submit.disabled = false;
    }
}

// Turns a rejection into a sentence a visitor can act on. The API explains what is
// wrong with the fields itself, and that explanation is more useful than any text
// written here in advance.
async function failureText(resp) {
    if (resp.status === 429) {
        const wait = Number(resp.headers.get('Retry-After'));
        return wait > 0
            ? `Too many requests. Please try again in ${wait} seconds.`
            : 'Too many requests. Please try again in a minute.';
    }
    let message = '';
    try {
        message = (await resp.json())?.message || '';
    } catch (e) { /* not JSON — fall back to the generic wording */ }
    if (resp.status === 400 && message) return message;
    return 'The message was not sent. Please try again later.';
}
