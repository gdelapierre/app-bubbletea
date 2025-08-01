package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

const uiWidth = 160
const uiHeight = 40

var (
	// highlight style for focused field
	focusedStyle = lipgloss.NewStyle().Background(lipgloss.Color("#FFEB3B")).Foreground(lipgloss.Color("#111")).Bold(true)
	normalStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#EEE"))
	boxStyle     = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Width(uiWidth)
	tooltipStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Foreground(lipgloss.Color("240")).Width(uiWidth - 4)
)

type FieldMeta struct {
	Label string `yaml:"label"`
	Help  string `yaml:"help"`
}
type FieldsYaml struct {
	Fields map[string]FieldMeta `yaml:"fields"`
}

func loadFieldMeta(path string) (map[string]FieldMeta, error) {
	var fy FieldsYaml
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, &fy); err != nil {
		return nil, err
	}
	return fy.Fields, nil
}

type Config struct {
	Repo         string `yaml:"repo"`
	AppsPath     string `yaml:"apps_path"`
	TemplatePath string `yaml:"template_path"`
	PresetsPath  string `yaml:"presets_path"`
}

func loadConfig(path string) (Config, error) {
	var cfg Config
	f, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	err = yaml.Unmarshal(f, &cfg)
	return cfg, err
}

type Preset struct {
	Name   string
	Values map[string]interface{}
}

func loadPresets(presetsDir string) ([]Preset, error) {
	entries, err := os.ReadDir(presetsDir)
	if err != nil {
		return nil, err
	}
	var out []Preset
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".yaml") {
			path := filepath.Join(presetsDir, e.Name())
			values, err := loadPreset(path)
			if err != nil {
				continue
			}
			name := strings.TrimSuffix(e.Name(), ".yaml")
			out = append(out, Preset{Name: name, Values: values})
		}
	}
	return out, nil
}
func loadPreset(path string) (map[string]interface{}, error) {
	var out map[string]interface{}
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(f, &out)
	return out, err
}

func loadTfvars(filename string) (map[string]string, error) {
	m := make(map[string]string)
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		m[key] = val
	}
	return m, scanner.Err()
}
func saveTfvars(filename string, updates map[string]string) error {
	input, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	lines := strings.Split(string(input), "\n")
	for i, line := range lines {
		for key, newval := range updates {
			if strings.HasPrefix(strings.TrimSpace(line), key+" ") || strings.HasPrefix(strings.TrimSpace(line), key+"=") {
				lines[i] = fmt.Sprintf("%s = %s", key, newval)
			}
		}
	}
	output := strings.Join(lines, "\n")
	return os.WriteFile(filename, []byte(output), 0644)
}

type deploymentInfo struct {
	Name         string
	Description  string
	LastModified string
	Path         string
}

func listDeployments(appsDir string) ([]deploymentInfo, error) {
	entries, err := os.ReadDir(appsDir)
	if err != nil {
		return nil, err
	}
	var infos []deploymentInfo
	for _, e := range entries {
		if e.IsDir() {
			full := filepath.Join(appsDir, e.Name())
			stat, err := os.Stat(full)
			if err != nil {
				continue
			}
			desc := ""
			tfvarsPath := filepath.Join(full, "terraform.tfvars")
			if vals, err := loadTfvars(tfvarsPath); err == nil {
				desc = strings.Trim(vals["platform_description"], "\"")
			}
			infos = append(infos, deploymentInfo{
				Name:         e.Name(),
				Description:  desc,
				LastModified: stat.ModTime().Format("2006-01-02 15:04"),
				Path:         full,
			})
		}
	}
	return infos, nil
}

