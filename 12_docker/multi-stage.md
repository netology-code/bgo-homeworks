# Multi-Stage Build

Multi-Stage Build подразумевает, что мы в одном контейнере собираем наше приложение и используем результаты сборки в другом контейнере.

В нашем случае это будет выглядеть вот так:
1. Берём образ, в котором установлена нужная нам версия Go
1. Создаём на базе него свой образ, в который копируем все файлы из нашего репо
1. Запускаем из образа контейнер (на самом деле, это сделает сам Docker), в котором и производим сборку
1. Результаты сборки копируем в новый образ

Мы сохраним всё в отдельном файле `Dockerfile.multistage`, чтобы отличать файлы друг от друга.

```dockerfile
FROM golang:1.14-alpine AS build
ADD . /app
ENV CGO_ENABLED=0
WORKDIR /app
RUN go build -o bank ./cmd/bank

FROM alpine:3.7
COPY --from=build /app/bank /app/bank
ENTRYPOINT ["/app/bank"]
```

Давайте рассмотрим те моменты, которые мы ещё не рассматривали:
* `AS build` - даём имя образу, собираемому на данном этапе (чтобы использовать его на следующих этапах сборки)
* `ADD . /app` - добавляем всё содержимое текущего каталога `.` (а поскольку сам файл будет в корне репо, то все файлы) в каталог `/app` в образе
* `ENV` - устанавливаем значение переменной окружения 
* `WORKDIR` - устанавливаем рабочий каталог для следующих команд
* `RUN` - запускаем сборку
* `COPY` - копируем файл из образа, созданного на одной из предыдущих стадий

Вот, в принципе, и всё.

Теперь попробуем это собрать:
```shell script
docker build -f Dockerfile.multistage -t multibank .
docker container run -p 9999:9999 multibank
```

Флаг `-f` указывает на то, какой `Dockerfile` мы собираемся использовать (по умолчанию используется `Dockerfile`)

Теперь попробуем включить это в GitHub Actions:

```yaml
name: Go

on:
  push:
    branches: [ master ]
  pull_request:
    branches: [ master ]

jobs:

  build:
    name: Build
    runs-on: ubuntu-latest
    steps:

      - name: Set up Go 1.x
        uses: actions/setup-go@v2
        with:
          go-version: 1.14
        id: go

      - name: Check out code into the Go module directory
        uses: actions/checkout@v2

      - name: Get dependencies
        run: go get -v -t -d ./...

      - name: Build
        run: go build -v ./...

      - name: Test
        run: go test -v ./...

      - name: Build binary for docker image
        run: go build -v -o bank ./cmd/bank

      - name: List
        run: ls -la

      - name: Push to GitHub Packages
        uses: docker/build-push-action@v1
        with:
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
          registry: docker.pkg.github.com
          repository: netology-code/bgo-docker/bank
          tag_with_ref: true
      
      # добавили эту часть    

      - name: Build from Docker multistage
        run: docker build -f Dockerfile.multistage -t multibank .
```

На самом деле, интересного тут не много, но позволит нам проверить, что образ собирается (правда мы его не тестируем, но об этом разговор ещё впереди).

Здесь, конечно же, всё зависит от того, насколько активно в организации используется Docker. Если всё построено на нём, то можно убрать всё, что не связано с Docker и перенести это внутрь Docker (сборку приложения, автотесты, сборку образа Docker).
