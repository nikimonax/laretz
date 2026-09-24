# Plain MPA: PRG, method spoofing, HTML dialogs

The app is a server-rendered MPA with no SPA framework. All mutations are POST forms followed by a redirect (PRG); deletes reuse the update endpoint with a `_method=DELETE` hidden field handled by a method-spoofing middleware; confirmations and create/edit forms are `<dialog>` modals requiring a single line of inline JS (`showModal()`); validation errors re-render the page. The only JSON is under `GET /api/...` (chart data for uPlot).

## Rejected alternatives

A REST API plus a SPA front-end — too much machinery for a single-user desktop app. htmx/Alpine-style partials — deliberately kept out of MVP.
