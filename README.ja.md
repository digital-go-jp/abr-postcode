# abr-postcode

[English](README.md)

日本の郵便番号と町字IDの相互変換を行う REST API。
デジタル庁が整備するアドレス・ベース・レジストリ (ABR) を利用する。

> **Note:** 事業所の個別郵便番号は含まれていない。

## Quick Start

```bash
make build         # バイナリをビルド
./abrp import      # ABR データ取得
./abrp serve       # サーバー起動 (http://localhost:8080)
```

`import` は ABR の全国データを CSV に変換して `data/` に置く。生成物は約 48MB、約 136 万行。DCAT Feed の更新日時が前回取得時から変わっていなければ、ダウンロードせずに終了する。

## API

| Endpoint | Description |
|----------|-------------|
| `/health` | ヘルスチェック |
| `/post_code/:post_code` | 郵便番号から町字IDを取得 |
| `/lg_code/:lg_code` | 市区町村情報を取得 |
| `/machiaza/:lg_code/:machiaza_id` | 町字IDから郵便番号一覧を取得 |

```bash
# 郵便番号 → 町字ID
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

# 町字ID → 郵便番号
$ curl http://localhost:8080/machiaza/131016/0006000
{"lg_code":"131016","machiaza_id":"0006000","pref":"東京都","county":"","city":"千代田区","ward":"","kyoto_st":"","oaza_cho":"千代田","chome":"","koaza":"","machiaza_dist":"","post_codes":["1000001"]}

# 市区町村情報
$ curl http://localhost:8080/lg_code/131016
{"lg_code":"131016","pref":"東京都","county":"","city":"千代田区","ward":""}
```

[OpenAPI 仕様書](https://redocly.github.io/redoc/?url=https://raw.githubusercontent.com/digital-go-jp/abr-postcode/main/openapi/openapi.yml)

## Docker

```bash
docker build --target dev --build-arg COMMIT=$(git rev-parse --short HEAD) -t abrp:latest .
docker run --rm -v $(pwd)/data:/app/data abrp:latest import
docker run --rm -p 8080:8080 -v $(pwd)/data:/app/data abrp:latest serve
```

## Configuration

| Environment Variable | Default | Description |
|---|---|---|
| `ABRP_DATA_DIR` | `data/` | CSV データディレクトリ |
| `PORT` | `8080` | サーバーポート (`serve`) |
| `GIN_MODE` | `debug` | `release` で本番モード (`serve`) |
| `LOG_LEVEL` | `INFO` | ログレベル (`DEBUG`, `INFO`, `WARN`, `ERROR`) |
| `LOG_FORMAT` | `json` | ログ形式 (`json`, `text`) |
| `CORS_ALLOW_ORIGIN` | `*` | CORS 許可オリジン。カンマ区切りで複数指定できる (`serve`) |
| `ABRP_FEED_URL` | ABR DCAT Feed | DCAT Feed URL の上書き (`import`) |
| `ABRP_DATA_URL` | ABR 配信元 | ABR データ ZIP の URL 上書き (`import`) |

## データソース

本ソフトウェアは[アドレス・ベース・レジストリ](https://www.digital.go.jp/policies/base_registry_address)（ABR）を利用している。

データの利用については[利用規約](https://www.digital.go.jp/policies/base_registry_address_tos)を参照。

## License

[MIT](LICENSE)
