# abr-postcode

[日本語](README.ja.md)

A REST API that converts between Japanese postcodes and machiaza IDs (town/block
identifiers), backed by the Address Base Registry (ABR) maintained by the Digital
Agency of Japan.

> **Note:** Postcodes assigned to individual business offices are not included.

## Quick Start

```bash
make build         # build the binary
./abrp import      # fetch ABR data
./abrp serve       # start the server (http://localhost:8080)
```

`import` converts the nationwide ABR dataset to CSV under `data/`, about 48MB
across roughly 1.36 million rows. It downloads nothing if the DCAT feed's
modification timestamp has not changed since the last run.

## API

| Endpoint | Description |
|----------|-------------|
| `/health` | Health check |
| `/post_code/:post_code` | Machiaza IDs for a postcode |
| `/lg_code/:lg_code` | Municipality for a local government code |
| `/machiaza/:lg_code/:machiaza_id` | Postcodes for a machiaza ID |

```bash
# postcode -> machiaza ID
$ curl -s http://localhost:8080/post_code/1000001 | jq
[
  {
    "lg_code": "131016",
    "machiaza_id": "0006000",
    "pref": "東京都",
    "county": "",
    "city": "千代田区",
    "ward": "",
    "kyoto_st": "",
    "oaza_cho": "千代田",
    "chome": "",
    "koaza": "",
    "machiaza_dist": "",
    "post_code": "1000001"
  }
]

# machiaza ID -> postcodes
$ curl http://localhost:8080/machiaza/131016/0006000
{"lg_code":"131016","machiaza_id":"0006000","pref":"東京都","county":"","city":"千代田区","ward":"","kyoto_st":"","oaza_cho":"千代田","chome":"","koaza":"","machiaza_dist":"","post_codes":["1000001"]}

# municipality
$ curl http://localhost:8080/lg_code/131016
{"lg_code":"131016","pref":"東京都","county":"","city":"千代田区","ward":""}
```

[OpenAPI specification](https://redocly.github.io/redoc/?url=https://raw.githubusercontent.com/digital-go-jp/abr-postcode/main/openapi/openapi.yml)

## Docker

```bash
docker build --target dev --build-arg COMMIT=$(git rev-parse --short HEAD) -t abrp:latest .
docker run --rm -v $(pwd)/data:/app/data abrp:latest import
docker run --rm -p 8080:8080 -v $(pwd)/data:/app/data abrp:latest serve
```

## Configuration

| Environment Variable | Default | Description |
|---|---|---|
| `ABRP_DATA_DIR` | `data/` | CSV data directory |
| `PORT` | `8080` | Server port (`serve`) |
| `GIN_MODE` | `debug` | Set to `release` for production (`serve`) |
| `LOG_LEVEL` | `INFO` | Log level (`DEBUG`, `INFO`, `WARN`, `ERROR`) |
| `LOG_FORMAT` | `json` | Log format (`json`, `text`) |
| `CORS_ALLOW_ORIGIN` | `*` | Allowed CORS origins, comma separated (`serve`) |
| `ABRP_FEED_URL` | ABR DCAT feed | Override the DCAT feed URL (`import`) |
| `ABRP_DATA_URL` | ABR distribution | Override the ABR data ZIP URL (`import`) |

## Data Source

This software uses the [Address Base Registry](https://www.digital.go.jp/policies/base_registry_address) (ABR).

See the [terms of use](https://www.digital.go.jp/policies/base_registry_address_tos)
for how the data may be used.

## License

[MIT](LICENSE)
