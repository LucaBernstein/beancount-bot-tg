from behave import given, when, then
from behave.api.async_step import async_run_until_complete
from client import TestBot
import os, asyncio
import requests

bot = None
async def getBotSingletonLazy():
    global bot
    if bot is None:
        bot = TestBot()
        await bot.client.connect()
        me = await bot.client.get_me()
        assert me.id is not None

        testChatId = int(os.getenv('TG_CHAT_ID'))
        assert testChatId is not None and testChatId < 0 # is group chat
        assert type(testChatId).__name__ == "int"
        bot.testChatId = testChatId
        print("Will communicate in test chat", testChatId)
    return bot

@given('I have a bot')
@async_run_until_complete
async def step_impl(context):
    context.chat = await getBotSingletonLazy()
    context.testChatId = context.chat.testChatId

@when('I send the message "{message}"')
@async_run_until_complete
async def step_impl(context, message):
    message = await context.chat.client.send_message(context.testChatId, message)
    context.offsetId = message.id
    print("Saving offset message ID", context.offsetId, "of message", message)

async def wait_seconds(seconds):
    await asyncio.sleep(seconds)

@when('I wait {seconds:f} seconds')
@async_run_until_complete
async def step_impl(context, seconds):
    await wait_seconds(seconds)

@then('{count:d} messages should be sent back')
@async_run_until_complete
async def step_impl(context, count):
    await wait_seconds(0.5)
    context.responses = await context.chat.client.get_messages(
        context.testChatId,
        limit=count+10,
        min_id=context.offsetId,
    )
    try:
        assert len(context.responses) == count
    except AssertionError:
        print(len(context.responses), "!=", count)
        assert False

@then('the response should include the message "{message}"')
@async_run_until_complete
async def step_impl(context, message):
    assert len(context.responses) > 0
    message = message.strip()
    response = context.responses.pop(-1) # messages are sorted by creation date from newest to oldest. Take oldest first
    try:
        assert message in response.text
    except AssertionError:
        print("substring", message, "could not be found in", response.text)
        assert False

@when('I get the server endpoint "{endpoint}"')
@async_run_until_complete
async def step_impl(context, endpoint):
    res = requests.get(url="http://localhost:8081"+endpoint)
    context.body = res.text

@then('the response body {shouldShouldNot} include "{include}"')
@async_run_until_complete
async def step_impl(context, shouldShouldNot, include):
    assert shouldShouldNot in ['should', 'should not']
    try:
        cond = include in context.body
        if 'not' in shouldShouldNot:
            cond = not cond
        assert cond
    except AssertionError:
        print("substring", include, "could not be found in", context.body)
        assert False
