# Beancount Telegram Bot

Supports you in keeping track of your beancount transactions also if you need to quickly note a transaction down or have no access to your file.

## Usage

You can use the bot [`@LB_Bean_Bot`](https://t.me/LB_Bean_Bot) to test/use it directly or to get started quickly.

Steps to get started after contacting the bot:

1. Start a conversation with the bot (usually `/start`).
1. If needed, the bot will lead you through any additional setup needed
1. You can type `/help` to see the commands available to control the bot (e.g. to add accounts, record transactions, change account settings, ...)

## Installation (self-hosted)

Check the [`SETUP.md`](./SETUP.md) guide for more information about hosting the bot backend application.

## Development

To setup a local development environment you need a running database.
The following commands help you to get started:

```bash
# Prerequisites
docker network create localdev
docker run --rm -d -p 5432:5432 --network localdev --name postgres_localdev -e POSTGRES_PASSWORD=password postgres
docker run --rm -d -p 8090:8080 --network localdev --name adminer adminer

# Fill your local .env file (template from .env.sample) with the respective values

# Backend
env $(cat .env | xargs) go run .
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

To update modules:

```
go get -u
go mod tidy
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

[Licensed under MIT license](./LICENSE)
