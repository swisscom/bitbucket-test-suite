Feature: bitbucket repositories

  Scenario: user commits and pushes a file
    Given the repository test_repo doesnt exist
    Given I create repository test_repo
    Given repository test_repo is accessible
    When clone the test_repo
    And commit a file
    And push to remote
    Then the commit should be visible in repository test_repo
