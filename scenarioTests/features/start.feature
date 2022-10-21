Feature: start bot interaction
  In order to communicate with the bot effectively
  As a beancount user leveraging the bot
  I need to receive an answer from the bot

  Scenario: Answer to start message
    Given I have a bot
    When I send the message "/start"
    Then 2 messages should be sent back
      And the response should include the message "Welcome to this beancount bot!"
      And the response should include the message "/cancel - Cancel any running commands or transactions"

  Scenario: Answer to help message
    Given I have a bot
    When I send the message "/help"
    Then 1 messages should be sent back
      And the response should include the message "/cancel - Cancel any running commands or transactions"

  Scenario: Don't answer to random text in group chat
    Given I have a bot
    When I send the message "this is not a command and not a number and I am not in a tx"
    Then 0 messages should be sent back
