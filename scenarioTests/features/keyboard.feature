Feature: Bot suggestions keyboard

  Scenario: Suggest last used values
    Given I have a bot
    When I send the message "/suggestions rm accFrom"
      And I wait 0.2 seconds
      And I send the message "/suggestions add accFrom fromAccount"
      And I wait 0.2 seconds
    When I send the message "1.00"
    Then 2 messages should be sent back
      And the response should include the message "Automatically created a new transaction for you"
      And the response should have a keyboard with the first entry being "fromAccount"
    When I send the message "/cancel"

  Scenario: Last used suggestion appears on top
    Given I have a bot
    When I send the message "/deleteAll yes"
      And I wait 0.2 seconds
      And I send the message "/suggestions rm accFrom"
      And I wait 0.2 seconds
      And I send the message "/suggestions add accFrom fromAccount"
      And I wait 0.2 seconds
      And I create a simple tx with amount 1.23 and accFrom someFromAccount and accTo someToAccount and desc Test Tx
      And I send the message "/list"
    Then 1 messages should be sent back
      And the response should include the message "  someFromAccount                              -1.23 EUR"
    When I send the message "1.00"
    Then 2 messages should be sent back
      And the response should include the message "Automatically created a new transaction for you"
      And the response should have a keyboard with the first entry being "someFromAccount"
    When I send the message "/cancel"
