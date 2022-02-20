package keystore

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
)

func getRand(size int) []byte {
	buf := make([]byte, size)
	rand.Read(buf)
	return buf
}

type hexString []byte

func (h hexString) MarshalJSON() ([]byte, error) {
	str := "\"" + hex.EncodeToString(h) + "\""
	return []byte(str), nil
}

func (h *hexString) UnmarshalJSON(data []byte) error {
	raw := string(data)
	raw = strings.Trim(raw, "\"")

	data, err := hex.DecodeString(raw)
	if err != nil {
		return err
	}
	*h = data
	return nil
}

func aesCTR(key, cipherText, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(block, iv)

	dst := make([]byte, len(cipherText))
	stream.XORKeyStream(dst, cipherText)

	return dst, nil
}

type pbkdf2Params struct {
	Dklen int       `json:"dklen"`
	Salt  hexString `json:"salt"`
	C     int       `json:"c"`
	Prf   string    `json:"prf"`
}

func (p *pbkdf2Params) Key(password []byte) []byte {
	return pbkdf2.Key(password, p.Salt, p.C, p.Dklen, sha256.New)
}

type scryptParams struct {
	Dklen int       `json:"dklen"`
	Salt  hexString `json:"salt"`
	N     int       `json:"n"`
	P     int       `json:"p"`
	R     int       `json:"r"`
}

func (s *scryptParams) Key(password []byte) ([]byte, error) {
	return scrypt.Key(password, s.Salt, s.N, s.R, s.P, s.Dklen)
}

func applyKdf(fn string, password, paramsRaw []byte) ([]byte, error) {
	var key []byte

	if fn == "pbkdf2" {
		var params pbkdf2Params
		if err := json.Unmarshal(paramsRaw, &params); err != nil {
			return nil, err
		}
		if params.Prf != "hmac-sha256" {
			return nil, fmt.Errorf("not found")
		}
		key = params.Key(password)
	} else if fn == "scrypt" {
		var params scryptParams
		err := json.Unmarshal(paramsRaw, &params)
		if err != nil {
			return nil, err
		}
		key, err = params.Key(password)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("kdf '%s' not supported", fn)
	}
	return key, nil
}
