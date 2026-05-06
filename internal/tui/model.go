package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mager/local-llm-quickstart/internal/llm"
)

type LLMClient interface {
	Chat(ctx context.Context, messages []llm.Message, maxTokens int, temperature float64) (llm.Response, error)
}

type Config struct {
	Client      LLMClient
	Endpoint    string
	Model       string
	MaxTokens   int
	Temperature float64
}

type Model struct {
	client      LLMClient
	endpoint    string
	model       string
	maxTokens   int
	autoTokens  bool
	lastTokens  int
	temperature float64
	textarea    textarea.Model
	viewport    viewport.Model
	messages    []llm.Message
	width       int
	height      int
	waiting     bool
	err         error
}

type responseMsg struct {
	content      string
	finishReason string
	tokens       int
	err          error
}

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "238", Dark: "252"})
	inputHint   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "236", Dark: "255"})
	userStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	llmStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

func New(config Config) Model {
	input := textarea.New()
	input.Placeholder = "Ask a coding question. Enter sends, Shift+Enter adds a newline."
	input.Focus()
	input.ShowLineNumbers = false
	input.SetHeight(3)
	input.FocusedStyle.Placeholder = inputHint
	input.FocusedStyle.Prompt = inputHint
	input.BlurredStyle.Placeholder = inputHint
	input.BlurredStyle.Prompt = inputHint

	vp := viewport.New(80, 20)
	m := Model{
		client:      config.Client,
		endpoint:    config.Endpoint,
		model:       config.Model,
		maxTokens:   config.MaxTokens,
		autoTokens:  config.MaxTokens <= 0,
		temperature: config.Temperature,
		textarea:    input,
		viewport:    vp,
	}
	m.refreshViewport()
	return m
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = max(1, msg.Height-7)
		m.textarea.SetWidth(msg.Width)
		m.refreshViewport()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "pgup":
			m.viewport.PageUp()
			return m, nil
		case "pgdown":
			m.viewport.PageDown()
			return m, nil
		case "ctrl+u":
			m.viewport.HalfPageUp()
			return m, nil
		case "ctrl+d":
			m.viewport.HalfPageDown()
			return m, nil
		case "ctrl+g":
			m.viewport.GotoTop()
			return m, nil
		case "ctrl+b":
			m.viewport.GotoBottom()
			return m, nil
		}
		if m.waiting {
			return m, nil
		}
		switch msg.String() {
		case "enter":
			value := strings.TrimSpace(m.textarea.Value())
			if value == "" {
				return m, nil
			}
			m.textarea.Reset()
			if strings.HasPrefix(value, "/") {
				return m.handleCommand(value)
			}
			m.messages = append(m.messages, llm.Message{Role: "user", Content: value})
			m.waiting = true
			m.lastTokens = m.tokensForMessages(m.messages)
			m.err = nil
			m.refreshViewport()
			return m, m.requestCompletion()
		}

	case responseMsg:
		m.waiting = false
		m.lastTokens = msg.tokens
		if msg.err != nil {
			m.err = msg.err
			m.messages = append(m.messages, llm.Message{Role: "assistant", Content: "Request failed. Is llama-server running?"})
		} else {
			m.err = nil
			content := msg.content
			if msg.finishReason == "length" {
				content = strings.TrimSpace(content) + "\n\n[Stopped at token limit. Try `/tokens auto` or `/tokens 4096` and ask again.]"
			}
			m.messages = append(m.messages, llm.Message{Role: "assistant", Content: content})
		}
		m.refreshViewport()
		return m, nil

	case tea.MouseMsg:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	tokenLabel := fmt.Sprintf("tokens=%d", m.maxTokens)
	if m.autoTokens {
		tokenLabel = fmt.Sprintf("tokens=auto next~%d", m.estimatedTokensForDraft())
		if strings.TrimSpace(m.textarea.Value()) == "" && m.lastTokens > 0 {
			tokenLabel = fmt.Sprintf("tokens=auto last=%d", m.lastTokens)
		}
	}
	header := headerStyle.Render("local-llm") + helpStyle.Render(
		fmt.Sprintf("  %s  %s temp=%.2f", m.endpoint, tokenLabel, m.temperature),
	)
	footer := helpStyle.Render("pgup/pgdn scroll, ctrl+u/d half, ctrl+g/b top/bottom | /help /tokens auto|2048 /temp 0.4 /quit")
	if m.waiting {
		if m.lastTokens > 0 {
			footer = helpStyle.Render(fmt.Sprintf("thinking... tokens=%d ", m.lastTokens)) + footer
		} else {
			footer = helpStyle.Render("thinking... ") + footer
		}
	}
	if m.err != nil {
		footer = errorStyle.Render(m.err.Error())
	}
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		m.viewport.View(),
		m.textarea.View(),
		footer,
	)
}

func (m Model) requestCompletion() tea.Cmd {
	messages := append([]llm.Message(nil), m.messages...)
	maxTokens := m.tokensForMessages(messages)
	temperature := m.temperature
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		resp, err := m.client.Chat(ctx, messages, maxTokens, temperature)
		return responseMsg{
			content:      resp.Content,
			finishReason: resp.FinishReason,
			tokens:       maxTokens,
			err:          err,
		}
	}
}

