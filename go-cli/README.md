# Runtipi CLI (Go Version)

This is the Go implementation of the Runtipi CLI tool. It provides a command-line interface for managing your Runtipi home server.

## Building from Source

```bash
go build -o runtipi ./cmd/runtipi
```

## Available Commands

- `start`: Start Runtipi
- `stop`: Stop Runtipi
- `restart`: Restart Runtipi
- `update`: Update Runtipi
- `reset-password`: Reset Runtipi password
- `app`: Manage Runtipi apps
- `debug`: Debug Runtipi
- `version`: Show Runtipi version

## Development

This project uses:

- [Cobra](https://github.com/spf13/cobra) for CLI command structure
- Go modules for dependency management

To contribute:

1. Fork the repository
2. Create your feature branch
3. Make your changes
4. Submit a pull request

## License

Same as the original Runtipi project
