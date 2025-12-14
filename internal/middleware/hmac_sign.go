package crypto

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
)

func hash(msg []byte, key string) string {
    hash := hmac.New(sha256.New, []byte(key))
    hash.Write(msg)

    return hex.EncodeToString(hash.Sum(nil))
}
