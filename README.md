# Codepix

This is a learning project of a Pix-like transaction system between banks. The Bank API handles the Pix keys and transactions. The Customer API handles the profile and API keys for each bank. The Example Bank API allows the end users to create Pix keys, send and read transactions.

Transactions are started by one bank and processed asynchronously in a 3-way handshake with the receiving bank (started, confirmed/failed, completed/failed), by the use of a persistent event bus (redis streams). Banks listen to this handshake through gRPC streaming endpoints.
Transaction update events are also stored in an event sourced store and published (atomically/outboxed) to other listeners such as the read-only projection, where they can be read and listed by the sender and the receiver.

[Customer API README](customer-api/README.md)

[Bank API README](bank-api/README.md)

[Example Bank API README](example-bank-api/README.md)

## Development

Create a local registry using [/k8s/create-registry.dev.sh](../k8s/create-registry.dev.sh)

Add the start command to a startup script or run it manually before use:
```
sudo podman start registry-codepix
```
