Feature:
  In order to test inputs and ouputs with the Nats provider
  As a developer
  I need to be able to test some features

Scenario: Succesfuly publishes action on receipt of a message
    Given Nats is running 
    When I receive a message
    Then I expect a action message to be published 
