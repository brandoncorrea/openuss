// OpenUSS against the public US implementation baseline (baseline_a), in monitoring's local environment.
//
// This is utm_implementation_us/environments/local/test_1.jsonnet with openuss added as an active participant.
// scripts/automated-tests-us.sh copies this folder to
// monitoring/uss_qualifier/configurations/dev/utm_implementation_us/environments/openuss/, which the relative
// imports below assume.

local baseline = import '../../definitions/baseline_a.libsonnet';
local env_template = import '../../definitions/env_template_a.libsonnet';
local mock_uss = import '../../participants/mock_uss.libsonnet';
local uss2 = import '../../participants/uss2.libsonnet';
local openuss = import 'openuss.libsonnet';

// OpenUSS uses uss1's DSS instance rather than providing its own, so it is listed as a user of that DSS to take
// credit for USS requirements enforced by the DSS.
local uss1 = (import '../../participants/uss1.libsonnet') + {
  local_env+: {
    dss_instances: [
      instance + { user_participant_ids+: ['openuss'] }
      for instance in super.dss_instances
    ],
  },
};

local env_code = 'local_env';
local participants = [uss1, uss2, openuss];
local env = env_template(
  env_code,
  'DummyOAuth(http://oauth.authority.localutm:8085/token,uss_qualifier)',  // Access tokens for the next-higher environment (to retrieve prod versions)
  participants,  // Active participants
  participants,  // Participants for which pre-existing flights should be cleared
  mock_uss,
);

// stop_fast restates baseline_a's value so scripts/automated-tests-us.sh can override it (QUALIFIER_STOP_FAST).
baseline(env) + {
  v1+: {
    test_run+: {
      execution+: {
        stop_fast: true,
      },
    },
  },
}
