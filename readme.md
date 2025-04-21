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

3. Отправить **https** запрос через прокси. Сертификат для домена `mail.ru`уже сгенерирован и нахожится в [/certs/hosts](/certs/hosts). Сертификаты для других доменов генерируются автоматически

```sh
curl --cacert certs/ca.crt -x http://localhost:8080 https://mail.ru
```

4. Отправить запрос к api серверу

- получить список запросов

```sh
curl localhost:8000/requests -vv
```

- получить запрос

```sh
curl localhost:8000/request/$request_id -vv
```

- повторить запрос

```sh
curl -X POST localhost:8000/repeat/$request_id -vv
```

- просканировать запрос на наличие command injection атаки

```sh
curl -X POST localhost:8000/scan/$request_id -vv
```
