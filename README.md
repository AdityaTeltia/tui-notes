# Terminal Notes - SSH-Based Note-Taking Application

A fully interactive, terminal-based note-taking application accessible via SSH. Users connect with `ssh notes.example.com` and are greeted with a beautiful TUI for writing, browsing, searching, and organizing notes—all directly in their terminal.

## Features

### Core Features
- 🎨 **Interactive TUI** - Beautiful text-based interface built with Bubble Tea
- 📁 **Folder/Notebook Structure** - Organize notes in folders
- 📝 **Live Markdown Preview** - See your notes rendered in real-time
- ⌨️ **Fast Keyboard Shortcuts** - Vim-like navigation and editing
- 🔐 **Secure Storage** - Optional encryption for sensitive notes
- 🔑 **Multiple Auth Methods** - Username/password or SSH key authentication
- 🔍 **Full-Text Search** - Search across titles, content, and tags
- 🏷️ **Tagging System** - Organize notes with tags
- 📤 **Export/Import** - Export to Markdown, JSON, TAR, or ZIP
- 💻 **CLI Commands** - Power user commands for export/import
- 🐧 **Cross-Platform** - Works on Linux, macOS, and WSL

### New Features (v1.0.0)
- 🔄 **Sorting & Filtering** - Sort by date, title, or filter by tags
- ⚡ **Quick Actions** - Jump to notes, recent notes, duplicate, pin notes
- 🔗 **Note Linking** - Internal links with `[[Note Title]]` syntax and backlinks
- ✅ **To-Do Lists** - Markdown checkboxes with completion tracking
- 📋 **Templates** - Pre-built templates for meetings, journals, code, todos
- 📜 **Version History** - Automatic snapshots with restore capability
- 🎨 **Themes** - 5 built-in themes (Default, Dark, Light, Monokai, Nord)

### Production Ready
- 📝 **Structured Logging** - Configurable log levels with file rotation
- ⚙️ **Configuration Files** - JSON-based configuration with CLI overrides
- 🔒 **Input Validation** - Comprehensive validation and sanitization
- 💾 **Auto-Backup** - Automatic backups before saves
- 🛡️ **Error Recovery** - Panic recovery and graceful shutdown
- 🏗️ **Build System** - Makefile with cross-platform builds

## Quick Start

### Prerequisites

- Go 1.21 or later
- SSH client
- Terminal with 256-color support

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/ssh-notes.git
cd ssh-notes

# Install dependencies
go mod download

# Build the server
go build -o ssh-notes-server

# Run the server
./ssh-notes-server -port 2222 -data ./data
```

### Connect

**Easy way (auto-configures SSH):**
```bash
# Install the wrapper script
cp ssh-notes /usr/local/bin/
# or add to PATH

# Then just run:
ssh-notes
```

The first time, it will prompt you to automatically add the SSH config, then connect.

**Manual way:**
```bash
ssh -p 2222 user@localhost
# or after setting up SSH config:
ssh notes.write
```

See [INSTALL.md](INSTALL.md) for detailed setup instructions.

## Usage

### Main Menu

- **Browse Notes** - Navigate your notes and folders
- **New Note** - Create a new note
- **Search Notes** - Full-text search across all notes
- **Tags** - Manage tags
- **Export/Import** - Export or import notes
- **Settings** - Configure the application
- **Quit** - Exit the application

### Keyboard Shortcuts

#### Navigation
- `↑/↓` or `j/k` - Navigate menu/list
- `Enter` - Select/Open
- `q` or `Esc` - Go back/Quit

#### Browser
- `n` - New note
- `e` - Edit selected note
- `d` - Delete selected note
- `p` - Preview selected note

#### Editor
- `Ctrl+S` - Save note
- `Ctrl+T` - Edit title
- `Ctrl+P` - Preview note
- `Ctrl+Space` - Toggle todo checkbox
- `i` - Enter insert mode
- `v` - Enter vim mode
- `Esc` - Exit editor

#### Quick Actions
- `g` - Quick jump to note
- `r` - Show recent notes
- `Ctrl+N` - New note from template
- `Ctrl+D` - Duplicate note
- `Ctrl+L` - Copy note link
- `p` - Pin/unpin note
- `s` - Cycle sort mode
- `f` - Filter by tag
- `Ctrl+F` - Clear filter
- `Ctrl+H` - Version history
- `Ctrl+T` - Cycle theme

#### Vim Mode
- `i` - Insert mode
- `Esc` - Normal mode
- `h/j/k/l` - Move cursor
- `w` - Save
- `q` - Quit

### CLI Commands

#### Export Notes

```bash
# Export to Markdown
./ssh-notes-server export -user alice -format markdown -output ./backup

