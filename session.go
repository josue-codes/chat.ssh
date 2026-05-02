package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/gliderlabs/ssh"
	"golang.org/x/term"
)

const (
	maxUsernameLen = 20
	outboxSize     = 64
)

func handleSession(s ssh.Session, hub *Hub) {
	pty, winCh, isPty := s.Pty()
	if !isPty {
		io.WriteString(s, "this server requires a PTY (try: ssh -t)\n")
		s.Exit(1)
		return
	}

	t := term.NewTerminal(s, "")
	t.SetSize(pty.Window.Width, pty.Window.Height)
	go func() {
		for win := range winCh {
			t.SetSize(win.Width, win.Height)
		}
	}()

	io.WriteString(t, "welcome to chat.ssh — type /quit to leave\n")

	username, ok := promptUsername(t)
	if !ok {
		return
	}

	t.SetPrompt(fmt.Sprintf("[%s] > ", username))

	client := &Client{
		Username: username,
		Out:      make(chan Message, outboxSize),
	}
	hub.Join(client)
	defer hub.Leave(client)

	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		for m := range client.Out {
			if _, err := io.WriteString(t, formatMessage(m)+"\n"); err != nil {
				return
			}
		}
	}()

	for {
		line, err := t.ReadLine()
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line == "/quit" {
			break
		}
		hub.Send(username, line)
	}

	close(client.Out)
	<-writerDone
}

func promptUsername(t *term.Terminal) (string, bool) {
	t.SetPrompt("username: ")
	for {
		line, err := t.ReadLine()
		if err != nil {
			return "", false
		}
		name := strings.TrimSpace(line)
		if !validUsername(name) {
			io.WriteString(t, "username must be 1-20 chars: letters, digits, '-' or '_'\n")
			continue
		}
		return name, true
	}
}

func validUsername(s string) bool {
	if len(s) == 0 || len(s) > maxUsernameLen {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_':
		default:
			return false
		}
	}
	return true
}

func formatMessage(m Message) string {
	ts := m.Time.Format("15:04:05")
	if m.From == "" {
		return fmt.Sprintf("%s %s", ts, m.Text)
	}
	return fmt.Sprintf("%s <%s> %s", ts, m.From, m.Text)
}
