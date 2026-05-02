package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"errors"
	"log"
	"os"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

func Serve(addr, keyPath string, hub *Hub) error {
	signer, err := loadOrCreateHostKey(keyPath)
	if err != nil {
		return err
	}

	server := &ssh.Server{
		Addr: addr,
		Handler: func(s ssh.Session) {
			handleSession(s, hub)
		},
		PasswordHandler: func(ssh.Context, string) bool {
			return true
		},
		PublicKeyHandler: func(ssh.Context, ssh.PublicKey) bool {
			return true
		},
	}
	server.AddHostKey(signer)

	log.Printf("chat.ssh listening on %s", addr)
	return server.ListenAndServe()
}

func loadOrCreateHostKey(path string) (gossh.Signer, error) {
	if data, err := os.ReadFile(path); err == nil {
		return gossh.ParsePrivateKey(data)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	block, err := gossh.MarshalPrivateKey(priv, "")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, pem.EncodeToMemory(block), 0600); err != nil {
		return nil, err
	}
	log.Printf("generated host key at %s", path)
	return gossh.NewSignerFromKey(priv)
}
