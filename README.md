# DOCKER

Для подключения к бд и запуска миграций нужно запустить
```
docker compose up
```

# Линтер

Для успешной работы линтера должны быть локально установлены:
1) Docker: https://docs.docker.com/get-docker/
2) jq: https://jqlang.github.io/jq/download/

```
Запуск golangci-lint: make golangci-lint-run
```

```
Удаления всех сгенерированный golangci-lint файлов: make golangci-lint-clean
```

# go-musthave-diploma-tpl

Шаблон репозитория для индивидуального дипломного проекта курса «Go-разработчик»

# Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без
   префикса `https://`) для создания модуля

# Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m master template https://github.com/yandex-praktikum/go-musthave-diploma-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/master .github
```

Затем добавьте полученные изменения в свой репозиторий.
