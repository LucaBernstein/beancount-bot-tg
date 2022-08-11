Feature: List transactions

  Scenario: List
    Given I have a bot
    When I send the message "/deleteAll yes"
      And I wait 0.2 seconds
    When I send the message "/list"
    Then 1 messages should be sent back
      And the response should include the message "You might also be looking for archived transactions using '/list archived'."
    When I create a simple tx with amount 1.23 and accFrom someFromAccount and accTo someToAccount and desc Test Tx
      And I wait 0.1 seconds
    When I send the message "/list"
    Then 1 messages should be sent back
      And the response should include the message "-1.23 EUR"
    When I send the message "/archiveAll"
      And I wait 0.2 seconds
      And I send the message "/list"
    Then 1 messages should be sent back
      And the response should include the message "You might also be looking for archived transactions using '/list archived'."

  Scenario: List dated
    Given I have a bot
    When I send the message "/deleteAll yes"
      And I wait 0.2 seconds
      And I create a simple tx with amount 1.23 and accFrom someFromAccount and accTo someToAccount and desc Test Tx
      And I wait 0.1 seconds
    When I send the message "/list dated"
    Then 1 messages should be sent back
      And the response should include the message "; recorded on $today "

  Scenario: List archived
    Given I have a bot
    When I send the message "/deleteAll yes"
        And I wait 0.2 seconds
    When I send the message "/list archived"
    Then 1 messages should be sent back
      And the response should include the message "You might also be looking for transactions using '/list'."
    When I create a simple tx with amount 1.23 and accFrom someFromAccount and accTo someToAccount and desc Test Tx
      And I wait 0.1 seconds
    When I send the message "/list archived"
    Then 1 messages should be sent back
      And the response should include the message "You might also be looking for transactions using '/list'."
    When I send the message "/archiveAll"
      And I wait 0.2 seconds
      And I send the message "/list archived"
    Then 1 messages should be sent back
      And the response should include the message "-1.23 EUR"

  Scenario: List numbered
    Given I have a bot
    When I send the message "/deleteAll yes"
      And I wait 0.2 seconds
      And I create a simple tx with amount 1.23 and accFrom someFromAccount and accTo someToAccount and desc Test Tx
      And I wait 0.1 seconds
      And I create a simple tx with amount 1.23 and accFrom someFromAccount and accTo someToAccount and desc Another tx
      And I wait 0.1 seconds
    When I send the message "/list numbered"
    Then 1 messages should be sent back
      And the response should include the message "1) $today * "Test Tx""
      And the same response should include the message "2) $today * "Another tx""
    When I send the message "/list rm 1"
    Then 1 messages should be sent back
      And the response should include the message "Successfully deleted the list entry specified"
    When I send the message "/list numbered"
      And I wait 0.1 seconds
    Then 1 messages should be sent back
      And the response should include the message "1) $today * "Another tx""
    When I send the message "/list rm 15"
    Then 1 messages should be sent back
      And the response should include the message "the number you specified was too high"
