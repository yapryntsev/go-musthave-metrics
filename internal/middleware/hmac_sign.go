package middleware

import (
    "bytes"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "errors"
    "go.uber.org/zap"
    "hash"
    "io"
    "net/http"
)

const SignedBodyHeader = "HashSHA256"

func SignBody(key string, log *zap.Logger) func(next http.Handler) http.Handler {
    if key == "" {
        log.Debug("body sign middleware disabled. no key found")
    }

    return func(next http.Handler) http.Handler {
        if key == "" {
            return next
        }

        fn := func(w http.ResponseWriter, r *http.Request) {
            signature := r.Header.Get(SignedBodyHeader)
            if signature == "" {
                next.ServeHTTP(w, r)
                return
            }

            body, err := io.ReadAll(r.Body)
            if err != nil {
                log.Error("failed to read body", zap.Error(err))
                return
            }
            r.Body = io.NopCloser(bytes.NewReader(body))

            if !isSignatureValid(signature, body, NewHasher(key)) {
                w.WriteHeader(http.StatusBadRequest)
                return
            }

            hw := &hashWriter{
                ResponseWriter: w,
                hasher:         NewHasher(key),
            }
            next.ServeHTTP(hw, r)

            err = hw.Flush()
            if err != nil {
                log.Error("failed to sign body", zap.Error(err))
                return
            }
        }

        return http.HandlerFunc(fn)
    }
}

func NewHasher(key string) hash.Hash {
    return hmac.New(sha256.New, []byte(key))
}

func isSignatureValid(signature string, body []byte, hasher hash.Hash) bool {
    if signature == "" {
        return false
    }

    hasher.Write(body)
    hashSum := hasher.Sum(nil)
    return hex.EncodeToString(hashSum) == signature
}

type hashWriter struct {
    http.ResponseWriter
    buf        bytes.Buffer
    hasher     hash.Hash
    statusCode int
}

func (h *hashWriter) Write(bytes []byte) (int, error) {
    var errs []error

    _, err := h.hasher.Write(bytes)
    if err != nil {
        errs = append(errs, err)
    }

    n, err := h.buf.Write(bytes)
    if err != nil {
        errs = append(errs, err)
    }

    return n, errors.Join(errs...)
}

func (h *hashWriter) WriteHeader(statusCode int) {
    h.statusCode = statusCode
}

func (h *hashWriter) Flush() error {
    defer func() {
        if h.statusCode != 0 {
            h.ResponseWriter.WriteHeader(h.statusCode)
        }
    }()

    if h.buf.Len() == 0 {
        return nil
    }

    hashSum := h.hasher.Sum(nil)
    h.ResponseWriter.Header().Set(SignedBodyHeader, hex.EncodeToString(hashSum))

    _, err := h.ResponseWriter.Write(h.buf.Bytes())
    if err != nil {
        return err
    }

    return nil
}