# Export to JSON
./ssh-notes-server export -user alice -format json -output ./notes.json

# Export to TAR archive
./ssh-notes-server export -user alice -format tar -output ./notes.tar.gz

# Export to ZIP archive
./ssh-notes-server export -user alice -format zip -output ./notes.zip
```

#### Import Notes

```bash
# Import from Markdown directory
./ssh-notes-server import -user alice -format markdown -input ./backup

# Import from JSON file
./ssh-notes-server import -user alice -format json -input ./notes.json
```

## Configuration

### Server Options

```bash
./ssh-notes-server \
  -port 2222 \              # SSH server port
  -data ./data \            # Data directory
  -hostkey ./host_key \     # SSH host key path
  -auth password            # Auth mode: password, key, or both
```

### Authentication

#### Password Authentication

By default, the server accepts any password (for demo purposes). In production, implement proper password hashing:

1. Create a `users.yaml` file:
```yaml
users:
  - username: alice
    password_hash: $2a$10$...
  - username: bob
    password_hash: $2a$10$...
```

2. Update `auth.go` to load and verify passwords using bcrypt.

#### SSH Key Authentication

1. Add public keys to `~/.ssh/authorized_keys` format
2. Update `publicKeyAuth` in `main.go` to verify keys

## Data Storage

Notes are stored as JSON files in the user's data directory:
```
data/
  alice/
    note_1234567890.json
    note_1234567891.json
    folder/
      note_1234567892.json
```

Each note file contains:
```json
{
  "title": "My Note",
  "content": "# Note Content\n\nMarkdown content here...",
  "tags": ["work", "important"],
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "path": "/path/to/note.json",
  "encrypted": false
}
```

## Encryption

To enable encryption for a note:

1. Set an encryption key in the application
2. Notes marked as encrypted will be encrypted using AES-256-GCM
3. Encryption key should be derived from user password

## Deployment

See [DEPLOYMENT.md](DEPLOYMENT.md) for detailed deployment instructions.

## Development

### Project Structure

```
ssh-notes/
├── main.go              # SSH server and entry point
├── app.go               # Application wrapper
├── auth.go              # Authentication logic
├── cli.go               # CLI commands
├── models/
│   ├── main.go          # Main TUI model
│   ├── note.go          # Note management
│   ├── export.go        # Export/import functionality
│   └── encryption.go    # Encryption utilities
├── go.mod
├── go.sum
├── README.md
└── DEPLOYMENT.md
```

### Building

```bash
# Build for current platform
go build -o ssh-notes-server

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o ssh-notes-server-linux

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o ssh-notes-server-macos
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Future Enhancements

- [ ] Collaboration features (shared notes, real-time editing)
- [ ] To-do lists and task management
- [ ] REST API for programmatic access
- [ ] WebDAV sync support
- [ ] Plugin system
- [ ] Rich text formatting beyond Markdown
- [ ] Image support
- [ ] Version history and undo
- [ ] Multi-user sharing and permissions
- [ ] Mobile app companion

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

For issues and questions, please open an issue on GitHub.