func (m Model) handleCommand(value string) (tea.Model, tea.Cmd) {
	name, arg, _ := strings.Cut(value, " ")
	name = strings.ToLower(strings.TrimSpace(name))
	arg = strings.TrimSpace(arg)

	switch name {
	case "/q", "/quit", "/exit":
		return m, tea.Quit
	case "/help":
		m.messages = append(m.messages, llm.Message{
			Role:    "assistant",
			Content: "Commands: /help, /reset, /save transcript.md, /tokens auto, /tokens 2048, /temp 0.4, /model, /quit.\n\nScroll: PageUp/PageDown, Ctrl+U/Ctrl+D, Ctrl+G top, Ctrl+B bottom, mouse wheel.",
		})
	case "/model":
		m.messages = append(m.messages, llm.Message{
			Role:    "assistant",
			Content: fmt.Sprintf("endpoint=%s\nmodel=%s", m.endpoint, m.model),
		})
	case "/reset":
		m.messages = nil
	case "/save":
		if arg == "" {
			arg = "transcripts/local-llm-transcript.md"
		}
		if err := saveTranscript(arg, m.messages); err != nil {
			m.err = err
		} else {
			m.messages = append(m.messages, llm.Message{
				Role:    "assistant",
				Content: "Saved transcript to " + arg,
			})
		}
	case "/tokens":
		if arg == "" || strings.EqualFold(arg, "auto") {
			m.autoTokens = true
			m.err = nil
			break
		}
		value, err := strconv.Atoi(arg)
		if err != nil || value < 1 {
			m.err = fmt.Errorf("usage: /tokens auto or /tokens 2048")
		} else {
			m.maxTokens = value
			m.autoTokens = false
			m.err = nil
		}
	case "/temp":
		value, err := strconv.ParseFloat(arg, 64)
		if err != nil || value < 0 {
			m.err = fmt.Errorf("usage: /temp 0.4")
		} else {
			m.temperature = value
			m.err = nil
		}
	default:
		m.err = fmt.Errorf("unknown command: %s", name)
	}

	m.refreshViewport()
	return m, nil
}

func (m *Model) refreshViewport() {
	if len(m.messages) == 0 {
		m.viewport.SetContent(helpStyle.Render("Start typing below. The model stays local; this TUI only talks to your localhost server."))
		return
	}

	var b strings.Builder
	for _, msg := range m.messages {
		switch msg.Role {
		case "user":
			b.WriteString(userStyle.Render("you"))
		default:
			b.WriteString(llmStyle.Render("llm"))
		}
		b.WriteString("\n")
		b.WriteString(strings.TrimSpace(msg.Content))
		b.WriteString("\n\n")
	}
	if m.waiting {
		b.WriteString(llmStyle.Render("llm"))
		b.WriteString("\nThinking...\n")
	}
	m.viewport.SetContent(strings.TrimSpace(b.String()))
	m.viewport.GotoBottom()
}

func (m Model) tokensForMessages(messages []llm.Message) int {
	if !m.autoTokens {
		return m.maxTokens
	}
	if len(messages) == 0 {
		return 512
	}
	return estimateMaxTokens(messages[len(messages)-1].Content)
}

func (m Model) estimatedTokensForDraft() int {
	value := strings.TrimSpace(m.textarea.Value())
	if value == "" && len(m.messages) > 0 {
		value = m.messages[len(m.messages)-1].Content
	}
	if value == "" {
		return 512
	}
	return estimateMaxTokens(value)
}

func estimateMaxTokens(prompt string) int {
	lower := strings.ToLower(prompt)
	promptTokens := roughTokenCount(prompt)
	estimated := 512 + promptTokens*4

	if isTinyPrompt(lower, promptTokens) {
		estimated = 128
	}
	if strings.Contains(lower, "itinerary") || strings.Contains(lower, "day-by-day") || strings.Contains(lower, "day by day") {
		estimated = max(estimated, 2048)
		if days := mentionedDays(lower); days > 0 {
			estimated = max(estimated, days*160)
		}
	}
	if strings.Contains(lower, "detailed") || strings.Contains(lower, "comprehensive") || strings.Contains(lower, "full ") {
		estimated = max(estimated, 2048)
	}
	if strings.Contains(lower, "blog post") || strings.Contains(lower, "essay") || strings.Contains(lower, "draft") {
		estimated = max(estimated, 2048)
	}
	if strings.Contains(lower, "code") || strings.Contains(lower, "implement") || strings.Contains(lower, "debug") {
		estimated = max(estimated, 1536)
	}

	return clamp(estimated, 128, 4096)
}

func roughTokenCount(text string) int {
	runes := utf8.RuneCountInString(text)
	if runes == 0 {
		return 0
	}
	return max(1, (runes+3)/4)
}

func isTinyPrompt(lower string, promptTokens int) bool {
	if promptTokens > 16 {
		return false
	}
	switch strings.TrimSpace(lower) {
	case "hi", "hello", "hey", "hello world", "test":
		return true
	default:
		return false
	}
}

func mentionedDays(lower string) int {
	re := regexp.MustCompile(`\b(\d{1,2})\s*-?\s*day\b`)
	matches := re.FindAllStringSubmatch(lower, -1)
	best := 0
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		value, err := strconv.Atoi(match[1])
		if err == nil {
			best = max(best, value)
		}
	}
	return best
}

func clamp(value int, low int, high int) int {
	return min(max(value, low), high)
}

func saveTranscript(path string, messages []llm.Message) error {
	expanded := expandHome(path)
	if err := os.MkdirAll(filepath.Dir(expanded), 0o755); err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString("# Local LLM Transcript\n\n")
	for _, msg := range messages {
		role := "Assistant"
		if msg.Role == "user" {
			role = "User"
		}
		b.WriteString("## ")
		b.WriteString(role)
		b.WriteString("\n\n")
		b.WriteString(strings.TrimSpace(msg.Content))
		b.WriteString("\n\n")
	}
	return os.WriteFile(expanded, []byte(b.String()), 0o644)
}

func expandHome(path string) string {
	if path == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
	}
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}
	return path
}
