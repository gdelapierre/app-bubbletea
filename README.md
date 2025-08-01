Absolutely! Here’s a **README.md** for your MVP Infrastructure Catalog launcher.

---

# Infrastructure Catalog Launcher

A terminal UI (TUI) launcher for deploying, updating, and managing Proxmox (or any Terraform-based) server deployments, using a catalog of team-managed YAML presets and a base Terraform project template.

---

## Features

* **Preset Catalog:** Pick from YAML “presets” for standardized deployment types (e.g., `elk`, `k8s-master`, `default`)—extend just by adding files!
* **New Deployments:** Create a new infrastructure deployment from a preset and customize fields in a guided form.
* **Update Existing Deployments:** Select and edit resource parameters for any existing deployment.
* **Self-contained:** Only depends on Go, Terraform, and your repo—no external services or complex setup.

---

## Prerequisites

* [Go 1.21+](https://go.dev/dl/) installed (`go version` to verify)
* [Terraform](https://developer.hashicorp.com/terraform/tutorials/aws-get-started/install-cli) installed (`terraform version` to verify)
* [git](https://git-scm.com/downloads) for managing your infrastructure repo

---

## Getting Started

### 1. Clone the Launcher Repository

```bash
git clone <your-launcher-repo-url>
cd <launcher-root>
```

### 2. Prepare Your Directory Structure

```text
launcher-root/
├── main.go
├── config.yaml
├── presets/              # For preset YAMLs (see below)
│   ├── default.yaml
│   ├── elk.yaml
│   └── ...etc
└── terraform/
    ├── apps/             # Each deployment gets its own subfolder here
    └── template/         # Your base Terraform project (main.tf, etc)
```

* `presets/`: Place one or more YAML files here (see **Presets** below)
* `terraform/template/`: Must contain all required `.tf` files (e.g., `main.tf`, `variables.tf`, `providers.tf`)

---

### 3. Configure `config.yaml`

Example:

```yaml
repo: git@github.com:yourorg/infra.git
apps_path: ./terraform/apps
template_path: ./terraform/template
presets_path: ./presets
```

Adjust the paths if your structure differs.

---

### 4. Add Presets

**`presets/default.yaml` (required):**

```yaml
vm_app: "myapp"
platform_id: "01"
vm_memory: "8192"
vm_cpu_cores: "2"
vm_disk_count: 1
vm_disk_size: ["100G"]
vm_count: 1
zone: "standard"
cluster: "cl10400"
# Add any other custom fields as needed
```

Add more (e.g., `elk.yaml`, `k8s-master.yaml`) to empower quick and consistent launches!

---

### 5. Build the Launcher

```bash
go build -o infra-launcher main.go
```

---

### 6. Run the Launcher

```bash
./infra-launcher
```

You’ll see a clean, navigable TUI:

```
╔══════════════════════════════════════════════════════════════════════╗
║                    Infrastructure Catalog Launcher                 ║
╠══════════════════════════════════════════════════════════════════════╣
Welcome! [N] New │ [U] Update │ [Q] Quit
╚══════════════════════════════════════════════════════════════════════╝
```

---

## Usage Overview

### New Deployment

* Press `N` to create a new deployment
* Use `←/→` to choose a preset (fields auto-fill)
* Tab through the form and customize any values
* Press `[Enter]` to save (future: will auto-create a deployment folder)

### Update Existing Deployment

* Press `U` to open a table of existing deployments
* Use `↑/↓` to select, `[Enter]` to edit
* Tab/edit any field, `[Enter]` to save

### Quit

* Press `Q` or `Esc` to quit at any time

---

## How It Works

* **Presets**: YAML files in `presets/` define default values for your form. Switching presets reloads the form instantly.
* **Template**: The Terraform project in `terraform/template/` is copied for every new deployment, so you get consistent infrastructure as code.
* **Deployments**: Each new deployment is created in `terraform/apps/` as a separate folder.
* **Edit**: You can edit resource counts, memory, CPU, and disks for existing deployments.

---

## Extending The Catalog

* **Add more YAML files to `presets/`**—the TUI will detect them automatically.
* **Add/edit fields in the form** by updating `main.go` and your YAMLs.
* **Team standards**: Update `default.yaml` to reflect best practices or requirements.

---

## Troubleshooting

* If you get an error on launch, check that all directories exist and are referenced correctly in `config.yaml`.
* The launcher assumes `terraform/template/` contains a valid Terraform project (`terraform init` should work in it).
* For Go errors, ensure dependencies (`github.com/charmbracelet/bubbletea`, `bubbles`, and `gopkg.in/yaml.v3`) are available:

  ```bash
  go get github.com/charmbracelet/bubbletea
  go get github.com/charmbracelet/bubbles/textinput
  go get github.com/charmbracelet/bubbles/table
  go get gopkg.in/yaml.v3
  ```

---

## Roadmap

* [ ] One-click `terraform apply`/`destroy` from the UI
* [ ] Git integration (status, commit, push, pull)
* [ ] In-app logs and error reporting
* [ ] Advanced preset features (inheritance, tags, search)

---

## License

*Your license and contact information here*

---

