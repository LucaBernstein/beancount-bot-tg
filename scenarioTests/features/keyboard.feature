Feature: Bot suggestions keyboard

  Scenario: Suggest last used values
    Given I have a bot
      And I have no open transaction
    When I send the message "/suggestions rm account:from"
      And I wait 0.2 seconds
      And I send the message "/suggestions add account:from fromAccount"
      And I wait 0.3 seconds
    When I send the message "1.00"
      And I wait 0.4 seconds
    Then 2 messages should be sent back
      And the response should include the message "Automatically created a new transaction for you"
    When I send the message "unimportant_description"
    Then 1 messages should be sent back
      And the response should have a keyboard with the first entry being "fromAccount"
    When I send the message "/cancel"

  Scenario: Last used suggestion appears on top
    Given I have a bot
      And I send the message "/config currency EUR"
    When I send the message "/deleteAll yes"
      And I wait 0.2 seconds
      And I send the message "/suggestions rm account:from"
      And I wait 0.2 seconds
      And I send the message "/suggestions add account:from fromAccount"
      And I wait 0.2 seconds
      And I create a simple tx with amount 1.23 and desc Test Tx and account:from someFromAccount and account:to someToAccount
      And I wait 0.3 seconds
      And I send the message "/list"
    Then 1 messages should be sent back
      And the response should include the message "  someFromAccount                              -1.23 EUR"
    When I send the message "1.00"
      And I wait 0.1 seconds
    Then 2 messages should be sent back
      And the response should include the message "Automatically created a new transaction for you"
    When I send the message "unimportant_description"
    Then 1 messages should be sent back
      And the response should have a keyboard with the first entry being "someFromAccount"
    When I send the message "/cancel"
