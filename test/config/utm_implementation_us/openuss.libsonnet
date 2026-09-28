{
  participant_id: 'openuss',
  participant_requirements: 'Basic SCD',
  local_env: {
    flight_planner: {
      participant_id: 'openuss',
      v1_base_url: 'http://openuss.uss5.localutm:8080/flight_planning/v1',
    },
    test_env_version_provider: {
      participant_id: 'openuss',
      interuss: {
        base_url: 'http://openuss.uss5.localutm:8080/versioning',
      },
    },
    prod_env_version_provider: {
      participant_id: 'openuss',
      interuss: {
        base_url: 'http://openuss.uss5.localutm:8080/versioning',
      },
    },
  },
}
