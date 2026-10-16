# Plain MPA: PRG, method spoofing, HTML dialogs

The app is a server-rendered MPA with no SPA framework. All mutations are POST forms followed by a redirect (PRG); deletes POST to the resource URL with a `_method=DELETE` hidden field that a method-spoofing middleware rewrites to DELETE (a real `DELETE` route is registered); confirmations and create/edit forms are `<dialog>` modals opened by delegated clicks via `data-dialog`; validation errors re-render the page with the modal reopened. The only JSON is under `GET /api/...` (chart data for uPlot).

## Rejected alternatives

A REST API plus a SPA front-end — too much machinery for a single-user desktop app. htmx/Alpine-style partials — deliberately kept out of MVP.
