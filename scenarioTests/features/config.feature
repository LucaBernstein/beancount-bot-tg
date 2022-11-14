Feature: config user settings

  Scenario: test config command handling
    Given I have a bot
    When I send the message "/config omit_slash on"
    Then 1 messages should be sent back
      And the response should include the message "has successfully been turned on"
