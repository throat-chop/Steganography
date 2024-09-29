package Data

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"golang.org/x/crypto/argon2"
	"io"
	"os"
	"path"
)

type Data struct {
	Data []byte
	Size uint32
}

func NewData(filepath string, passphrase string, erchan chan<- error, result chan<- *Data) {

	ext := make([]byte, 16)
	copy(ext[:8], path.Ext(filepath))
	file, err := os.Open(filepath)
	defer file.Close()
	if err != nil {
		erchan <- err
		return
	}

	stat, err := file.Stat()
	if err != nil {
		erchan <- err
		return
	}

	size := make([]byte, 4)

	data, err := io.ReadAll(file)

	if err != nil {
		erchan <- err
		return
	}

	if passphrase != "" {

		encdata, enc, err := encrypt(data, ext, passphrase)
		if err != nil {
			erchan <- err
			return
		}
		binary.BigEndian.PutUint32(size, uint32((len(encdata)+48)*-1))
		meta := append(size, enc...)
		encdata = append(meta, encdata...)
		result <- &Data{
			Data: encdata,
			Size: uint32(len(encdata)),
		}

	} else {
		binary.BigEndian.PutUint32(size, uint32(stat.Size())+64)
		meta := append(size, make([]byte, 44)...)
		meta = append(meta, ext...)
		_, _ = rand.Read(meta[4:48])
		data = append(meta, data...)
		result <- &Data{
			Data: data,
			Size: uint32(len(data)),
		}
	}

}

func hash(passphrase string) ([]byte, []byte) {

	salt := make([]byte, 32)
	_, _ = rand.Read(salt)
	hash := argon2.IDKey([]byte(passphrase), salt, 13, 26*1024, 4, 32)
	return hash[:32], salt
}

func encrypt(data []byte, ext []byte, passphrase string) ([]byte, []byte, error) {
	key, salt := hash(passphrase)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	iv := make([]byte, aesGCM.NonceSize())
	_, _ = rand.Read(iv)
	encrypted := aesGCM.Seal(nil, iv, append(ext, data...), nil)
	ext = aesGCM.Seal(nil, iv, ext, nil)
	return encrypted, append(salt, iv...), nil
}

func Decode(resultfile string, data []byte, passphrase string) (string, error) {

	var filedata []byte
	var xtn string
	size := int32(binary.BigEndian.Uint32(data[:4]))

	if size > 0 {

		filedata = data[64:]
		xtn = string(bytes.TrimRight(data[48:64], "\x00"))

	} else if size < 0 {

		if passphrase == "" {
			return "", fmt.Errorf("encrypted data provide passphrase")
		}

		decrypted, err := decrypt(data, passphrase)
		if err != nil {
			return "", err
		}

		filedata = decrypted[16:]
		xtn = string(bytes.TrimRight(decrypted[:16], "\x00"))
	}

	file, err := os.OpenFile(resultfile+xtn, os.O_RDWR|os.O_CREATE, 0600)
	defer file.Close()
	if err != nil {
		return "", err
	}

	_, err = file.Write(filedata)
	if err != nil {
		return "", err
	}
	return xtn, nil
}
func decrypt(data []byte, passphrase string) ([]byte, error) {
	salt := data[4:36]
	iv := data[36:48]

	key := argon2.IDKey([]byte(passphrase), salt, 13, 26*1024, 4, 32)

	block, err := aes.NewCipher(key[:32])
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	decrypted, err := aesGCM.Open(nil, iv, data[48:], nil)
	if err != nil {
		return nil, err
	}
	return decrypted, nil

}
