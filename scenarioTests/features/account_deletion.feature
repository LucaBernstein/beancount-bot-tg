Feature: account deletion

  Scenario: test template creation
    Given I have a bot
    When I create a test template
    Then I should have templates defined

  Scenario: delete account
    Given I have a bot
    When I create a test template
      And I wait 0.5 seconds
      And I send the message "/config delete_account yes"
    Then 2 messages should be sent back
      And the response should include the message "deleted all of your data stored in the bot"
      And I should not have templates defined
