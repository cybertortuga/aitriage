package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type LogLevel int

const (
	LogDebug LogLevel = iota
	LogInfo
	LogWarn
	LogError
	LogSuccess
)

type LogEntry struct {
	Time    time.Time
	Level   LogLevel
	Source  string // "SCAN", "LLM", "EXPORT", "TRIAGE", "SAST", "NAV", "SYSTEM", "FIX"
	Message string
}

const maxLogEntries = 500

// AddLog adds a new log entry to the circular buffer.
// Note: DashboardModel is a value type in tea, so we must mutate the pointer
// if we call this outside of Update, or return the modified model if inside.
func (m *DashboardModel) AddLog(level LogLevel, source, message string) {
	entry := LogEntry{
		Time:    time.Now(),
		Level:   level,
		Source:  source,
		Message: message,
	}

	m.LogEntries = append(m.LogEntries, entry)
	if len(m.LogEntries) > maxLogEntries {
		m.LogEntries = m.LogEntries[1:]
	}

	// Only trigger format if we are already initialized
	if m.LogViewport.Width > 0 {
		m.formatLogViewport()
	}
}

// formatLogViewport formats the log entries and updates the viewport content
func (m *DashboardModel) formatLogViewport() {
	var content string
	for _, entry := range m.LogEntries {
		if m.LogFilter != -1 && entry.Level != m.LogFilter {
			continue
		}

		timeStr := entry.Time.Format("15:04:05")

		var levelStr string
		switch entry.Level {
		case LogDebug:
			levelStr = "DEBUG"
		case LogInfo:
			levelStr = "INFO "
		case LogWarn:
			levelStr = "WARN "
		case LogError:
			levelStr = "ERROR"
		case LogSuccess:
			levelStr = "OK   "
		}

		var formattedLevel string
		switch entry.Level {
		case LogDebug:
			formattedLevel = lipgloss.NewStyle().Foreground(colorOutline).Render(levelStr)
		case LogInfo:
			formattedLevel = lipgloss.NewStyle().Foreground(colorTextVariant).Render(levelStr)
		case LogWarn:
			formattedLevel = lipgloss.NewStyle().Foreground(colorTertiaryCont).Render(levelStr)
		case LogError:
			formattedLevel = lipgloss.NewStyle().Foreground(colorError).Render(levelStr)
		case LogSuccess:
			formattedLevel = lipgloss.NewStyle().Foreground(colorPrimaryDim).Render(levelStr)
		}

		sourceStr := fmt.Sprintf("[%s]", entry.Source)
		// Pad source to 8 chars
		if len(sourceStr) < 8 {
			sourceStr = sourceStr + fmt.Sprintf("%*s", 8-len(sourceStr), "")
		}

		formattedSource := lipgloss.NewStyle().Foreground(colorAccent).Render(sourceStr)

		messageStr := entry.Message

		// Combine
		line := fmt.Sprintf("%s %s %s %s\n",
			lipgloss.NewStyle().Foreground(colorOutline).Render(timeStr),
			formattedLevel,
			formattedSource,
			messageStr,
		)
		content += line
	}

	m.LogViewport.SetContent(content)

	if m.LogAutoScroll {
		m.LogViewport.GotoBottom()
	}
}
