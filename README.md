# Integration testing of Go and Kafka using Gnomock

This is a sample program that features `producer`, `consumer`, `reporter` and
`handler` packages. `handler` package exposes an API to submit new events, and
get current submitted event numbers per account.  In the backend, this program
uses Kafka message broker to submit all received events and consume them to
calculate usage statistics.

To test the program, I used [Gnomock](https://github.com/orlangure/gnomock). In
this case, I used it to create a temporary Kafka docker container, run tests
against it using only the external web API (exposed by `handler` package), and
verify the results.

This project includes a sample Github Actions job that runs integration test
using a real Kafka container, and reporting code coverage for all packages.
