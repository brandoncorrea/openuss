# Gap Analysis

OpenUSS is an ASTM F3548-21 USS built to do the minimum needed to pass the InterUSS automated test suite. It is deliberately degraded, and it is hardened only when the suite forces a change.

Because it passes, every shortcut it takes shows something the suite does not verify. This document records those findings.

## Summary

- **Result.** OpenUSS passes `suites.astm.utm.f3548_21` on the public US implementation baseline
  ([`baseline_a`](https://github.com/interuss/monitoring/blob/main/monitoring/uss_qualifier/configurations/dev/utm_implementation_us/definitions/baseline_a.libsonnet)):
  - no scenario execution errors
  - no failed checks above Low (12 Low findings, all ["Retrieve pre-existing notifications"](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/nominal_planning/conflict_higher_priority/conflict_higher_priority.md#%E2%84%B9%EF%B8%8F-retrieve-pre-existing-notifications-check))
  - the expected 7 skipped actions

  Its requirements status is **NotFullyVerified** (same as the two InterUSS reference USSs in the same run).
- **The pass does not show conformance.** The suite accepts a USS that:
  - sends notifications with missing or empty required fields
  - never produces conflict notifications for users
  - ignores notifications it receives
  - enforces no authentication on any endpoint it serves
  - ignores time and altitude when detecting conflicts
  - reports an area as cleared without clearing it
- **Where the gaps come from.** Most gaps fall into four kinds:
  1. Checks that are documented but not implemented, or that cannot fail
  2. Checks that verify only a status code
  3. Test data too narrow to exercise conflict geometry
  4. No negative testing against the USS under test

## Findings

Each finding states what the qualifier does and what OpenUSS does that still passes. Severity markers follow the scenario documentation: 🛑 High, ⚠️ Medium, ℹ️ Low.

### A. Checks that are documented but not implemented, or cannot fail

**A1. Notification content is never validated ([SCD0085](https://brandoncorrea.github.io/openuss/requirements/openuss.html#req-astm-f3548-v21-SCD0085)).**
- **Qualifier:**
  - ["Notification data is valid"](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/data_exchange_validation/test_steps/validate_notification_operational_intent.md#-notification-data-is-valid-check) 🛑 is documented, but no code performs it, so it is always _Not tested_.
  - The only check on notifications a tested USS sends is ["Expect Notification sent"](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/data_exchange_validation/test_steps/validate_notification_operational_intent.md#%EF%B8%8F-expect-notification-sent-check) ⚠️. It [matches](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/data_exchange_validation/test_steps/expected_interactions_test_steps.py#L45-L61) on operation, `operational_intent_id` and a 204 from mock_uss.
  - mock_uss [returns 204](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/mock_uss/f3548v21/routes_scd.py#L116-L150) for any body that parses.
- **OpenUSS:**
  - It notifies only the first subscriber.
  - Its body carries `subscriptions: []` ([utm.yaml](https://github.com/astm-utm/Protocol/blob/63ba6c5ff9c9abb93c55fbdf2bffc0b089d10c4a/utm.yaml#L1092-L1119) requires at least one), `details: {}`,
    `version: 0`, an empty `uss_base_url` and `subscription_id`, and no `ovn` ([utm.yaml](https://github.com/astm-utm/Protocol/blob/63ba6c5ff9c9abb93c55fbdf2bffc0b089d10c4a/utm.yaml#L1107-L1113) requires
    the OVN to be populated).
  - It sends no notification when an intent is deleted.
  - The run passes.

**A2. Missing conflict notifications cannot fail [SCD0090](https://brandoncorrea.github.io/openuss/requirements/openuss.html#req-astm-f3548-v21-SCD0090) or [SCD0095](https://brandoncorrea.github.io/openuss/requirements/openuss.html#req-astm-f3548-v21-SCD0095).**
- **Qualifier:**
  - The notification checker [records a latency note](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/notifications_to_operator/notification_checker.py#L150-L222) only when it finds a
    notification.
  - [AggregateChecks](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/aggregate_checks.md) [skips](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/aggregate_checks.py#L217-L218) any
    participant with no latencies.
  - So a USS that never notifies its users is _Not tested_, never _Fail_.
  - A missing `GET /user_notifications` endpoint [yields only](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/notifications_to_operator/notification_checker.py#L47-L58)
    ["Retrieve pre-existing notifications"](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/nominal_planning/conflict_higher_priority/conflict_higher_priority.md#%E2%84%B9%EF%B8%8F-retrieve-pre-existing-notifications-check) ℹ️.
- **OpenUSS:** it has no `user_notifications` endpoint. The result is 12 Low findings, and
  [SCD0090](https://brandoncorrea.github.io/openuss/requirements/openuss.html#req-astm-f3548-v21-SCD0090)/[SCD0095](https://brandoncorrea.github.io/openuss/requirements/openuss.html#req-astm-f3548-v21-SCD0095) are _Not tested_.

**A3. Requirements in the `scd` set that no scenario verifies.**
- **Qualifier:** the [`scd` requirement set](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/requirements/astm/f3548/v21/scd.md) lists these under
  "Automated verification", but no scenario or suite document references them:
  - [GEN0400](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/requirements/astm/f3548/v21.md#common-requirements), [GEN0405](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/requirements/astm/f3548/v21.md#common-requirements)
  - [SCD0055](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/requirements/astm/f3548/v21.md#strategic-conflict-detection-service), [SCD0060](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/requirements/astm/f3548/v21.md#strategic-conflict-detection-service), [SCD0065](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/requirements/astm/f3548/v21.md#strategic-conflict-detection-service), [SCD0070](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/requirements/astm/f3548/v21.md#strategic-conflict-detection-service)
  - [LOG0005-LOG0035](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/requirements/astm/f3548/v21.md#general-logging-requirements-applicable-to-all-roles), [LOG0040](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/requirements/astm/f3548/v21.md#logging-associated-with-operational-intents-applicable-to-strategic-coordination-role), [LOG0045](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/requirements/astm/f3548/v21.md#logging-associated-with-operational-intents-applicable-to-strategic-coordination-role), [LOG0050](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/requirements/astm/f3548/v21.md#logging-associated-with-conformance-monitoring-applicable-to-strategic-coordination-role)
- **Consequence:** no participant can be fully verified for the `scd` capability.

### B. Checks that verify only a status code or the presence of a field

**B1. Receiving notifications is checked only for a 2xx.**
- **Qualifier:** ["Tested USS receives valid notification"](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/subscription_notifications/test_steps/validate_notification_received.md#%EF%B8%8F-tested-uss-receives-valid-notification-check) ⚠️
  [asserts only the HTTP status](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/subscription_notifications/test_steps/validate_notification_received.py#L75-L84).
- **OpenUSS:** it answers 204 and discards the notification. It never uses subscriptions to learn
  about peer intents ([OPIN0015](https://brandoncorrea.github.io/openuss/requirements/openuss.html#req-astm-f3548-v21-OPIN0015), [SCD0080](https://brandoncorrea.github.io/openuss/requirements/openuss.html#req-astm-f3548-v21-SCD0080)).

**B2. `clear_area_requests` is trusted, and the verification is unattributed.**
- **Qualifier:**
  - The flight planning client reads only `outcome.success` (["Area cleared successfully"](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/prep_planners.md#%EF%B8%8F-area-cleared-successfully-check) ⚠️).
  - The independent DSS query,
    ["Area is clear of op intents"](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/clear_area_validation.md#-area-is-clear-of-op-intents-check) 🛑,
    [records no participant](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/clear_area_validation.py#L16-L49).
- **OpenUSS:** it returns `success: true` without clearing anything. In a clean environment this
  never shows, because every scenario deletes its own flights. When a restart leaves operational
  intent references behind, the failure is attributed to no one.

**B3. Deletion results are trusted.**
- **Qualifier:** a delete that returns Completed/Closed passes. The DSS is [re-checked](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/test_steps.py#L172-L199)
  only in [FlightIntentValidation](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/flight_intent_validation/flight_intent_validation.md)'s ["Remove Valid Flight"](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/flight_intent_validation/flight_intent_validation.md#remove-valid-flight-test-step)
  (["Operational intent not shared"](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/flight_intent_validation/flight_intent_validation.md#-operational-intent-not-shared-check) 🛑).
- **OpenUSS:** it reports Completed/Closed without checking whether the DSS delete succeeded.

**B4. Versioning content is not evaluated.**
- **Qualifier:** [GetSystemVersions](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/versioning/get_system_versions.md) and
  [EvaluateSystemVersions](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/versioning/evaluate_system_versions.md)
  [require](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/monitorlib/clients/versioning/client_interuss.py#L33-L61) a 200, a
  `system_identity` field that is present, and a non-empty version. The returned identity is never
  compared with the one requested.
- **OpenUSS:** it returns `system_version: "blah"` and echoes back any requested identity, where
  the [versioning spec](https://github.com/interuss/automated_testing_interfaces/blob/3e6060bd6d5cd665eb5cc60ec7aaeab104547769/versioning/versioning.yaml#L98-L99) says an unknown identity should get 404.

**B5. MakeUssReport accepts an echo.**
- **Qualifier:** [MakeUssReport](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/make_uss_report.md) [requires](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/make_uss_report.py#L129-L165)
  201, a body that parses as `ErrorReport`, and a non-empty `report_id`.
- **OpenUSS:** it echoes the submitted body back with a random id.

**B6. Unimplemented optional response fields silently reduce coverage.**
- **Qualifier:**
  - A missing `as_planned` in flight planning responses
    [skips every "Injection fidelity" check](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/flight_planning/test_steps.py#L337-L356):
    24 _Not tested_ rows for openuss, with no finding raised.
  - `api_name` and `api_version` in `/status`, and `notes`/`includes_advisories`, are never read.

### C. Test data too narrow to exercise conflict detection

**C1. Geometry.**
- **Qualifier:** the standard flight intents never include:
  - an empty area
  - more than one volume
  - a circular outline
  - a self-intersecting polygon
  - a pair separated only in time, or only in altitude

  The closest non-conflicting pair is about 11 m apart horizontally; the tiny-overlap case is a
  genuine ≈1 cm overlap.
- **OpenUSS:**
  - Conflict detection uses only the first volume and ignores time entirely.
  - It compares altitude one-sidedly (`ours.lower < theirs.upper`), so a volume wholly below another counts as intersecting.
  - An empty area would crash it.
  - A circular outline would crash it.
  - Every conflict scenario passes.

**C2. Multiplicity.**
- **Qualifier:** no scenario gives the tested USS more than one foreign subscriber, or more than
  one missing operational intent in a DSS 409.
- **OpenUSS:** it notifies only the first subscriber and fetches only the first missing intent.

**C3. Flight-plan states and execution styles.**
- **Qualifier:**
  - Flights always end with DELETE: monitorlib [has no `Closed` usage state](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/monitorlib/clients/flight_planning/flight_info.py#L182-L189).
  - `execution_style` is always `IfAllowed`.
  - `Contingent` never appears.
  - The only OffNominal intent goes to the control USS, which may answer NotSupported (a
    [documented, accepted outcome](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/nominal_planning/conflict_equal_priority_not_permitted/conflict_equal_priority_not_permitted.md#declare-flight-2-non-conforming-test-step) that halts the scenario).
- **OpenUSS:**
  - It submits `usage_state: Closed` as a live intent.
  - It ignores `execution_style` and `Contingent`.
  - It maps OffNominal to Nonconforming even when the flight is not activated.
  - It rejects creating an intent directly in `Activated`, which F3548 permits.

### D. No fault injection or negative testing aimed at the USS under test

**D1. Authentication on USS endpoints is never probed.**
- **Qualifier:** the negative-auth helpers
  [exist only for the DSS](https://github.com/interuss/monitoring/tree/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/dss/authentication).
  InterUSS's own [`f3548_self_contained`](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/configurations/dev/f3548_self_contained.yaml#L77-L79)
  configuration even requests an empty scope "for authentication test purposes", but nothing uses
  it against a USS.
- **OpenUSS:** `/uss/v1`, `/flight_planning/v1` and `/versioning` serve unauthenticated requests.

**D2. The USS's behaviour under DSS and peer faults cannot be observed.**
- **Qualifier:** the DSS and mock_uss always behave correctly. No scenario makes the DSS return
  4xx/5xx, a malformed 409, a response without an OVN, or unparseable times, and no peer USS
  returns an error to a details request.
- **OpenUSS:**
  - It decodes any non-409 DSS response as success.
  - It ignores DSS delete failures, and then forgets the intent locally, which orphans the OIR.
  - It crashes on a DSS response without an OVN.
  - It treats a peer's error body as operational intent details.
  - It treats any peer error as a conflict. That happens to satisfy
    [DownUSSEqualPriorityNotPermitted](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/off_nominal_planning/down_uss_equal_priority_not_permitted.md), so a Down USS and a transient error are indistinguishable.

  A USS that reports success while the DSS is failing would publish nothing and still pass.

**D3. Robustness against malformed input is never tested.**
- **Qualifier:**
  - It never sends malformed JSON, a non-UUID `flight_plan_id`, or a duplicate `request_id`; the
    client always generates `uuid4`s ([request ID](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/monitorlib/clients/flight_planning/client_v1.py#L52),
    [flight plan ID](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/monitorlib/clients/flight_planning/client_v1.py#L123)).
  - It never requests an unknown `/uss/v1/operational_intents/{id}` ([utm.yaml: 404](https://github.com/astm-utm/Protocol/blob/63ba6c5ff9c9abb93c55fbdf2bffc0b089d10c4a/utm.yaml#L3513-L3518)).
  - It never deletes an unknown flight plan.
- **OpenUSS:**
  - It returns 200 with an empty intent for unknown ids.
  - It panics on a non-UUID flight plan id and on deleting an unknown flight.
  - It does not handle malformed bodies.

## What this baseline does not cover

- **Roles.**
  - Constraint management and processing ([CSTM\*](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/requirements/astm/f3548/v21.md#constraint-management-service), [CSTP\*](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/requirements/astm/f3548/v21.md#constraint-processing-service), [USS0110](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/requirements/astm/f3548/v21.md#constraint-management-service)).
  - CMSA, telemetry and aggregate conformance monitoring ([CMSA\*](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/requirements/astm/f3548/v21.md#conformance-monitoring-for-situational-awareness-service), [ACM\*](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/requirements/astm/f3548/v21.md#aggregate-operational-intent-conformance-monitoring-service), [USS0105,2](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/requirements/astm/f3548/v21.md#strategic-conflict-detection-service)).
  - USS-side availability arbitration: the DSS arbitrates through OVN keys, so a USS never needs
    to read `uss_availability`.
  - Remote ID.

  OpenUSS implements none of these, and nothing in this configuration reveals that.
- **CMSA branches are skipped.** The baseline's `utm_auth` does not include
  `utm.conformance_monitoring_sa`. That skips:
  - the Nonconforming and Contingent cases
  - telemetry
  - every step after ["Declare Flight 2 non-conforming"](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/nominal_planning/conflict_equal_priority_not_permitted/conflict_equal_priority_not_permitted.md#declare-flight-2-non-conforming-test-step)
- **Four DSS scenarios are skipped.** The constraint-reference DSS scenarios ([CRSimple](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/dss/constraint_ref_simple.md) and [CRSynchronization](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/dss/synchronization/constraint_ref_synchronization.md), once per DSS) need `utm.constraint_management`, which the baseline doesn't include. Together with [DatastoreAccess](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/dss/datastore_access.md) ×2 and [PoolInfo](https://github.com/interuss/monitoring/blob/e1e3ad2e5c018eb7afe2eb72c81eceb3ea05b860/monitoring/uss_qualifier/scenarios/astm/utm/dss/pool_info.md), these make up the 7 expected skipped actions.
- **Requirements status.**
  - "Basic SCD" includes DSS requirements that a USS without its own DSS never exercises, even
    when listed as a user of another participant's DSS.
  - Combined with A2 and A3, a flight-planning-only USS is structurally limited to
    NotFullyVerified.

## Test setup

| Item | Value |
|---|---|
| Suite | [`suites.astm.utm.f3548_21`](https://github.com/interuss/monitoring/blob/main/monitoring/uss_qualifier/suites/astm/utm/f3548_21.md) |
| Configuration | Public [`baseline_a`](https://github.com/interuss/monitoring/blob/main/monitoring/uss_qualifier/configurations/dev/utm_implementation_us/definitions/baseline_a.libsonnet) with OpenUSS added as a participant ([`openuss_local.jsonnet`](https://github.com/brandoncorrea/openuss/blob/master/test/config/utm_implementation_us/openuss_local.jsonnet)), mirroring [`utm_implementation_us/environments/local/test_1.jsonnet`](https://github.com/interuss/monitoring/blob/main/monitoring/uss_qualifier/configurations/dev/utm_implementation_us/environments/local/test_1.jsonnet) |
| Participants | [openuss](https://github.com/brandoncorrea/openuss/blob/master/test/config/utm_implementation_us/openuss.libsonnet) (flight planner, versioning; uses uss1's DSS), [uss1](https://github.com/interuss/monitoring/blob/main/monitoring/uss_qualifier/configurations/dev/utm_implementation_us/participants/uss1.libsonnet) (uss1_core + uss1_dss), [uss2](https://github.com/interuss/monitoring/blob/main/monitoring/uss_qualifier/configurations/dev/utm_implementation_us/participants/uss2.libsonnet) (uss2_core + uss2_dss), [mock_uss](https://github.com/interuss/monitoring/blob/main/monitoring/uss_qualifier/configurations/dev/utm_implementation_us/participants/mock_uss.libsonnet) |
| monitoring | [`interuss/monitoring`](https://github.com/interuss/monitoring) `main` |
| Result | Validation passed. openuss: `NotFullyVerified` |
| Reports | [Sequence view](https://brandoncorrea.github.io/openuss/) and [requirements](https://brandoncorrea.github.io/openuss/requirements/openuss.html), from the latest `master` run |

**OpenUSS takes every role.** Three flight planners and no combination selector mean OpenUSS
runs as tested USS, as control USS, and against itself.
