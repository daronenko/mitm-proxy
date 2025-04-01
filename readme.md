# MITM Proxy

## Getting Started

1. Запуск прокси на порту `8080`

```sh
./start.sh
```

2. Отправить **http** запрос через прокси

```sh
curl -x http://localhost:8080 http://mail.ru
```

1. Отправить **https** запрос через прокси. Сертификат для домена `mail.ru`уже сгенерирован и нахожится в [/certs/hosts](/certs/hosts). Сертификаты для других доменов генерируются автоматически

```sh
curl --cacert certs/ca.crt -x http://localhost:8080 https://mail.ru
```
