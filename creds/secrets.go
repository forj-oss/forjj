package creds

import (
	"errors"
	"gopkg.in/yaml.v2"
	"io"
	"crypto/cipher"
	"crypto/aes"
	"io/ioutil"
	"fmt"
	"os"
	"encoding/base64"
	"crypto/rand"
)

// Secrets is internal secret structured shared with ci to run forjj jobs
type Secrets struct{
	key         []byte
	key64       string
	Envs        map[string]yamlSecure `yaml:";inline"`
}

const KeySize = 32

// NewSecrets creates the internal secret object to shared with CI running infra/deploy repositories.
func NewSecrets() (ret *Secrets) {
	ret = new(Secrets)
	ret.Envs = make(map[string]yamlSecure)
	return
}

// GenerateKey help to create a random key for the encryption
func (s *Secrets)GenerateKey() error {
	if s == nil {
		return fmt.Errorf("Secret object is nil")
	}
	key := make([]byte, KeySize)
	if _, err := rand.Read(key) ; err != nil {
		return err
	}
	s.key = key
	s.key64 = base64.StdEncoding.EncodeToString(key)

	return nil
}

// SaveKey save the key in a file
func (s *Secrets)SaveKey(file string) (error) {
	if s == nil {
		return fmt.Errorf("Secret object is nil")
	}
	if s.key == nil || len(s.key) != KeySize {
		return fmt.Errorf("Key is missing")
	}

	fd, err := os.Create(file)
	if err != nil {
		return err
	}
	defer fd.Close()

	_, err = fd.WriteString(s.key64)
	return err
}

// ReadKey read a file containing the key
func (s *Secrets)ReadKey(file string) (error) {
	if s == nil {
		return fmt.Errorf("Secret object is nil")
	}
	fd, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fd.Close()
	var key64 []byte

	key64, err = ioutil.ReadAll(fd)
	if err != nil {
		return err
	}
	s.key64 = string(key64)
	s.key, err = base64.StdEncoding.DecodeString(s.key64)
	if err != nil {
		return err
	}
	if v := len(s.key) ; v != KeySize {
		return fmt.Errorf("Invalid key. Size incorrect. must be %d. Got %d", KeySize, v)
	}
	return nil
}

// Key64 return the base64 of the internal key
func (s *Secrets)Key64() string {
	if s == nil {
		return ""
	}
	return s.key64
}

// Export provides an extraction of forjj secrets encrypted.
func (s *Secrets) Export() ([]byte, error) {
	if s == nil {
		return nil, fmt.Errorf("Secret object is nil")
	}
	if s.key == nil || len(s.key) != KeySize {
		return nil, fmt.Errorf("Key is missing")
	}

	c, err := aes.NewCipher(s.key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(c)
    if err != nil {
        return nil, err
    }

    nonce := make([]byte, gcm.NonceSize())
    if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
	}
	
	var secretData []byte
	secretData, err = yaml.Marshal(s)

    return gcm.Seal(nonce, nonce, secretData, nil), nil
}

// Import read an encrypted data, decrypt it and save it in Secrets
func (s *Secrets) Import(ciphertext []byte) error {
	if s == nil {
		return fmt.Errorf("Secret object is nil")
	}
	if s.key == nil || len(s.key) != KeySize {
		return fmt.Errorf("Key is missing")
	}

	c, err := aes.NewCipher(s.key)
    if err != nil {
        return err
    }

    gcm, err := cipher.NewGCM(c)
    if err != nil {
        return err
    }

    nonceSize := gcm.NonceSize()
    if len(ciphertext) < nonceSize {
        return errors.New("ciphertext too short")
    }

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	var secretData []byte
	secretData, err = gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return err
    }

	return yaml.Unmarshal(secretData, s)
}