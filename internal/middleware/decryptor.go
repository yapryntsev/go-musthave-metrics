package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"net/http"

	"go.uber.org/zap"
)

func Decryptor(key *rsa.PrivateKey, log *zap.Logger) func(next http.Handler) http.Handler {
	if key == nil {
		log.Debug("body sign middleware disabled. no key found")
	}

	return func(next http.Handler) http.Handler {
		if key == nil {
			return next
		}

		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					log.Error("failed to read body", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				msg, err := rsa.DecryptPKCS1v15(rand.Reader, key, body)
				if err != nil {
					log.Error("failed to decode body", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				r.Body = io.NopCloser(bytes.NewReader(msg))
			},
		)
	}
}
