# Beancount Telegram Bot

Supports you in keeping track of your beancount transactions also if you need to quickly note a transaction down or have no access to your file.

## Links to access bot and webinterface

* [Webinterface (coming soon...)]()
* [Telegram Bot @LB_Bean_Bot](https://t.me/LB_Bean_Bot)

## Getting started

1. Send the Telegram bot any initial message to get it started.
1. If needed, the bot will lead you through any additional setup needed
1. You can type `/help` to see the commands available to control the bot (e.g. to add accounts, record transactions, change account settings, ...)

## Setup self-hosted

Check the [`SETUP.md`](SETUP.md) guide for more information about hosting the bot backend application.

## Setup local dev environment

To start the db and the backend server:

```bash
# Prerequisites
docker network create localdev
docker run --rm -d -p 5432:5432 --network localdev --name postgres_localdev -e POSTGRES_PASSWORD=password postgres
docker run --rm -d -p 8090:8080 --network localdev --name adminer adminer

# Backend
POSTGRES_HOST=localhost POSTGRES_PASSWORD=password go run .
```

Install [`Air`](https://github.com/cosmtrek/air) to hot-reload while developing

```bash
curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

air -v

# or, if you don't have air in your PATH:

$(go env GOPATH)/bin/air -v
```

To add module requirements and sums:

```
go mod tidy
```
