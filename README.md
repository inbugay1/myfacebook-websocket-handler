# MyFacebook Websocket Handler

## Настройки окружения

* LOG_LEVEL - Уровень логирования в приложении. По умолчанию info
* SERVICE_NAME - Имя сервиса. По умолчанию myfacebook
* VERSION - Версия сервиса. По умолчанию version_not_set
* HTTP_INT_PORT - HTTP порт приложения. По умолчанию 9093
* REQUEST_HEADER_MAX_SIZE - максимальный размер header для входящих запросов. По умолчанию 10000 байт.
* REQUEST_READ_HEADER_TIMEOUT_MILLISECONDS - максимальное время отпущенное клиенту на чтение header в мс. По умолчанию
  2000мс.

* RMQ_HOST - Хост для подключения к RabbitMQ. По умолчанию: localhost
* RMQ_HOST_PORT - Порт для подключения к RabbitMQ. По умолчанию: 5672
* RMQ_USERNAME - Имя пользователя. По умолчанию: admin
* RMQ_PASSWORD - Пароль. По умолчанию: admin

* MYFACEBOOK_API_BASE_URL - Адрес монолита. По умолчанию localhost:9092

* CONNECTION_WATCHER_PING_INTERVAL_SECONDS - Интервал в секундах, с которым пингуются сервисы, чтобы проверить состояние соединения. По умолчанию 5 сек.
* CONNECTION_WATCHER_PING_TIMEOUT_SECONDS - Таймаут пинга в секундах. По умолчанию 2 сек.
* CONNECTION_WATCHER_RECONNECT_TIMEOUT_SECONDS - Таймаут на переподключение к сервису в секундах. По умолчанию 2 сек.

## Локальный запуск приложения

Для запуска приложения необходим установленный docker

Version:           24.0.5
API version:       1.43

- Скопируйте .env.example в .env файл.
- Запустите следующие команды по порядку.

```
docker network create myfacebook
make build
make run
```