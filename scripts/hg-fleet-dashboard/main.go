package main

import (
	"fmt"
	"os"
	"time"

	"charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type model struct {
	repos []repoStatus
	err   error
}

type repoStatus struct {
	name     string
	ci       string
	tests    int
	latency  string
	errors   int
}

func (m model) Init() (tea.Model, tea.Cmd) {
	return m, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	s := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("hairglasses-studio Fleet Dashboard") + "\n\n"
	s += fmt.Sprintf("%-25s %-10s %-10s %-10s %-10s\n", "Repository", "CI", "Tests", "Latency", "Errors")
	s += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(fmt.Sprintf("%-25s %-10s %-10s %-10s %-10s", "──────────", "──", "─────", "───────", "──────")) + "\n"

	for _, r := range m.repos {
		ciStyle := lipgloss.NewStyle()
		if r.ci == "pass" {
			ciStyle = ciStyle.Foreground(lipgloss.Color("42"))
		} else if r.ci == "fail" {
			ciStyle = ciStyle.Foreground(lipgloss.Color("196"))
		}
		s += fmt.Sprintf("%-25s %-10s %-10d %-10s %-10d\n", r.name, ciStyle.Render(r.ci), r.tests, r.latency, r.errors)
	}

	s += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("q: quit")
	return s
}

func main() {
	m := model{
		repos: []repoStatus{
			{name: "mcpkit", ci: "pass", tests: 450, latency: "12ms", errors: 0},
			{name: "ralphglasses", ci: "pass", tests: 5972, latency: "45ms", errors: 2},
			{name: "art-mcp", ci: "fail", tests: 12, latency: "800ms", errors: 5},
			{name: "systemd-mcp", ci: "pass", tests: 34, latency: "5ms", errors: 0},
		},
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
