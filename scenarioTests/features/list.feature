Feature: List transactions

  Scenario: List
    Given I have a bot
    When I send the message "/deleteAll yes"
      And I wait 0.4 seconds
    When I send the message "/list"
      And I wait 0.1 seconds
    Then 1 messages should be sent back
      And the response should include the message "You might also be looking for archived transactions using '/list archived'."
    When I create a simple tx with amount 1.23 and desc Test Tx and account:from someFromAccount and account:to someToAccount
      And I wait 0.2 seconds
    When I send the message "/list"
      And I wait 0.2 seconds
    Then 1 messages should be sent back
      And the response should include the message "-1.23 EUR"
    When I send the message "/archiveAll"
      And I wait 0.2 seconds
      And I send the message "/list"
      And I wait 0.3 seconds
    Then 1 messages should be sent back
      And the response should include the message "You might also be looking for archived transactions using '/list archived'."

  Scenario: List dated
    Given I have a bot
    When I send the message "/deleteAll yes"
      And I wait 0.2 seconds
      And I create a simple tx with amount 1.23 and desc Test Tx and account:from someFromAccount and account:to someToAccount
      And I wait 0.3 seconds
    When I send the message "/list dated"
      And I wait 0.2 seconds
    Then 1 messages should be sent back
      And the response should include the message "; recorded on $today "

  Scenario: List archived
    Given I have a bot
    When I send the message "/deleteAll yes"
      And I wait 0.2 seconds
    When I send the message "/list archived"
      And I wait 0.2 seconds
    Then 1 messages should be sent back
      And the response should include the message "You might also be looking for transactions using '/list'."
    When I create a simple tx with amount 1.23 and desc Test Tx and account:from someFromAccount and account:to someToAccount
      And I wait 0.2 seconds
    When I send the message "/list archived"
      And I wait 0.2 seconds
    Then 1 messages should be sent back
      And the response should include the message "You might also be looking for transactions using '/list'."
    When I send the message "/archiveAll"
      And I wait 0.3 seconds
      And I send the message "/list archived"
      And I wait 0.2 seconds
    Then 1 messages should be sent back
      And the response should include the message "-1.23 EUR"

  Scenario: List numbered
    Given I have a bot
    When I send the message "/deleteAll yes"
      And I wait 0.2 seconds
      And I create a simple tx with amount 1.23 and desc Test Tx and account:from someFromAccount and account:to someToAccount
      And I wait 0.3 seconds
      And I create a simple tx with amount 1.23 and desc Another tx and account:from someFromAccount and account:to someToAccount
      And I wait 0.3 seconds
    When I send the message "/list numbered"
      And I wait 0.2 seconds
    Then 1 messages should be sent back
      And the response should include the message "1) $today * "Test Tx""
      And the same response should include the message "2) $today * "Another tx""
    When I send the message "/list rm 1"
      And I wait 0.1 seconds
    Then 1 messages should be sent back
      And the response should include the message "Successfully deleted the list entry specified"
    When I send the message "/list numbered"
      And I wait 0.2 seconds
    Then 1 messages should be sent back
      And the response should include the message "1) $today * "Another tx""
    When I send the message "/list rm 15"
      And I wait 0.1 seconds
    Then 1 messages should be sent back
      And the response should include the message "the number you specified was too high"
