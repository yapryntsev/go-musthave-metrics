package middleware

import (
    "compress/gzip"
    "io"
    "net/http"
    "strings"
)

func Compress(next http.Handler) http.Handler {
    fn := func(w http.ResponseWriter, r *http.Request) {
        ow := w

        if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
            zr, err := gzip.NewReader(r.Body)
            if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }

            cr := &compressReader{
                r:  r.Body,
                zr: zr,
            }
            defer cr.Close()

            r.Body = cr
        }

        if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
            cw := &compressWriter{
                w:  w,
                zw: gzip.NewWriter(w),
                writeHeader: func(w http.ResponseWriter) {
                    w.Header().Set("Content-Encoding", "gzip")
                },
            }
            defer cw.Close()

            ow = cw
        }

        next.ServeHTTP(ow, r)
    }

    return http.HandlerFunc(fn)
}

type compressWriter struct {
    w           http.ResponseWriter
    zw          *gzip.Writer
    wroteHeader bool
    writeHeader func(w http.ResponseWriter)
}

func (c *compressWriter) Header() http.Header {
    return c.w.Header()
}

func (c *compressWriter) Write(bytes []byte) (int, error) {
    if c.writeHeader != nil && !c.wroteHeader {
        c.writeHeader(c.w)
        c.wroteHeader = true
    }

    return c.zw.Write(bytes)
}

func (c *compressWriter) WriteHeader(statusCode int) {
    c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
    return c.zw.Close()
}

type compressReader struct {
    r  io.ReadCloser
    zr *gzip.Reader
}

func (c *compressReader) Read(p []byte) (n int, err error) {
    return c.zr.Read(p)
}

func (c *compressReader) Close() error {
    return c.zr.Close()
}
