from steps.client import TestBot

if __name__ == "__main__":
    print("Authenticating telegram user and creating session...")
    bot = TestBot()
    with bot.client:
        bot.client.loop.run_until_complete(bot.client.get_me())
