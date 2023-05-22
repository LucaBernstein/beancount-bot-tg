from behave import given, when, then
from behave.api.async_step import async_run_until_complete
from client import TestBot
import os, asyncio
import requests
from datetime import datetime

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
    await wait_seconds(0.1)

async def bot_send_message(bot: TestBot, chat, message):
    message = await bot.client.send_message(chat, message)
    return message

@when('I send the message "{message}"')
@async_run_until_complete
async def step_impl(context, message):
    context.offsetId = (await bot_send_message(context.chat, context.testChatId, message)).id
    print("Saving offset message ID", context.offsetId, "of message", message)

async def wait_seconds(seconds):
    await asyncio.sleep(seconds)

@when('I wait {seconds:f} seconds')
@async_run_until_complete
async def step_impl(context, seconds):
    await wait_seconds(seconds)

async def collect_responses(bot: TestBot, chat: str, offsetId, count = 1):
    return await bot.client.get_messages(
        chat,
        limit=count+10,
        min_id=offsetId,
    )

@then('{count:d} messages should be sent back')
@async_run_until_complete
async def step_impl(context, count):
    await wait_seconds(0.5)
    context.responses = await collect_responses(context.chat, context.testChatId, context.offsetId)
    try:
        assert len(context.responses) == count
    except AssertionError:
        print(len(context.responses), "!=", count)
        assert False

def replacePlaceholders(msg: str) -> str:
    # date placeholder '$today'
    msg = msg.replace("$today", datetime.today().strftime('%Y-%m-%d'))
    return msg

@then('the{same}response should include the message "{message}"')
@async_run_until_complete
async def step_impl(context, same, message):
    message = replacePlaceholders(message.strip())
    assert same in [" ", " same "]
    if same == " ":
        assert len(context.responses) > 0
        response = context.responses.pop(-1) # messages are sorted by creation date from newest to oldest. Take oldest first
        context.lastResponse = response
    else:
        response = context.lastResponse
    try:
        assert message in response.text
    except AssertionError:
        print("substring '", message, "' could not be found in '", response.text, "'")
        assert False

@then('the response should have a keyboard with {position} entry being "{keyboardEntry}"')
@async_run_until_complete
async def step_impl(context, position, keyboardEntry):
    positionMapping = {
        'the first': 0,
        'the second': 1,
        'any': -1
    }
    assert position in positionMapping
    assert len(context.responses) > 0
    response = context.responses[-1]
    replyKeyboard = response.reply_markup
    # TODO: Add column support if needed
    print("Asserting", replyKeyboard.rows[positionMapping[position]].buttons[0].text, "equals", keyboardEntry)
    assert replyKeyboard.rows[positionMapping[position]].buttons[0].text == keyboardEntry

@when('I get the server endpoint "{endpoint}"')
@async_run_until_complete
async def step_impl(context, endpoint):
    res = requests.get(url="http://localhost:8080"+endpoint, timeout=3)
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

@when('I create a test template')
@async_run_until_complete
async def step_impl(context):
    for message in [
        "/t add mytpl",
        "testTemplate"
    ]:
        context.offsetId = (await bot_send_message(context.chat, context.testChatId, message)).id
        await wait_seconds(0.1)
    print("Saved offset message ID", context.offsetId, "for created test template")

@then('I {shouldShouldNot} have templates defined')
@async_run_until_complete
async def step_impl(context, shouldShouldNot):
    # list templates
    context.offsetId = (await bot_send_message(context.chat, context.testChatId, "/t list")).id
    await wait_seconds(0.5)
    context.responses = await collect_responses(context.chat, context.testChatId, context.offsetId)
    print("Having", len(context.responses), "responses")
    for response in context.responses:
        print(" - ", response.text[:20])
    assert len(context.responses) == 1
    assert shouldShouldNot in ['should', 'should not']
    response = context.responses[0].text
    expectedResponse = "You have not created any template yet. Please see /template"
    cond = expectedResponse != response
    try:
        if 'not' in shouldShouldNot:
            cond = not cond
        assert cond
    except AssertionError:
        print("expected response", expectedResponse, "did not match actual response", response)
        assert False

@when('I create a simple tx with amount {amount} and desc {desc} and account:from {account_from} and account:to {account_to}')
@async_run_until_complete
async def step_impl(context, amount, desc, account_from, account_to):
    for command in ["/cancel", amount, desc, account_from, account_to]:
        context.offsetId = (await bot_send_message(context.chat, context.testChatId, command)).id
        await wait_seconds(0.1)
