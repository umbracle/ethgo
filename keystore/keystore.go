package keystore

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/umbracle/go-web3"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
)

func getRand(size int) []byte {
	buf := make([]byte, size)
	rand.Read(buf)
	return buf
}

// EncryptV3 encrypts data in v3 format
func EncryptV3(content []byte, password string, customScrypt ...int) ([]byte, error) {

	// default scrypt values
	scryptN, scryptP := 1<<18, 1

	if len(customScrypt) >= 1 {
		scryptN = customScrypt[0]
	}
	if len(customScrypt) >= 2 {
		scryptP = customScrypt[1]
	}

	iv := getRand(aes.BlockSize)

	scrypt := scryptParams{
		N:     scryptN,
		R:     8,
		P:     scryptP,
		Dklen: 32,
		Salt:  hexString(getRand(32)),
	}
	kdf, err := scrypt.Key([]byte(password))
	if err != nil {
		return nil, err
	}

	cipherText, err := aesCTR(kdf[:16], content, iv)
	if err != nil {
		return nil, err
	}

	// generate mac
	mac := web3.Keccak256(kdf[16:32], cipherText)

	v3 := &v3Encoding{
		Version: 3,
		Crypto: &cryptoEncoding{
			Cipher:     "aes-128-ctr",
			CipherText: hexString(cipherText),
			CipherParams: struct{ IV hexString }{
				IV: hexString(iv),
			},
			KDF:       "scrypt",
			KDFParams: scrypt,
			Mac:       hexString(mac),
		},
	}

	encrypted, err := v3.Marshal()
	if err != nil {
		return nil, err
	}
	return encrypted, nil
}

// DecryptV3 decodes bytes in the v3 keystore format
func DecryptV3(content []byte, password string) ([]byte, error) {
	encoding := v3Encoding{}
	if err := encoding.Unmarshal(content); err != nil {
		return nil, err
	}
	if encoding.Version != 3 {
		return nil, fmt.Errorf("only version 3 supported")
	}
	if encoding.Crypto.Cipher != "aes-128-ctr" {
		return nil, fmt.Errorf("cipher %s not supported", encoding.Crypto.Cipher)
	}

	// decode the kdf
	kdf, err := encoding.Crypto.getKDF([]byte(password))
	if err != nil {
		return nil, err
	}

	// validate mac
	mac := web3.Keccak256(kdf[16:32], encoding.Crypto.CipherText)
	if !bytes.Equal(mac, encoding.Crypto.Mac) {
		return nil, fmt.Errorf("incorrect mac")
	}

	dst, err := aesCTR(kdf[:16], encoding.Crypto.CipherText, encoding.Crypto.CipherParams.IV)
	if err != nil {
		return nil, err
	}
	return dst, nil
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

type v3Encoding struct {
	ID      string          `json:"id"`
	Version int64           `json:"version"`
	Crypto  *cryptoEncoding `json:"crypto"`
}

func (j *v3Encoding) Marshal() ([]byte, error) {
	params, err := json.Marshal(j.Crypto.KDFParams)
	if err != nil {
		return nil, err
	}
	j.Crypto.KDFParamsRaw = json.RawMessage(params)
	return json.Marshal(j)
}

func (j *v3Encoding) Unmarshal(data []byte) error {
	return json.Unmarshal(data, j)
}

type cryptoEncoding struct {
	Cipher       string `json:"cipher"`
	CipherParams struct {
		IV hexString
	} `json:"cipherparams"`
	CipherText   hexString `json:"ciphertext"`
	KDF          string    `json:"kdf"`
	KDFParams    interface{}
	KDFParamsRaw json.RawMessage `json:"kdfparams"`
	Mac          hexString       `json:"mac"`
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

type scryptParams struct {
	Dklen int
	Salt  hexString
	N     int
	P     int
	R     int
}

func (s *scryptParams) Key(password []byte) ([]byte, error) {
	return scrypt.Key(password, s.Salt, s.N, s.R, s.P, s.Dklen)
}

func (c *cryptoEncoding) getKDF(password []byte) ([]byte, error) {
	var key []byte

	if c.KDF == "pbkdf2" {
		var params struct {
			Dklen int
			Salt  hexString
			C     int
			Prf   string
		}
		if err := json.Unmarshal(c.KDFParamsRaw, &params); err != nil {
			return nil, err
		}
		if params.Prf != "hmac-sha256" {
			return nil, fmt.Errorf("not found")
		}
		key = pbkdf2.Key(password, params.Salt, params.C, params.Dklen, sha256.New)
	} else if c.KDF == "scrypt" {
		var params scryptParams
		err := json.Unmarshal(c.KDFParamsRaw, &params)
		if err != nil {
			return nil, err
		}
		key, err = params.Key(password)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("kdf '%s' not supported", c.KDF)
	}
	return key, nil
}
