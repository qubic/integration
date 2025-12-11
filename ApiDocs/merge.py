import json
import re

base = {
    "swagger": "2.0",
    "info": {"title": "Qubic RPC API", "description": "API for querying all qubic related data.", "version": "1.0.0"},
    "tags": [],
    "host": "rpc.qubic.org",
    "schemes": [
        "https"
    ],
    "consumes": [
        "application/json"
    ],
    "produces": [
        "application/json"
    ],
    "paths": {},
    "definitions": {},
}

if __name__ == "__main__":
    with open("qubic-http/protobuff/qubic.swagger.json", "r") as f:
        content = f.read()
        content = content.replace("#/definitions/", "#/definitions/QubicHTTP_")
        qubic_http_swagger = json.loads(content)

    with open("archive-query-service/v2/api/archive-query-service/v2/query_services.swagger.json", "r") as f:
        content = f.read()
        content = content.replace("#/definitions/", "#/definitions/QubicQuery_")
        query_service_swagger = json.loads(content)

    base["tags"] += qubic_http_swagger["tags"]

    for path, path_item in qubic_http_swagger["paths"].items():
        base["paths"]["/live/v1" + path] = path_item

    for definition, definition_item in qubic_http_swagger["definitions"].items():
        base["definitions"]["QubicHTTP_" + definition] = definition_item

    for path, path_item in query_service_swagger["paths"].items():
        base["paths"]["/query/v1" + path] = path_item

    for definition, definition_item in query_service_swagger["definitions"].items():
        base["definitions"]["QubicQuery_" + definition] = definition_item

    with open("qubic-combined.swagger.json", "w") as f:
        json.dump(base, f, indent=4)