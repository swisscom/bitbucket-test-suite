Feature: repository

Background:
* url demoBaseUrl
* header Authorization = call read('classpath:basic-auth.js')

Scenario: get default greeting

    Given path '/rest/api/1.0/projects/', project, '/repos'
    When method get
    Then status 200
    And match response contains { size: '#number' }

