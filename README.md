# GO-CLOUD-Customer (RPC)

This service exposes RPC methods over TCP using `net/rpc` and runs with mutual TLS (mTLS).

## RPC Methods

- `HealthCheckService.Ping`
- `CustomerService.RegisterCustomerUser`

## Config

Config is loaded from:

1. `configs/appsettings.json`
2. `configs/appsettings.{ENV}.json` (default `ENV=Development`)
3. Environment variable overrides
4. Optional Azure Key Vault (replaces values that are still `FROM_KEY_VAULT`)

## TLS

The server expects these files in the working directory:

- `ca-{ENV}.crt`
- `ca-{ENV}.key`

It enforces client certificate validation: `tls.RequireAndVerifyClientCert`.
