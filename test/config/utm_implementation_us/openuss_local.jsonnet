// OpenUSS against the public US implementation baseline (baseline_a).

local baseline = import '../../definitions/baseline_a.libsonnet';
local env_template = import '../../definitions/env_template_a.libsonnet';
local mock_uss = import '../../participants/mock_uss.libsonnet';
local uss2 = import '../../participants/uss2.libsonnet';
local openuss = import 'openuss.libsonnet';

// OpenUSS uses uss1's DSS instance rather than providing its own
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

// stop_fast restates baseline_a's value so scripts/automated-tests.sh can override it (QUALIFIER_STOP_FAST).
baseline(env) + {
  v1+: {
    test_run+: {
      execution+: {
        stop_fast: true,
      },
    },
  },
}
