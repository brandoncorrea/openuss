# Open USS

An open-source USS, built for the US UTM cohort and developed against the
[InterUSS monitoring](https://github.com/interuss/monitoring) test suite.

> **Status: early.** No conformance claims yet.

## Scope

[ASTM F3548-21](https://store.astm.org/f3548-21.html), all three USS roles: strategic coordination, constraint management, and availability arbitration. [ASTM F3411](https://store.astm.org/f3411-22a.html) (Network Remote ID) is scoped for future work.

## Goals

1. Enable US operators to quickly onboard, clear [Gate 1](https://github.com/utmimplementationus/getstarted#how-to-get-started), and confidently proceed through Gates 2 and 3.
2. Exercise the InterUSS [DSS](https://github.com/interuss/dss) and [automated test suite](https://github.com/interuss/monitoring) and contribute feedback upstream.

## Decisions

### US Cohort

This project is intended to target _only_ the US UTM cohort (FAA UTM Implementation); EU U-Space requirements are out of scope.

If you need to satisfy EU U-Space requirements, I recommend you look into the [OpenUTM](https://github.com/openutm/) project.

### Go

Go has been chosen for this project primarily because the DSS speaks Go. This allows the project to take advantage of existing InterUSS tooling and keeps it familiar for those who already know InterUSS.

### Naïve, then Hardening

The first pass is getting green in the automated test suite by doing the bare minimum: an intentionally degraded service. This helps expose any gaps that may exist in the automated test suite.

The idea is to take a similar philosophy from the [Three Laws of TDD](https://blog.cleancoder.com/uncle-bob/2014/12/17/TheCyclesOfTDD.html): OpenUSS writes no more code than is needed to pass the currently failing check.

There may be cases where a production-level USS needs behavior the automated test suite doesn't cover, or covers ambiguously. In these cases, a finding may be filed upstream _only_ if it adds value to the InterUSS project.

After getting a _technically working_ USS shell that passes the test suite, then the system is ready for hardening. The goal here is to make the system production-ready.

## Development

    make build    # Build the binary
    make test     # Run unit tests
    make image    # Build the docker image
    make run      # Start a docker service
    make stop     # Stop the docker service

## License

See [LICENSE](LICENSE).
