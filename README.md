# Codepix

[Customer API README](customer-api/README.md)

<br>

## Development

Docker stack is used as the development environment.

To deploy all services use `entry.sh`.

To update and redeploy services also use `entry.sh` (docker caches will be respected).

To remove all Codepix docker artifacts use `clean.sh`.

A service is any root directory containing a `docker-compose.yml` file.
