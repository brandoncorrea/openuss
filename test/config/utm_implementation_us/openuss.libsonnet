// OpenUSS as a participant in monitoring's utm_implementation_us configurations.
// Shape follows monitoring/uss_qualifier/configurations/dev/utm_implementation_us/participants/uss1.libsonnet.
{
  // ID of participant
  participant_id: 'openuss',

  // Set of requirements this participant wants to satisfy
  participant_requirements: 'Basic SCD',

  // Definition of this participant's systems in the local environment
  local_env: {
    // Means by which to interact with the participant as a flight planner
    flight_planner: {
      participant_id: 'openuss',
      v1_base_url: 'http://openuss.uss5.localutm:8080/flight_planning/v1',
    },

    // Means by which to obtain this participant's software version in the test environment
    test_env_version_provider: {
      participant_id: 'openuss',
      interuss: {
        base_url: 'http://openuss.uss5.localutm:8080/versioning',
      },
    },

    // Means by which to obtain this participant's software version in the prod environment
    // (there is no production OpenUSS, so this is the test instance, as uss1 does locally)
    prod_env_version_provider: {
      participant_id: 'openuss',
      interuss: {
        base_url: 'http://openuss.uss5.localutm:8080/versioning',
      },
    },
  },
}
