Feature: Health Endpoint

  Scenario: Open transactions in the cache
    Given I have a bot
    When I send the message "/cancel"
      And I send the message "/simple"
      And I wait 0.5 seconds
      And I get the server endpoint "/health"
      And I wait 0.5 seconds
    Then the response body should include "bc_bot_tx_states_count 1"
    When I send the message "/cancel"
      And I wait 0.5 seconds
      And I get the server endpoint "/health"
      And I wait 0.5 seconds
    Then the response body should include "bc_bot_tx_states_count 0"
