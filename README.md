# ThighPads

ThighPads is a simple yet powerful terminal-based note-taking application built with Go. Organize your notes into tables and entries with a clean, responsive Terminal UI.

![ThighPads Screenshot]()

## Warning: Under Development

Please note that ThighPads is currently under active development. Expect bugs, breaking changes, and incomplete features.

## Features

- **Clean Terminal Interface** - Navigate your notes with an intuitive terminal UI
- **Hierarchical Organization** - Group your notes into tables and entries
- **Tag Support** - Add tags to entries for easy filtering and organization
- **Import/Export** - Easily share your tables with the `.thighpad` file format
- **Multiple Export Options** - Export to your config folder, desktop, or both
- **Automatic Updates** - Keep your application up to date with the latest features
- **Cross-Platform** - Works on Windows, macOS, and Linux

## Installation

### Quick Install

Download the latest binary from the [releases page](https://github.com/s42yt/thighpads/releases).

The application will automatically offer to install as a global command on first run, or you can force installation with:

```bash
thighpads --install
```

### Manual Installation

If you have Go installed, you can build from source:

```bash
# Clone the repository
git clone https://github.com/s42yt/thighpads.git

# Navigate to the directory
cd thighpads

# Build the application
go build -o thighpads ./cmd/thighpads

# Run the application
./thighpads
```

## Usage

### Getting Started

1. On first run, you'll be prompted to enter a username
2. From the home screen, press `n` to create your first table
3. Inside a table, press `n` again to create entries
4. Navigate with arrow keys and Enter

### Keyboard Shortcuts

#### Global
- `Ctrl+C` - Quit
- `Esc` - Go back or cancel

#### Home Screen
- `Enter` - Select table
- `n` - New table
- `i` - Import table
- `q` - Quit

#### Table Screen
- `Enter` - View entry
- `n` - New entry
- `d` - Delete entry
- `e` - Export table
- `b` - Back to home
- `q` - Quit

#### Entry Screens
- `Tab` - Switch between fields
- `Ctrl+S` - Save entry/changes
- `Esc` - Cancel

#### Export Screen
- `1-3` - Select export location (Default/Desktop/Both)
- `Enter` - Confirm export

## Configuration

ThighPads stores all configuration and data in:

- Windows: `%USERPROFILE%\.config\thighpads\`
- macOS/Linux: `~/.config/thighpads/`

### Command Line Options

```
thighpads [options]

Options:
  --version        Show version information
  --check-update   Check for updates
  --update         Update ThighPads to the latest version
  --wipe           Wipe all ThighPads data and start fresh
  --install        Force global installation
  --skip-install   Skip global installation
```

## Export Formats

ThighPads exports tables in a `.thighpad` file format, which contains:
- The table metadata
- All entries in the table
- Export timestamp and author information

## License

ThighPads is released under the MIT License. See [`LICENSE`](LICENSE) for details.

## Contributing

Contributions are welcome! Feel free to submit issues or pull requests on GitHub.