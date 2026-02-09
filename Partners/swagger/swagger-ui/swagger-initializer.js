window.onload = function() {
  //<editor-fold desc="Changeable Configuration Block">

  var HideInternalTagsPlugin = function() {
    return {
      statePlugins: {
        spec: {
          wrapActions: {
            updateJsonSpec: function(ori) {
              return function(spec) {
                if (spec.paths) {
                  Object.keys(spec.paths).forEach(function(path) {
                    Object.keys(spec.paths[path]).forEach(function(method) {
                      var op = spec.paths[path][method];
                      if (op && op.tags) {
                        op.tags = op.tags.filter(function(t) {
                          return !t.match(/Service$/);
                        });
                      }
                    });
                  });
                }
                if (spec.tags) {
                  spec.tags = spec.tags.filter(function(t) {
                    return !t.name.match(/Service$/);
                  });
                }
                return ori(spec);
              };
            }
          }
        }
      }
    };
  };

  // the following lines will be replaced by docker/configurator, when it runs in a docker-container
  window.ui = SwaggerUIBundle({
    urls: [
      {url: "archive.swagger.json", name: "Qubic RPC Archive Tree"},
      {url: "qubic-http.openapi.yaml", name: "Qubic RPC Live Tree"},
      {url: "stats-api.swagger.json", name: "Qubic Stats API"},
      {url: "query_services.openapi.yaml", name: "Qubic Query V2 Tree"},
    ],
    dom_id: '#swagger-ui',
    deepLinking: true,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl,
      HideInternalTagsPlugin
    ],
    layout: "StandaloneLayout"
  });

  //</editor-fold>
};
