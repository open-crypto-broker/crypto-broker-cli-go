# Crypto Broker CLI

## Usage

The Crypto Broker CLI is a CLI-type example program written in Golang that allow users to interact with a Crypto Broker Server using crypto-broker-client-go library.

## Development

This section covers how to contribute to the project and develop it further.

### Pre-requisites

A version of [Golang](https://go.dev/doc/install) > 1.25 installed on your local machine is required in order to run it locally from terminal. For building the Docker image, you need to have Docker/Docker Desktop.

For running the commands using the `Taskfile` tool, you need to have Taskfile installed. Please check the documentation on [how to install Taskfile](https://taskfile.dev/installation/). If you don't have Taskfile support, you can directly use the commands specified in the Taskfile on your local terminal, provided you meet the requirements.

To contribute to this project please configure the custom githooks for this project:

```bash
git config core.hooksPath .githooks
```

This commit hook will make sure the code follows the standard formatting and keep everything consistent.

### Building

#### Compiling the Go binaries

For testing the application, you can build the local CLI with the following command:

```shell
task build
```

This will also save a checksum of all the file `sources` in the Taskfile cache `.task`.
This means that, if no new changes are done, re-running the task will not build the app again.

#### Building the Docker image

For building the image for local use, you can use the command:

```shell
task build-docker [TAG=opt]
```

The TAG argument is optional and will apply a custom image tag to the built images. If not specified, it defaults to `latest`. This will create a local image tagged as `server_app:TAG`, which will be saved in your local Docker repository. If you want to modify or append args to the build command, please refer to the one from the Taskfile.

### Testing

To invoke local CI pipeline run

```shell
task ci
```

You can do a local end2end testing of the application yourself with the provided CLI. To run the CLI, you first need to have the [Crypto Broker server](https://github.com/open-crypto-broker/crypto-broker-server/) running in your Unix localhost environment. Once done, you can run one of the following in another terminal:

```shell
task test-hash
# or
task test-sign
```

For the sign command you need to have the [deployment repository](https://github.com/open-crypto-broker/crypto-broker-deployment) in the same parent directory as this repository. Check the command definitions in the `Taskfile` file to run your own custom commands.

More thorough testing is also provided in the deployment repository. The same pipeline will run in GitHub Actions when submitting a Pull Request, so it is recommended to also clone and run the testing of the deployment repository.

## Support, Feedback, Contributing

Contribution and feedback are encouraged and always welcome. For more information about how to contribute, the project structure, as well as additional contribution information, see our [Contribution Guidelines](CONTRIBUTING.md).

## Security / Disclosure

If you find any bug that may be a security problem, please follow our instructions at [in our security policy](./SECURITY.md) on how to report it. Please do not create GitHub issues for security-related doubts or problems.

## Code of Conduct

We as members, contributors, and leaders pledge to make participation in our community a harassment-free experience for everyone. By participating in this project, you agree to abide by its [Code of Conduct](https://github.com/open-crypto-broker/.github/blob/main/CODE_OF_CONDUCT.md) at all times.

## Licensing

Copyright 2025 SAP SE or an SAP affiliate company and Open Crypto Broker contributors. Please see our [LICENSE](LICENSE) for copyright and license information.
