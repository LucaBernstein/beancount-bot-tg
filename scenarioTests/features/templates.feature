Feature: Templates

  Scenario: Create a template
    Given I have a bot
    When I send the message "/t add mytpl"
    Then 1 messages should be sent back
      And the response should include the message "provide a full transaction template"
    When I send the message "${description} weird template ${-amount/3}"
    Then 1 messages should be sent back
      And the response should include the message "Successfully created your template"
    When I send the message "/t list"
    Then 1 messages should be sent back
      And the response should include the message "mytpl"
    When I send the message "/t my"
    Then 2 messages should be sent back
      And the response should include the message "Creating a new transaction from your template 'mytpl'"
      And the response should include the message "Please enter the *amount*"
    When I send the message "15,15"
    Then 1 messages should be sent back
      And the response should include the message "Please enter a **description**"
    When I send the message "some description"
    Then 1 messages should be sent back
      And the response should include the message "Successfully recorded your transaction."
    When I send the message "/list"
    Then 1 messages should be sent back
      And the response should include the message "some description weird template                -5.05 EUR"
    When I send the message "/t rm mytpl"
    Then 1 messages should be sent back
      And the response should include the message "Successfully removed your template 'mytpl'"
    When I send the message "/t list"
    Then 1 messages should be sent back
      And the response should include the message "You have not created any template yet."
