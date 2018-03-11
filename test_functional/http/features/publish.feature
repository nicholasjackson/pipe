Feature:
  In order to test inputs and ouputs with the HTTP provider
  As a developer
  I need to be able to test some features

Scenario: Succesfuly listens to HTTP endpoint and publishes action on call
    Given Pipe is running and configured
    When I call an http endpoint
    Then I expect pipe to make an outbound http call 
