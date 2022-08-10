Feature: Health Endpoint

  Scenario: Recently active users
    Given I have a bot
    When I send the message "/help"
      And I wait 0.5 seconds
      And I get the server endpoint "/health"
    Then the response body should include "bc_bot_users_active_last{timeframe="1h"}"
      But the response body should not include "bc_bot_users_active_last{timeframe="1h"} 0"
      And the response body should include "bc_bot_users_count"
      But the response body should not include "bc_bot_users_count 0"
