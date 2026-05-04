# Yartime

Parental control system for Linux.

## Structure

- `server/` — REST API server
- `client/` — daemon on child's machine
- `scripts/` — installation scripts

## Quick Start

### Server
```bash
cd server && go run ./cmd/server
```

### Client
```bash
cd client && go run ./cmd/client --server-url=http://localhost:8080
```

## Install on target machine

```bash
sudo ./scripts/install.sh --server-url=http://your-server:8080 --client-id=child1
```