func copyDir(src string, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			in, err := os.Open(srcPath)
			if err != nil {
				return err
			}
			defer in.Close()
			out, err := os.Create(dstPath)
			if err != nil {
				return err
			}
			defer out.Close()
			if _, err = io.Copy(out, in); err != nil {
				return err
			}
			info, _ := os.Stat(srcPath)
			if err = out.Chmod(info.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

// Navigation options for select/dropdown fields
var (
	clusterOptions = []string{"cl10400", "cl12600k", "cl12900h", "cl13600k"}
	zoneOptions    = []string{"standard", "admin", "dmz"}
)

func cycleOption(current string, options []string, dir int) string {
	for i, opt := range options {
		if opt == current {
			newIdx := (i + dir + len(options)) % len(options)
			return options[newIdx]
		}
	}
	return options[0]
}

type scene int

const (
	sceneLauncher scene = iota
	sceneCreateForm
	sceneEditTable
	sceneEditForm
)

type model struct {
	cfg           Config
	presets       []Preset
	presetIdx     int
	fieldMeta     map[string]FieldMeta
	helpText      string
	currentScene  scene
	statusMessage string

	createInputs []textinput.Model
	createLabels []string
	createFocus  int

	editTableData []deploymentInfo
	editStatus    string

	editFormInputs []textinput.Model
	editFormLabels []string
	editFormPath   string
	editFocusIndex int

	// Add future git/vault status here (strings for now, wiring later)
	gitStatus   string
	vaultStatus string
}

func (m model) Init() tea.Cmd {
	return nil
}

func main() {
	cfg, err := loadConfig("config.yaml")
	if err != nil {
		fmt.Println("ERROR: could not load config.yaml:", err)
		os.Exit(1)
	}
	presets, err := loadPresets(cfg.PresetsPath)
	if err != nil {
		fmt.Println("ERROR: could not load presets from presets dir:", err)
		os.Exit(1)
	}
	if len(presets) == 0 {
		fmt.Println("No presets found in presets dir!")
		os.Exit(1)
	}
	fieldMeta, err := loadFieldMeta("fields.yaml")
	if err != nil {
		fmt.Println("ERROR: could not load fields.yaml:", err)
		os.Exit(1)
	}
	m := initialModel(cfg, presets, fieldMeta)
	if _, err := tea.NewProgram(m).Run(); err != nil {
		log.Fatal(err)
	}
}

func initialModel(cfg Config, presets []Preset, fieldMeta map[string]FieldMeta) model {
	labels := []string{
		"vm_app", "platform_description", "zone", "platform_id", "vm_network_suffix", "vm_id_prefix",
		"vm_memory", "vm_cpu_cores", "vm_disk_count", "vm_disk_size", "vm_count", "cluster",
	}
	inputs := make([]textinput.Model, len(labels))
	presetIdx := 0
	for i, name := range labels {
		ti := textinput.New()
		ti.Placeholder = name
		if val, ok := presets[presetIdx].Values[name]; ok {
			switch v := val.(type) {
			case string:
				ti.SetValue(v)
			case int:
				ti.SetValue(fmt.Sprintf("%d", v))
			case []interface{}:
				strs := []string{}
				for _, e := range v {
					strs = append(strs, fmt.Sprintf("%v", e))
				}
				ti.SetValue(strings.Join(strs, ","))
			default:
				ti.SetValue(fmt.Sprintf("%v", v))
			}
		}
		inputs[i] = ti
	}
	inputs[0].Focus()
	return model{
		cfg:            cfg,
		presets:        presets,
		presetIdx:      presetIdx,
		currentScene:   sceneLauncher,
		createInputs:   inputs,
		createLabels:   labels,
		createFocus:    0,
		fieldMeta:      fieldMeta,
		helpText:       "",
		editFormLabels: []string{"vm_cpu_cores", "vm_memory", "vm_count", "vm_disk_count", "vm_disk_size"},
		gitStatus:      "Git: ⏳",   // placeholder for future wiring
		vaultStatus:    "Vault: ⏳", // placeholder for future wiring
	}
}

func (m model) View() string {
	var header, body, tooltip, footer string

	// Build status for top right corner
	status := padLeft(fmt.Sprintf("%s   %s", m.gitStatus, m.vaultStatus), uiWidth-2-len("Infrastructure Catalog"))

	switch m.currentScene {
	case sceneLauncher:
		header = boxTop("Infrastructure Catalog", uiWidth, status)
		// No tooltip or preset on launcher
		body = ""
		footer = centerText("[N] New │ [U] Update │ [Q] Quit", uiWidth)

	case sceneCreateForm:
		header = boxTop("Create Deployment", uiWidth, status)
		body += fmt.Sprintf("[Preset: %s] (F2/F3 to switch)\n\n", m.presets[m.presetIdx].Name)

		for i, ti := range m.createInputs {
			cursor := " "
			val := ti.Value()
			isFocused := i == m.createFocus
			label := m.fieldMeta[m.createLabels[i]].Label
			var field string
			if isFocused {
				// Highlight select fields
				if m.createLabels[i] == "zone" || m.createLabels[i] == "cluster" {
					field = focusedStyle.Render(fmt.Sprintf("%s %-25s: %s", cursor, label, val))
				} else {
					field = focusedStyle.Render(fmt.Sprintf("%s %-25s: %s", cursor, label, ti.View()))
				}
			} else {
				field = normalStyle.Render(fmt.Sprintf("%s %-25s: %s", cursor, label, ti.View()))
			}
			body += field + "\n"
		}
		// Tooltip immediately below fields
		tooltip = tooltipBox(m.fieldMeta[m.createLabels[m.createFocus]].Help)
		footer = navFooter()

	case sceneEditTable:
		header = boxTop("Select Deployment to Edit", uiWidth, status)
		body = tableHeader([]string{"Deployment", "Description", "Last Modified"}, []int{35, 60, 20}, uiWidth)
		for _, row := range m.editTableData {
			body += fmt.Sprintf(" %-33s │ %-58s │ %-18s\n", row.Name, row.Description, row.LastModified)
		}
		tooltip = tooltipBox("Select a deployment to edit")
		footer = navFooter()

	case sceneEditForm:
		header = boxTop("Edit Deployment", uiWidth, status)
		body += fmt.Sprintf("[Preset: %s]\n\n", m.presets[m.presetIdx].Name)
		for i, ti := range m.editFormInputs {
			cursor := " "
			val := ti.Value()
			isFocused := i == m.editFocusIndex
			label := m.fieldMeta[m.editFormLabels[i]].Label
			var field string
			if isFocused {
				if m.editFormLabels[i] == "zone" || m.editFormLabels[i] == "cluster" {
					field = focusedStyle.Render(fmt.Sprintf("%s %-25s: %s", cursor, label, val))
				} else {
					field = focusedStyle.Render(fmt.Sprintf("%s %-25s: %s", cursor, label, ti.View()))
				}
			} else {
				field = normalStyle.Render(fmt.Sprintf("%s %-25s: %s", cursor, label, ti.View()))
			}
			body += field + "\n"
		}
		if m.editStatus != "" {
			body += "\n" + m.editStatus + "\n"
		}
		tooltip = tooltipBox(m.fieldMeta[m.editFormLabels[m.editFocusIndex]].Help)
		footer = navFooter()

	default:
		header, body, tooltip, footer = "", "", "", ""
	}

	// Compose: header + body + tooltip + (vertical pad) + footer + boxBottom
	s := header + body + tooltip
	linesSoFar := countLines(s)
	footerLines := countLines(footer)
	boxBottomLines := 1
	paddingLines := uiHeight - linesSoFar - footerLines - boxBottomLines
	if paddingLines < 0 {
		paddingLines = 0
	}
	s += strings.Repeat("\n", paddingLines)
	s += footer + "\n"
	s += boxBottom(uiWidth)
	return s
}

func boxTop(title string, width int, status string) string {
	// status right-justified
	left := centerText(title, width-len(status))
	return fmt.Sprintf("╔%s╗\n║%s%s║\n╠%s╣\n",
		strings.Repeat("═", width),
		left, status,
		strings.Repeat("═", width))
}
func boxBottom(width int) string {
	return fmt.Sprintf("╚%s╝\n", strings.Repeat("═", width))
}
func centerText(s string, width int) string {
	if len(s) >= width {
		return s
	}
	padding := (width - len(s)) / 2
	return strings.Repeat(" ", padding) + s + strings.Repeat(" ", width-len(s)-padding)
}
func padLeft(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(" ", width-len(s)) + s
}
func countLines(s string) int {
	return strings.Count(s, "\n")
}
func tableHeader(cols []string, widths []int, totalWidth int) string {
	row := ""
	for i, c := range cols {
		row += " " + c + strings.Repeat(" ", widths[i]-len(c)) + "│"
	}
	spacing := totalWidth - len(row)
	if spacing > 0 {
		row += strings.Repeat(" ", spacing)
	}
	return row + "\n" + strings.Repeat("-", totalWidth) + "\n"
}
func tooltipBox(msg string) string {
	return tooltipStyle.Render("\n Tooltip\n" + strings.Repeat("─", uiWidth-4) + "\n" + msg + "\n")
}
func navFooter() string {
	return centerText("[↑/↓] Field │ [←/→/Space] Cycle Dropdown │ [Tab] Next │ [Enter] Save │ [Esc] Cancel", uiWidth)
}

// update logic
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.currentScene {
	case sceneLauncher:
		return updateLauncher(m, msg)
	case sceneCreateForm:
		return updateCreateForm(m, msg)
	case sceneEditTable:
		return updateEditTable(m, msg)
	case sceneEditForm:
		return updateEditForm(m, msg)
	}
	return m, nil
}

func (m model) withScene(s scene) model {
	m.currentScene = s
	return m
}

func updateLauncher(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "n":
			return m.withScene(sceneCreateForm), nil
		case "u":
			infos, err := listDeployments(m.cfg.AppsPath)
			if err != nil || len(infos) == 0 {
				m.statusMessage = "No deployments found"
				return m, nil
			}
			m.editTableData = infos
			m.currentScene = sceneEditTable
			return m, nil
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func updateCreateForm(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		curLabel := m.createLabels[m.createFocus]
		switch msg.String() {
		case "ctrl+c", "esc":
			return m.withScene(sceneLauncher), nil
		case "tab":
			m.createFocus = (m.createFocus + 1) % len(m.createInputs)
		case "shift+tab":
			m.createFocus = (m.createFocus - 1 + len(m.createInputs)) % len(m.createInputs)
		case "up":
			m.createFocus = (m.createFocus - 1 + len(m.createInputs)) % len(m.createInputs)
		case "down":
			m.createFocus = (m.createFocus + 1) % len(m.createInputs)
		case "left":
			if curLabel == "zone" {
				cur := m.createInputs[m.createFocus].Value()
				m.createInputs[m.createFocus].SetValue(cycleOption(cur, zoneOptions, -1))
			} else if curLabel == "cluster" {
				cur := m.createInputs[m.createFocus].Value()
				m.createInputs[m.createFocus].SetValue(cycleOption(cur, clusterOptions, -1))
			}
		case "right":
			if curLabel == "zone" {
				cur := m.createInputs[m.createFocus].Value()
				m.createInputs[m.createFocus].SetValue(cycleOption(cur, zoneOptions, +1))
			} else if curLabel == "cluster" {
				cur := m.createInputs[m.createFocus].Value()
				m.createInputs[m.createFocus].SetValue(cycleOption(cur, clusterOptions, +1))
			}
		case " ":
			if curLabel == "zone" {
				cur := m.createInputs[m.createFocus].Value()
				m.createInputs[m.createFocus].SetValue(cycleOption(cur, zoneOptions, +1))
			} else if curLabel == "cluster" {
				cur := m.createInputs[m.createFocus].Value()
				m.createInputs[m.createFocus].SetValue(cycleOption(cur, clusterOptions, +1))
			}
		case "f2":
			m.presetIdx = (m.presetIdx + len(m.presets) - 1) % len(m.presets)
			m.prefillCreateFormFromPreset()
		case "f3":
			m.presetIdx = (m.presetIdx + 1) % len(m.presets)
			m.prefillCreateFormFromPreset()
		case "enter":
			// Save logic here!
			return m.withScene(sceneLauncher), nil
		}
		// Focus logic
		for i := range m.createInputs {
			if i == m.createFocus {
				m.createInputs[i].Focus()
			} else {
				m.createInputs[i].Blur()
			}
		}
	}
	var cmds []tea.Cmd
	for i := range m.createInputs {
		ti, cmd := m.createInputs[i].Update(msg)
		m.createInputs[i] = ti
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m *model) prefillCreateFormFromPreset() {
	cur := m.presets[m.presetIdx]
	for i, label := range m.createLabels {
		if label == "zone" || label == "cluster" {
			// Do not touch zone or cluster—leave user selection as-is!
			continue
		}
		if val, ok := cur.Values[label]; ok {
			switch v := val.(type) {
			case string:
				m.createInputs[i].SetValue(v)
			case int:
				m.createInputs[i].SetValue(fmt.Sprintf("%d", v))
			case []interface{}:
				strs := []string{}
				for _, e := range v {
					strs = append(strs, fmt.Sprintf("%v", e))
				}
				m.createInputs[i].SetValue(strings.Join(strs, ","))
			default:
				m.createInputs[i].SetValue(fmt.Sprintf("%v", v))
			}
		} else {
			m.createInputs[i].SetValue("")
		}
	}
	m.createInputs[0].Focus()
	m.createFocus = 0
}

func updateEditTable(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return m.withScene(sceneLauncher), nil
		case "up":
			if len(m.editTableData) > 1 {
				last := len(m.editTableData) - 1
				copy(m.editTableData[1:], m.editTableData[:last])
				m.editTableData[0] = m.editTableData[last]
			}
		case "down":
			if len(m.editTableData) > 1 {
				first := m.editTableData[0]
				copy(m.editTableData, m.editTableData[1:])
				m.editTableData[len(m.editTableData)-1] = first
			}
		case "enter":
			selected := 0
			info := m.editTableData[selected]
			tfvars := filepath.Join(info.Path, "terraform.tfvars")
			vals, err := loadTfvars(tfvars)
			if err != nil {
				m.editStatus = "Could not load tfvars: " + err.Error()
				return m, nil
			}
			labels := m.editFormLabels
			inputs := make([]textinput.Model, len(labels))
			for i, key := range labels {
				ti := textinput.New()
				ti.Placeholder = key
				val := vals[key]
				val = strings.Trim(val, "\"")
				val = strings.Trim(val, "[]")
				ti.SetValue(val)
				inputs[i] = ti
			}
			inputs[0].Focus()
			m.editFormInputs = inputs
			m.editFormPath = tfvars
			m.editFocusIndex = 0
			m.editStatus = ""
			return m.withScene(sceneEditForm), nil
		}
	}
	return m, nil
}

