Feature: bitbucket repositories

  Scenario: user creates a repository
    Given the repository test_repo doesnt exist
    When I create repository test_repo
    Then repository test_repo should be accessible
