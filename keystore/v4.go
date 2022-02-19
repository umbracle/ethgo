package keystore

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/text/unicode/norm"
)

func EncryptV4(content []byte, password string) ([]byte, error) {
	password = normalizePassword(password)

	// decryption key
	scrypt := scryptParams{
		N:     1 << 18,
		R:     8,
		P:     1,
		Dklen: 32,
		Salt:  hexString(getRand(32)),
	}
	key, err := scrypt.Key([]byte(password))
	if err != nil {
		return nil, err
	}

	// decrypt
	iv := getRand(16)
	cipherText, err := aesCTR(key[:16], content, iv)
	if err != nil {
		return nil, err
	}

	// checksum
	hash := sha256.New()
	hash.Write(key[16:32])
	hash.Write(cipherText)

	checksum := hash.Sum(nil)

	kdfParams, err := json.Marshal(scrypt)
	if err != nil {
		return nil, err
	}
	cipherParams, err := json.Marshal(&cipherParams{Iv: hexString(iv)})
	if err != nil {
		return nil, err
	}

	encoding := &v4Encoding{
		Version: 4,
		Crypto: &v4crypto{
			Kdf: &v4Module{
				Function: "scrypt",
				Params:   kdfParams,
			},
			Cipher: &v4Module{
				Function: "aes-128-ctr",
				Params:   cipherParams,
				Message:  hexString(cipherText),
			},
			Checksum: &v4Module{
				Function: "sha256",
				Message:  hexString(checksum),
			},
		},
	}
	return encoding.Marshal()
}

type cipherParams struct {
	Iv hexString `json:"iv"`
}

func DecryptV4(content []byte, password string) ([]byte, error) {
	encoding := v4Encoding{}
	if err := encoding.Unmarshal(content); err != nil {
		return nil, err
	}
	if encoding.Version != 4 {
		return nil, fmt.Errorf("only version 4 supported")
	}

	password = normalizePassword(password)

	// decryption key
	key, err := applyKdf(encoding.Crypto.Kdf.Function, []byte(password), encoding.Crypto.Kdf.Params)
	if err != nil {
		return nil, err
	}

	// checksum
	hash := sha256.New()
	hash.Write(key[16:32])
	hash.Write(encoding.Crypto.Cipher.Message)

	checksum := hash.Sum(nil)
	if !bytes.Equal(checksum, encoding.Crypto.Checksum.Message) {
		return nil, fmt.Errorf("bad checksum")
	}

	// decrypt
	var msg []byte
	if encoding.Crypto.Cipher.Function == "aes-128-ctr" {
		var params cipherParams
		if err := json.Unmarshal(encoding.Crypto.Cipher.Params, &params); err != nil {
			return nil, err
		}
		res, err := aesCTR(key[:16], encoding.Crypto.Cipher.Message, params.Iv)
		if err != nil {
			return nil, err
		}
		msg = res
	} else {
		return nil, fmt.Errorf("cipher '%s' not supported", encoding.Crypto.Cipher.Function)
	}
	return msg, nil
}

type v4Encoding struct {
	Crypto      *v4crypto `json:"crypto"`
	Description string    `json:"description"`
	PubKey      hexString `json:"pubkey"`
	Path        string    `json:"path"`
	Version     int       `json:"version"`
	Uuid        string    `json:"uuid"`
}

func (j *v4Encoding) Marshal() ([]byte, error) {
	return json.Marshal(j)
}

func (j *v4Encoding) Unmarshal(data []byte) error {
	return json.Unmarshal(data, j)
}

type v4crypto struct {
	Kdf      *v4Module `json:"kdf"`
	Checksum *v4Module `json:"checksum"`
	Cipher   *v4Module `json:"cipher"`
}

type v4Module struct {
	Function string          `json:"function"`
	Params   json.RawMessage `json:"params"`
	Message  hexString       `json:"message"`
}

// normalizePassword normalizes the password following the next rules
// https://eips.ethereum.org/EIPS/eip-2335#password-requirements
func normalizePassword(password string) string {
	str := norm.NFKD.String(password)

	skip := func(i byte) bool {
		// skip runes in the range 0x00 - 0x1F, 0x80 - 0x9F and 0x7F
		if i == 0x7F {
			return true
		}
		if 0x00 <= i && i <= 0x1F {
			return true
		}
		if 0x80 <= i && i <= 0x9F {
			return true
		}
		return false
	}

	normalized := strings.Builder{}
	for _, r := range str {
		elem := string(r)
		if len(elem) == 1 {
			if skip(elem[0]) {
				continue
			}
		}
		normalized.WriteRune(r)
	}
	return normalized.String()
}
