package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
)

type AES struct {
	secretKey []byte
}

func NewAES(secretKey []byte) *AES {
	return &AES{secretKey: secretKey}
}

func (a *AES) Encrypt(data []byte) ([]byte, error) {
	aesInst, err := aes.NewCipher(a.secretKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(aesInst)
	if err != nil {
		return nil, err
	}

	// We need a 12-byte nonce for GCM (modifiable if you use cipher.NewGCMWithNonceSize())
	// A nonce should always be randomly generated for every encryption.
	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, err
	}

	// ciphertext here is actually nonce+ciphertext
	// So that when we decrypt, just knowing the nonce size
	// is enough to separate it from the ciphertext.
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

func (a *AES) Decrypt(cipherData []byte) ([]byte, error) {
	aesInst, err := aes.NewCipher(a.secretKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(aesInst)
	if err != nil {
		return nil, err
	}

	// Since we know the ciphertext is actually nonce+ciphertext
	// And len(nonce) == NonceSize(). We can separate the two.
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := cipherData[:nonceSize], cipherData[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
