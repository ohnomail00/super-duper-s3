# Super Duper S3

**Super Duper S3** – это распределённое хранилище, реализованное на Go, которое обеспечивает загрузку и выгрузку файлов, разбитых на части. Проект использует механизмы хэширования для распределения частей файлов между различными серверами, а также параллельную обработку для повышения производительности при загрузке и выгрузке данных.

## Запуск проекта

### Команда docker-run
```shell
make docker-run
```

Запуск 6 серверов хранения и 1 шлюза. Шлюз доступен по адресу [http://localhost:8000](http://localhost:8000)

### Команда docker-infra
```shell
make docker-infra
```

Запуск 6 серверов хранения, доступных по адресам:
- [http://localhost:8001](http://localhost:8001)
- [http://localhost:8002](http://localhost:8002)
- [http://localhost:8003](http://localhost:8003)
- [http://localhost:8004](http://localhost:8004)
- [http://localhost:8005](http://localhost:8005)
- [http://localhost:8006](http://localhost:8006)

## Примеры HTTP запросов

Файл [http/requests/gateway.http](./http/requests/gateway.http) содержит примеры запросов для [JetBrains HTTP Client](https://www.jetbrains.com/help/idea/http-client-in-product-code-editor.html) для загрузки и выгрузки объектов, а также для добавления серверов хранения.

## Структура проекта

Проект организован по нескольким ключевым компонентам:

- [**cmd/**](./cmd)  
  Содержит entrypoint-ы для разных сервисов:
  - [`cmd/storage/main.go`](./cmd/storage/main.go) – запускает сервис для хранения (Storage Service). Обрабатывает операции PUT/GET для отдельных частей файлов.
  - [`cmd/gateway/main.go`](./cmd/gateway/main.go) – запускает шлюз (Gateway Service), который агрегирует операции загрузки и выгрузки, а также управляет серверной группой.

- [**config/**](./config)  
  Определяет конфигурационные структуры для сервисов с использованием библиотеки [kong](https://github.com/alecthomas/kong).

- [**database/**](./database)  
  Реализует in-memory хранилище для сохранения плана загрузки файла (структура `FileUploadPlan`), который описывает, как части файла распределены между серверами. Определение плана можно найти, например, в [database/mem.go](./database/mem.go).

- [**engine/**](./engine)  
  Содержит бизнес-логику:
  - [**storage/part/**](./engine/storage/part) – логика записи и чтения частей файлов с диска.
  - [**gateway/**](./engine/gateway) – реализация параллельной загрузки (Uploader) и выгрузки (Downloader) файлов.
  - [**hash/**](./engine/hash) – реализация консистентного хэширования для распределения частей файла между серверами.
  - [`server.go`](./engine/server.go) и [`plan.go`](./engine/plan.go) – определение базовых структур.

- [**http/**](./http)  
  Реализует HTTP-слои для взаимодействия с клиентами:
  - [**storage/**](./http/storage) – API для загрузки и получения отдельных частей файлов.
  - [**gateway/**](./http/gateway) – API для загрузки/выгрузки файлов целиком и управления списком серверов.
  - [**clients/**](./http/clients) – HTTP клиенты для взаимодействия между шлюзом и серверами хранения. Подробнее в [http/clients/](./http/clients/).

- [**middlewares/**](./middlewares)  
  Набор middleware для логирования, CORS, трассировки запросов и т.п.

- [**api/**](./api)  
  Определяет обработчики HTTP-запросов для разных эндпоинтов.

- [**tests/**](./tests)  
  Набор интеграционных тестов, демонстрирующих успешные сценарии загрузки и выгрузки, а также добавление серверов в кластер.

## Механизм хэширования

### Консистентное хэширование

В проекте используется механизм консистентного хэширования для равномерного распределения частей файлов по доступным серверам. Основные моменты реализации:

- **Создание кольца**  
  В пакете [`engine/hash`](./engine/hash) реализована структура `Ring`, которая содержит:
  - Слайс `nodes` – отсортированные хэш-значения виртуальных узлов.
  - Массив `nodeMap` – отображение хэш-значений в реальные серверы (структура `Server`).

- **Добавление узлов**  
  При добавлении нового сервера вызывается метод `AddNode`, который:
  - Генерирует несколько виртуальных реплик сервера (количество задаётся параметром `VirtualReplicas`).
  - Для каждой реплики создаётся уникальный ключ, который затем хэшируется с использованием алгоритма [FNV](https://ru.wikipedia.org/wiki/FNV) (для боевого применения лучше выбрать более устойчивую к коллизиям хэш-функцию).
  - Полученные хэш-значения добавляются в слайс `nodes`, после чего список сортируется, что обеспечивает быстрый поиск по кольцу.

- **Поиск сервера для частей файлов**  
  При загрузке части файла генерируется ключ с использованием функции `hash.GeneratePartKey`, который учитывает индекс части и смещение. Затем вызывается метод `GetNode`, который:
  - Вычисляет хэш ключа.
  - Находит ближайший по значению узел в отсортированном слайсе `nodes` (используется бинарный поиск).
  - Возвращает сервер, соответствующий найденному узлу, что гарантирует равномерное распределение нагрузки.

## Параллельная загрузка и выгрузка

### Параллельная загрузка (Uploader)

В пакете [`engine/gateway/upload.go`](./engine/gateway/upload.go) реализован механизм параллельной загрузки, который позволяет обрабатывать поток данных без полного буферизования файла в памяти:

- **Разбиение файла на части**  
  Файл разделяется на заданное число частей (`PartCount`), при этом вместо того, чтобы загружать весь файл в память, сервер работает с потоком данных, используя `io.Reader`.

- **Чтение данных по требованию**  
  Для каждой части данные считываются из ридера непосредственно во время загрузки. Таким образом, сервер не держит весь файл в памяти, а обрабатывает его порционно, снижая нагрузку на ресурсы.

- **Конкурентная отправка**  
  Для каждой части запускается отдельная горутина (с использованием `errgroup`):
  - Сначала выбирается сервер для загрузки части с помощью консистентного хэширования.
  - Затем данные для части считываются из ридера и отправляются через HTTP клиент (созданный с помощью фабрики клиентов [clients.Factory](./http/clients/factory.go)) методом `UploadPart`.
  - После успешной отправки информация о части (структура `PartPlan`) передаётся в канал для дальнейшей агрегации.

- **Агрегация результата**  
  По завершении всех горутин собирается общий план загрузки ([FileUploadPlan](./database/mem.go)), части сортируются по индексу и возвращаются для выгрузки файла.

Такой подход, основанный на ожидании не полного файла, а ридера, позволяет серверу эффективно обрабатывать большие файлы, избегая чрезмерного потребления памяти.

### Параллельная выгрузка (Downloader)

В пакете [`engine/gateway/download.go`](./engine/gateway/download.go) реализован механизм параллельной выгрузки, который позволяет работать с потоком данных, не буферизуя весь файл в памяти:

- **Запуск параллельных загрузок**  
  Для каждой части файла, описанной в [FileUploadPlan](./database/mem.go), запускается отдельная горутина (с использованием `errgroup`):
  - Каждая горутина инициирует вызов метода `DownloadPart` HTTP клиента, созданного через фабрику [clients.Factory](./http/clients/factory.go), для соответствующего сервера.
  - Вместо загрузки полной части файла, сервер возвращает `io.ReadCloser`, который предоставляет поток данных по требованию.

- **Объединение потоков данных**  
  После успешного получения всех потоковых ридеров, данные из них объединяются с помощью функции `io.MultiReader`. Это позволяет сформировать единый поток для передачи данных клиенту, не загружая весь файл в память.

- **Обработка ошибок**  
  Если любая из горутин завершится с ошибкой, происходит немедленная отмена всех оставшихся операций, что обеспечивает корректное завершение процесса выгрузки.

Такой подход позволяет эффективно передавать данные клиенту, используя стриминг, и значительно снижает потребление памяти на сервере при работе с большими файлами.

## Known Issues

В файле [docs/known_issues.md](./docs/known_issues.md) описаны проблемы проекта, которые не вошли в реализацию тестового задания ввиду обширности проекта.  
Для целей "понять образ мышления и умение найти подход к решению задач" достаточно текущей реализации.

Комментарии приветствуются.  
Спасибо за внимание!
