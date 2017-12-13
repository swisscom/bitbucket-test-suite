Feature: repository

Background:
* url demoBaseUrl

Scenario: create repo, delete repo

Given path '/rest/api/1.0/projects/', project, '/repos'
    And request '{ "name": "test_repo", "scmId": "git", "forkable": true }'
    And header Content-Type = 'application/json'
    And header Authorization = call read('classpath:basic-auth.js')
    When method post
    Then status 201
    And match response contains { slug: 'test_repo' }

Given path '/rest/api/1.0/projects/', project, '/repos/test_repo'
    And header Content-Type = 'application/json'
    And header Authorization = call read('classpath:basic-auth.js')
    When method delete
    Then status 202

