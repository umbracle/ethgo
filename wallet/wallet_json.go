package wallet

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/mitchellh/mapstructure"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
)

func NewJSONWalletFromFile(path string, password string) (*Key, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return NewJSONWalletFromContent(data, password)
}

func NewJSONWalletFromContent(content []byte, password string) (*Key, error) {
	var encoding jsonWalletV3Encoding
	if err := json.Unmarshal(content, &encoding); err != nil {
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

	block, err := aes.NewCipher(kdf[:16])
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(block, encoding.Crypto.CipherParams.IV)

	dst := make([]byte, len(encoding.Crypto.CipherText))
	stream.XORKeyStream(dst, encoding.Crypto.CipherText)

	key, err := NewWalletFromPrivKey(dst)
	if err != nil {
		return nil, err
	}
	return key, nil
}

type jsonWalletV3Encoding struct {
	ID      string          `json:"id"`
	Version int64           `json:"version"`
	Crypto  *cryptoEncoding `json:"crypto"`
}

type cryptoEncoding struct {
	Cipher       string `json:"cipher"`
	CipherParams struct {
		IV hexString
	} `json:"cipherparams"`
	CipherText hexString              `json:"ciphertext"`
	KDF        string                 `json:"kdf"`
	KDFParams  map[string]interface{} `json:"kdfparams"`
	Mac        hexString              `json:"mac"`
}

type hexString []byte

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

func (c *cryptoEncoding) getKDF(password []byte) ([]byte, error) {
	var key []byte

	if c.KDF == "pbkdf2" {
		var params struct {
			Dklen int
			Salt  string
			C     int
			Prf   string
		}
		if err := mapstructure.Decode(c.KDFParams, &params); err != nil {
			return nil, err
		}
		if params.Prf != "hmac-sha256" {
			return nil, fmt.Errorf("not found")
		}
		salt, err := hex.DecodeString(params.Salt)
		if err != nil {
			return nil, err
		}
		key = pbkdf2.Key(password, salt, params.C, params.Dklen, sha256.New)
	} else if c.KDF == "scrypt" {
		var params struct {
			Dklen int
			Salt  string
			N     int
			P     int
			R     int
		}
		if err := mapstructure.Decode(c.KDFParams, &params); err != nil {
			return nil, err
		}
		salt, err := hex.DecodeString(params.Salt)
		if err != nil {
			return nil, err
		}
		key, err = scrypt.Key(password, salt, params.N, params.R, params.P, params.Dklen)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("kdf '%s' not supported", c.KDF)
	}

	// validate mac
	mac := keccak256(key[16:32], c.CipherText)
	if !bytes.Equal(mac, c.Mac) {
		return nil, fmt.Errorf("incorrect mac")
	}
	return key, nil
}
