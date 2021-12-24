# Setup for self-hosting this beancount bot and webapp on your own infrastructure

## Create Telegram bot

1. Contact `@BotFather` on Telegram
1. Send message: `/newbot`
1. Give your bot a name.
1. Copy and store API key and create a `.env` file in the main directory and write it into `BOT_API_KEY` variable
1. You can enrich the bot with more metadata as you like

## Setup backend

Check out `docker-compose.yml` as an example on how to run the app. It uses the Docker image builds from GitHub Actions.

```bash
docker-compose up -d
```

## Backup and restore

```bash
# Backup
docker exec bc_db pg_dump -U beancount beancount > beancount-db-dump.sql

# Restore
docker-compose up -d db
docker exec -i bc_db psql --set ON_ERROR_STOP=on -U beancount beancount < beancount-db-dump.sql
docker-compose up -d
```
