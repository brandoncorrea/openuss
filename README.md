# Open USS

An open-source, readily-deployable USS, built to conform to
[ASTM F3548-21](https://www.astm.org/f3548-21.html) and verified against the
[InterUSS](https://github.com/interuss) monitoring test suite.

> **Status: early.** No conformance claims yet.

## Scope

ASTM F3548-21, all three USS roles: strategic coordination, constraint management, and availability arbitration. ASTM F3411 (Network Remote ID) is not in scope today.

## Development

    make build    # Build the binary
    make test     # Run unit tests
    make image    # Build the docker image
    make run      # Start a docker service
    make stop     # Stop the docker service

## License

See [LICENSE](LICENSE).
