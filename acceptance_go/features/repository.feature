Feature: bitbucket repositories

  Scenario: user creates a repository
    Given the repository test_repo doesnt exist
    When I create repository test_repo
    Then repository test_repo should be accessible

  Scenario: user commits and pushes a file
	Given the repository test_repo exists
	When clone the test_repo
	And commit a file
	And push to remote
	Then the commit should be visible in repository test_repo
