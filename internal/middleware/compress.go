package middleware

import (
    "compress/gzip"
    "fmt"
    "go.uber.org/zap"
    "io"
    "net/http"
    "strings"
)

func Compress(log *zap.Logger) func(next http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        fn := func(w http.ResponseWriter, r *http.Request) {
            if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
                zr, err := gzip.NewReader(r.Body)
                if err != nil {
                    w.WriteHeader(http.StatusInternalServerError)
                    log.Error("failed to create gzip reader", zap.Error(err))

                    return
                }

                cr := &compressReader{
                    r:  r.Body,
                    zr: zr,
                }
                defer cr.Close()

                r.Body = cr
            }

            ow := w
            if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
                cw := &compressWriter{
                    w:  w,
                    zw: gzip.NewWriter(w),
                }
                defer cw.Close()

                ow = cw
            }

            next.ServeHTTP(ow, r)
        }

        return http.HandlerFunc(fn)
    }
}

type compressReader struct {
    r  io.ReadCloser
    zr *gzip.Reader
}

func (c *compressReader) Read(p []byte) (n int, err error) {
    return c.zr.Read(p)
}

func (c *compressReader) Close() error {
    gzipErr := c.zr.Close()
    rErr := c.r.Close()

    var err error

    if gzipErr != nil {
        err = fmt.Errorf("failed to close gzip reader: %w", gzipErr)
    }

    if rErr != nil {
        err = fmt.Errorf("failed to close body reader: %w: %w", rErr, err)
    }

    return err
}

type compressWriter struct {
    w         http.ResponseWriter
    zw        *gzip.Writer
    wroteBody bool
}

func (c *compressWriter) Header() http.Header {
    return c.w.Header()
}

func (c *compressWriter) Write(bytes []byte) (int, error) {
    if !c.wroteBody {
        c.w.Header().Set("Content-Encoding", "gzip")
        c.wroteBody = true
    }

    return c.zw.Write(bytes)
}

func (c *compressWriter) WriteHeader(statusCode int) {
    c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
    return c.zw.Close()
}
