/* Project Encore: BFG - Localized Private Game Restoration Server
 * Copyright (C) 2026 Paficent <paficent@tutamail.com> & Contributors
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package utils

import (
	"crypto/aes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"os"
)

// should eventually use bcrypt, this isnt secure
func HashPassword(plain string) string {
	sum := sha256.Sum256([]byte("bfg:" + plain))
	return hex.EncodeToString(sum[:])
}

func CheckPassword(plain, hashed string) bool {
	return HashPassword(plain) == hashed
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	pad := blockSize - len(data)%blockSize
	out := make([]byte, len(data)+pad)
	copy(out, data)
	for i := len(data); i < len(out); i++ {
		out[i] = byte(pad)
	}
	return out
}

func aesCFB8Encrypt(plaintext, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	shift := make([]byte, bs)
	copy(shift, iv)
	tmp := make([]byte, bs)
	out := make([]byte, len(plaintext))
	for i := range plaintext {
		block.Encrypt(tmp, shift)
		c := plaintext[i] ^ tmp[0]
		out[i] = c
		copy(shift, shift[1:])
		shift[bs-1] = c
	}
	return out, nil
}

// session token
func EncryptToken(message, iv, key string) string {
	k := []byte(key)
	if len(k) > 16 {
		k = k[:16]
	}
	ivb := make([]byte, 16)
	copy(ivb, []byte(iv))
	ct, err := aesCFB8Encrypt(pkcs7Pad([]byte(message), 16), k, ivb)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(ct)
}

func MD5File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
