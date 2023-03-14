# Usage as Server

The following sections will detail the steps necessary to run the Dex server.

> Note: For installation instructions, refer [Installation](./installation.md) page.

## Pre-requisites

Dex is stateless and as such has no requirement of stateful components like databases, cache, etc.

But Dex is an orchestrator and needs the following ODPF services:

1. [Entropy](https://github.com/goto/entropy) - Used for deployment and management of applications (e.g., firehose)
2. [Siren](https://github.com/goto/siren) - Used for configuring and managing alerting for tools deployed using Dex.
3. [Shield](https://github.com/goto/shield) - Used for access-control and project level metadata.

## Configurations

All the supported server configurations are documented in the [dex_server.yml](../dex_server.yml) file that is also shipped with the release package.

This configuration file comes with sensible defaults wherever possible.

## Running

Once you tune the configuration file as needed, dex server can be started by running the following command:

```shell
dex server start --config dex_server.yml
```
