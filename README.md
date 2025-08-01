Absolutely! Here’s an updated `README.md` for your Proxmox Infra Launcher, reflecting your latest improvements:

---

````markdown
# Proxmox Infra Launcher

A terminal-based Infrastructure Catalog and deployment launcher for Proxmox, written in Go using Bubble Tea and Lipgloss.

## Features

- Modern, full-screen TUI with sticky footer, tooltips, and focus highlights
- Multi-preset YAML-driven VM configurations (just add presets in the `presets/` directory)
- Create and update deployments via forms with keyboard navigation (up/down, tab, F2/F3 for presets, left/right for select fields)
- Dedicated tooltip box for field help, always visible in the UI
- Real-time status indicators for Git and Vault (wiring pending)
- Safe config handling (sample config provided, real config ignored by git)
- Extensible: easily adapt fields via `fields.yaml` and add presets as you grow!

## Quick Start

### 1. **Clone the repo**
```sh
git clone git@github.com:your-org/infra-catalog.git
cd infra-catalog
````

### 2. **Install Go dependencies**

```sh
go mod tidy
```

### 3. **Configure your environment**

* **Copy the example config and edit it:**

  ```sh
  cp config_example.yaml config.yaml
  # then edit config.yaml with your real paths/settings
  ```
* **Presets**: Add your own presets (YAML files) in the `presets/` directory.

### 4. **Run the launcher**

```sh
go run main.go
```

## Configuration

All user/site-specific settings live in `config.yaml`.
The file is **gitignored** (not committed!)—edit `config_example.yaml` and copy to `config.yaml` before first run.

### Example `config_example.yaml`:

```yaml
repo: "git@github.com:your-org/infra-catalog.git"
apps_path: "/home/username/terraform/apps"
template_path: "/home/username/terraform/template"
presets_path: "/home/username/launcher/presets"
```

## Keyboard Shortcuts

| Key         | Action                                       |
| ----------- | -------------------------------------------- |
| **N**       | Create new deployment                        |
| **U**       | Update an existing deployment                |
| **Q / Esc** | Quit launcher                                |
| **↑/↓**     | Move between form fields                     |
| **←/→**     | Cycle select/dropdown fields (zone, cluster) |
| **Space**   | Cycle select/dropdown fields                 |
| **F2/F3**   | Switch presets in Create view                |
| **Tab**     | Move to next field                           |
| **Enter**   | Save form / proceed                          |

## Directory Structure

```
.
├── main.go
├── config_example.yaml   # <- commit this, not your real config.yaml
├── .gitignore
├── fields.yaml           # field metadata for form UI
├── presets/              # put your YAML presets here
├── apps/                 # auto-managed by launcher
├── template/             # your Terraform template for new deployments
└── ...
```

## FAQ

**Q: Where do I set my Vault and Git status?**
A: The UI will display real-time status indicators in the top-right (wiring coming soon—update `m.gitStatus` and `m.vaultStatus` in code).

**Q: How do I add more VM types?**
A: Just drop a new preset YAML in the `presets/` directory!

**Q: How do I add a new field?**
A: Add to `fields.yaml`, update your presets, and rebuild.