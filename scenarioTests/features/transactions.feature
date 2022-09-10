Feature: Transactions

  Scenario: Cancel a transaction if is in none
    Given I have a bot
    When I send the message "/cancel"
    Then 1 messages should be sent back
      And the response should include the message "did not currently have any state or transaction open"

  Scenario: Cancel a valid running transaction
    Given I have a bot
    When I send the message "12.34"
    Then 2 messages should be sent back
    When I send the message "/cancel"
    Then 1 messages should be sent back
      And the response should include the message "currently running transaction has been cancelled"

  Scenario: Record a transaction
    Given I have a bot
    When I send the message "12.34"
    Then 2 messages should be sent back
      And the response should include the message "created a new transaction for you"
      And the response should include the message "enter the **account** the money came **from**"
    When I send the message "FromAccount"
    Then 1 messages should be sent back
      And the response should include the message "enter the **account** the money went **to**"
    When I send the message "ToAccount"
    Then 1 messages should be sent back
      And the response should include the message "enter a **description**"
    When I send the message "any random tx description"
    Then 1 messages should be sent back
      And the response should include the message "Successfully recorded your transaction."
    When I send the message "/list"
    Then 1 messages should be sent back
      And the response should include the message "any random tx description"
