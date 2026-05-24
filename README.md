# Mind Forge

A collection of cognitive training games built with Go, templ, and Tailwind CSS.

## Development

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [VS Code](https://code.visualstudio.com/) with the [Dev Containers](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) extension

### Setup

**1. Build the dev image**

```sh
docker compose -f docker-compose.dev.yml build
```

**2. Start the container in the background**

```sh
docker compose -f docker-compose.dev.yml up -d
```

**2. Attach VS Code to the running container**

Open the Command Palette (`Ctrl+Shift+P`) → **Dev Containers: Attach to Running Container** → select `mind-forge-app-1`.

VS Code will reload inside the container. Open the `/app` folder as the workspace.

**3. Start the dev watchers**

In the integrated terminal inside the container:

```sh
make dev
```

This runs three watchers in parallel:

| Watcher | Command | Description |
|---|---|---|
| `templ` | `templ generate --watch` | Regenerates `*_templ.go` on `.templ` file changes |
| `tailwind` | `tailwindcss --watch` | Rebuilds `static/css/output.css` on CSS/template changes |
| `air` | `air` | Rebuilds and restarts the Go server on `.go` file changes |

The server is available at [http://localhost:8080](http://localhost:8080).

### Other make targets

```sh
make build        # Production build (templ + css + go binary)
make templ/gen    # One-shot templ generation
make css/build    # One-shot minified CSS build
make test         # Run tests
make test/cover   # Run tests with HTML coverage report
make lint         # golangci-lint + go vet
make clean        # Remove binary, output CSS, and generated templ files
```

### Project structure

```
.
├── main.go                      # Entry point, router setup
├── internal/
│   ├── handlers/                # HTTP handlers
│   └── templates/               # templ components
│       ├── layout/              # Base HTML layout
│       ├── components/          # Shared components (cards, etc.)
│       ├── home/                # Home page
│       └── game/                # Game pages (entry, how-to, machine, match, leaderboards)
├── static/
│   └── css/output.css           # Generated — do not edit
├── styles/
│   └── input.css                # Tailwind CSS v4 source
├── Dockerfile.dev
└── docker-compose.dev.yml
```
