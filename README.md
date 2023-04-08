# Beancount Telegram Bot

Supports you in keeping track of your beancount transactions also if you need to quickly note a transaction down or have no access to your file.

## Features and advantages

* [x] REST API for additional interaction with the bot. E.g. read created transactions from bot automatically.
* [x] Quickly record beancount transactions while on-the-go. Start as simple as entering the amount - no boilerplate
* [x] Suggestions for accounts and descriptions used in the past or configured manually
* [x] Templates with variables and advanced amount splitting for recurring or more complex transactions
* [x] Reminder notifications of recorded transactions with flexible schedule
* [x] Many optional commands, shorthands and parameters, leaving the full flexibility up to you
* [x] Automatically apply tags to transactions, e.g. when on vacation
* [x] Auto-format amount decimal point alignment to match [VSCode Beancount plugin](https://marketplace.visualstudio.com/items?itemName=Lencerf.beancount)
* [x] Bot works in group chat (required to disable [privacy mode](https://core.telegram.org/bots#privacy-mode) with BotFather)
* [x] Code Quality: Unit test covered
* [x] Supported database drivers: `PostgreSQL`, `SQLite`

Check out `/help` in the bot for all available commands and don't forget to configure your bot with `/config`. Just give it a try.

Are you missing a feature, have an idea or a question? Feel free to [create an issue](https://github.com/LucaBernstein/beancount-bot-tg/issues/new/choose).

## Basic Usage

You can use the bot [`@LB_Bean_Bot`](https://t.me/LB_Bean_Bot) ([https://t.me/LB_Bean_Bot](https://t.me/LB_Bean_Bot)) to test/use it directly or to get started quickly.

* `/help`: Get a list of all the available commands
* `/config`: Get an overview of all the available commands for configuring the bot, e.g. default currency, reminder notification schedule, timezone offset, ...
* `/simple`: Create a new questionnaire-based transaction. The transaction date defaults to the current date. To override the date, provide it as parameter, i.e. `/simple 2022-01-24`. To shorten the date parameter, the year and the month can be left out, defaulting to the current year/month, i.e. if the current year is 2022, the following command has the same result: `/simple 01-24`.
  * `123.45`: Entering an amount also starts a new transaction directly, leaving out the step shown above. It also guides you through the rest of the questionnaire of accounts to use for the transactions and so on.
* `/template` or `/t`: Get an overview of the commands to use for managing templates.
  * `/t add myTemplate`: Create a new template under the specified name. In the next step enter the full template. Variables can be inserted as shown in the help text sent back by the bot. This help also contains an example transaction.
  * `/t myTemplate`: Use the template created before. For all variables used, the value to use will be asked. It is possible to call a template with only a subset of its name, as long as it's uniquely identifiable, e.g. `/t myTempl`
* `/cancel`: Cancel either the current transaction recording questionnaire or the creation of a new template.
* `/comment` or `/c`: Add arbitrary text to the transaction list (e.g. for follow-ups). Example: `/c Checking account balance needs to be asserted`. (Note that no comment prefix (`;`) is added automatically, so that by default the entered comment string causes a syntax error in a beancount file to ease follow-up and so that comments don't drown in long transaction lists)
* `/list`: Show a list of all currently recorded transactions (for easy copy-and-paste into your beancount file). The parameter `/list dated` adds a comment prior to each transaction in the list with the date and time the transaction has been added. `/list archived` shows all archived transactions. The parameters can also be used in conjunction, i.e. `/list archived dated`.
  * `/list [archived] numbered`: Shows the transactions list with preceded number identifier. 
  * `/list [archived] rm <number>`: Remove a single transaction from the list
* `/archiveAll`: Mark all currently opened transactions as archived. They can be revisited using `/list archived`.
* `/deleteAll yes`: Permanently delete all transactions, both open and archived.

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
