# Image Processor - очередь фоновой обработки изображений
Cервис принимает изображения, кладёт задачу на обработку в очередь (Apache Kafka), в фоне обрабатывает файл (ресайз, наложение водяного знака, миниатюра). 

[Старт](https://github.com/andreyxaxa/Image-Processor?tab=readme-ov-file#%D0%B7%D0%B0%D0%BF%D1%83%D1%81%D0%BA)

## Обзор

Сервис принимает изображение по REST API, генерирует ключ для оригинала, сохраняет изображение в S3(Garage), в единой транзакции - метаданные(ключ, размер, статус, ...) в основную таблицу postgresql, метаданные(payload для обработки) в outbox-таблицу postgresql, в случае неудачи удаляет из S3.

Фоновый outbox-relay воркер по тикеру читает outbox-таблицу, забирает pending ивенты, отправляет их в очередь.

Kafka в качестве controller'а слушает топик, забирает задачи и отдает их на выполнение сервису обработки изображений(disintegration/imaging). Реализован worker pool.

В случае успеха сохраняет обработанное изображение в S3 по новому ключу, обновляет метаданные в БД, коммитит прочитанные из топика сообщения.

Поддерживаемые форматы - .jpg .jpeg .png .webp

Видео запуска и работы - https://drive.google.com/file/d/1KgmaMPTDyw14cH_3X2S7K_lSqsyngBMU/view

- UI - http://localhost:8080/v1
- Документация API - Swagger - http://localhost:8080/swagger
- Конфиг - [config/config.go](https://github.com/andreyxaxa/Image-Processor/blob/main/config/config.go). Читается из `.env` файла.
- Удобная и гибкая конфигурация HTTP сервера - [pkg/httpserver/options.go](https://github.com/andreyxaxa/Image-Processor/blob/main/pkg/httpserver/options.go).
  Позволяет конфигурировать сервер в конструкторе таким образом:
  ```go
  httpServer := httpserver.New(httpserver.Port(cfg.HTTP.Port), httpserver.Prefork(cfg.HTTP.UsePreforkMode))
  ```
  Подобный подход используется и в остальных pkg.
- В слое контроллеров, для REST API применяется версионирование - [internal/controller/restapi/v1](https://github.com/andreyxaxa/Image-Processor/tree/main/internal/controller/restapi/v1).
  Для версии v2 нужно будет просто добавить папку `restapi/v2` с таким же содержимым, в файле [internal/controller/restapi/router.go](https://github.com/andreyxaxa/Image-Processor/blob/main/internal/controller/restapi/router.go) добавить строку:
```go
{
		v1.NewImageRoutes(apiV1Group, img, l) // v1
}

{
		v2.NewImageRoutes(apiV1Group, img, l) // v2
}
```
- Graceful shutdown - [internal/app/app.go](https://github.com/andreyxaxa/Image-Processor/blob/main/internal/app/app.go).

## Запуск

1. Клонируйте репозиторий
2. В корне создайте `.env` файл, скопируйте туда содержимое [env.example](https://github.com/andreyxaxa/Image-Processor/blob/main/.env.example):
   ```
   cp .env.example .env
   ```
3. В корне создайте `garage.toml`, скопируйте туда содержимое [garage.toml.example](https://github.com/andreyxaxa/Image-Processor/blob/main/garage.toml.example):
   ```
   cp garage.toml.example garage.toml
   ```
4. Запуск S3 Garage, выполните, дождитесь запуска:
   ```
   make compose-up-garage
   ```
5. Конфигурация кластера Garage(ёмкость ноды, зона, создание бакета, ключа, определение прав доступа ключа к бакету).

   Для windows:
   ```
   make setup-garage-win
   ```
   Для linux:
   ```
   make setup-garage-lin
   ```
7. Запуск остальных сервисов(postgres, kafka, backend):
   ```
   make compose-up-all
   ```
8. Перейдите на http://localhost:8080/v1 и пользуйтесь сервисом.
<img width="1399" height="976" alt="image" src="https://github.com/user-attachments/assets/1292a068-9b53-4534-8151-f9a7648f3efe" />

- Перейдите на http://localhost:8080/swagger и ознакомьтесь с API, если хотите взаимодействовать с сервисом вручную или из стороннего сервиса.

## Прочие `make` команды
docker compose down:
```
make compose-down
```
docker compose down -v, удаление всех данных Garage:
```
make down-n-clean
```
Зависимости:
```
make deps
```
