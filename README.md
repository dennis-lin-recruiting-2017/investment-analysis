# App Template

Version: 0.0.4

## Build targets

- `make frontend` builds the React frontend into `frontend/dist`
- `make backend` builds the Go server into `dist/app-template`
- `make build` runs both steps

## Run

```bash
go run . --ui-port 3000 --api-port 8080
```

Flags:
- `--ui-port` optional UI port
- `--api-port` optional API port
- `--no-browser` disable automatic browser launch
- `--version` print version

## Notes

- The SQLite database file is created beside the built executable as `app-template.db`.
- Static HTML routes switch to an expiration page after **April 30, 2026 at 23:59 GMT**.
- LLM settings are stored in SQLite and fall back to reasonable defaults when no records exist.