func updateEditForm(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		curLabel := m.editFormLabels[m.editFocusIndex]
		switch msg.String() {
		case "esc", "q":
			return m.withScene(sceneLauncher), nil
		case "tab":
			m.editFocusIndex = (m.editFocusIndex + 1) % len(m.editFormInputs)
		case "shift+tab":
			m.editFocusIndex = (m.editFocusIndex - 1 + len(m.editFormInputs)) % len(m.editFormInputs)
		case "up":
			m.editFocusIndex = (m.editFocusIndex - 1 + len(m.editFormInputs)) % len(m.editFormInputs)
		case "down":
			m.editFocusIndex = (m.editFocusIndex + 1) % len(m.editFormInputs)
		case "left":
			if curLabel == "zone" {
				cur := m.editFormInputs[m.editFocusIndex].Value()
				m.editFormInputs[m.editFocusIndex].SetValue(cycleOption(cur, zoneOptions, -1))
			} else if curLabel == "cluster" {
				cur := m.editFormInputs[m.editFocusIndex].Value()
				m.editFormInputs[m.editFocusIndex].SetValue(cycleOption(cur, clusterOptions, -1))
			}
		case "right":
			if curLabel == "zone" {
				cur := m.editFormInputs[m.editFocusIndex].Value()
				m.editFormInputs[m.editFocusIndex].SetValue(cycleOption(cur, zoneOptions, +1))
			} else if curLabel == "cluster" {
				cur := m.editFormInputs[m.editFocusIndex].Value()
				m.editFormInputs[m.editFocusIndex].SetValue(cycleOption(cur, clusterOptions, +1))
			}
		case " ":
			if curLabel == "zone" {
				cur := m.editFormInputs[m.editFocusIndex].Value()
				m.editFormInputs[m.editFocusIndex].SetValue(cycleOption(cur, zoneOptions, +1))
			} else if curLabel == "cluster" {
				cur := m.editFormInputs[m.editFocusIndex].Value()
				m.editFormInputs[m.editFocusIndex].SetValue(cycleOption(cur, clusterOptions, +1))
			}
		case "enter":
			labels := m.editFormLabels
			updates := map[string]string{}
			for i, key := range labels {
				v := m.editFormInputs[i].Value()
				if key == "vm_disk_size" {
					arr := []string{}
					for _, part := range strings.Split(v, ",") {
						s := strings.TrimSpace(part)
						s = strings.Trim(s, "\"")
						arr = append(arr, fmt.Sprintf("\"%s\"", s))
					}
					updates[key] = "[" + strings.Join(arr, ", ") + "]"
				} else {
					updates[key] = v
				}
			}
			if err := saveTfvars(m.editFormPath, updates); err != nil {
				m.editStatus = "Save failed: " + err.Error()
			} else {
				m.editStatus = "Saved! (You may now apply changes as needed.)"
			}
			return m, nil
		}
		for i := range m.editFormInputs {
			if i == m.editFocusIndex {
				m.editFormInputs[i].Focus()
			} else {
				m.editFormInputs[i].Blur()
			}
		}
	}
	var cmds []tea.Cmd
	for i := range m.editFormInputs {
		ti, cmd := m.editFormInputs[i].Update(msg)
		m.editFormInputs[i] = ti
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}
