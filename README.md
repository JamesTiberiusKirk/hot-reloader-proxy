# HRP (Hot Reloader Proxy)
Inspiration (and some of the code) borrowed from a feature in [templ](https://github.com/a-h/templ) cli.

This is a proxy which is meant to be used in development of a website. It works as follows:
- On every `text/html` which has the body close tag `</body>`
- A script tag will be injected
    - This script tag will be trying to subscribe to server sent event endpoint served by the proxy
- A file watcher is then started, generating SSE event whcih will trigger the js to refresh the page


More details and usage is to be annouced.

