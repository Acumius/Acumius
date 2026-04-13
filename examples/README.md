# Examples

This directory contains runnable examples and request snippets for the Acumius service.

## Health check example

Start the service:

```bash
go run ./cmd/acumius
```

In another terminal:

```bash
curl -s http://localhost:8080/health
```

## Environment variable naming

Runtime environment variables follow the `ACUMIUS_<DOMAIN>_<SETTING>` convention:

- `ACUMIUS_HTTP_ADDR`
- `ACUMIUS_HTTP_READ_TIMEOUT`
- `ACUMIUS_HTTP_WRITE_TIMEOUT`
- `ACUMIUS_HTTP_IDLE_TIMEOUT`
- `ACUMIUS_SHUTDOWN_TIMEOUT`
