let src = new EventSource("/_hot_reloader_proxy/sse");
src.onmessage = (event) => {
	if (event && event.data === "reload") {
		window.location.reload();
	}
};
