// Package secretbox seals short user-facing secrets — currently only lottery
// activation codes — with AES-256-GCM before they reach the database.
//
// Encryption is the second line here, not the first. The code's real exposure
// is an access path that returns it to the wrong reader, so the first line is
// that the sealed value never enters a DTO. This package exists so that a
// database dump, a backup, or a stray admin query is not also a code dump.
package secretbox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
)

type Box struct {
	aead cipher.AEAD
}

// New takes a 64-character hex string (32 bytes). An empty key yields a nil Box
// rather than an error: a forum without the key configured still boots, it just
// refuses to accept escrowed codes.
func New(hexKey string) (*Box, error) {
	if hexKey == "" {
		return nil, nil
	}
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("密钥不是合法的十六进制: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("密钥长度应为 32 字节 (64 个十六进制字符), 实际 %d 字节", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Box{aead: aead}, nil
}

func (b *Box) Enabled() bool { return b != nil && b.aead != nil }

func (b *Box) Seal(plaintext string) (string, error) {
	if !b.Enabled() {
		return "", errors.New("secretbox 未配置")
	}
	nonce := make([]byte, b.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := b.aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

func (b *Box) Open(sealed string) (string, error) {
	if !b.Enabled() {
		return "", errors.New("secretbox 未配置")
	}
	raw, err := base64.StdEncoding.DecodeString(sealed)
	if err != nil {
		return "", err
	}
	if len(raw) < b.aead.NonceSize() {
		return "", errors.New("密文长度不足")
	}
	nonce, body := raw[:b.aead.NonceSize()], raw[b.aead.NonceSize():]
	plaintext, err := b.aead.Open(nil, nonce, body, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
