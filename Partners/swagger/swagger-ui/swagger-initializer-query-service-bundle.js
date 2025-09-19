window.onload = function() {
    //<editor-fold desc="Changeable Configuration Block">

    // the following lines will be replaced by docker/configurator, when it runs in a docker-container
    window.ui = SwaggerUIBundle({
        urls: [
            {url: "query_services.swagger.json", name: "Qubic Query V2 Tree"},
            {url: "qubic-http.swagger.json", name: "Qubic RPC Live Tree"},
            {url: "stats-api.swagger.json", name: "Qubic Stats API"},
        ],
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
            SwaggerUIBundle.presets.apis,
            SwaggerUIStandalonePreset
        ],
        plugins: [
            SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout"
    });

    //</editor-fold>
};
