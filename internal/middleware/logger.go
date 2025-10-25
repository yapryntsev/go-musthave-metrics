package middleware

import (
    "go.uber.org/zap"
    "net/http"
    "time"
)

func Logger(l *zap.Logger) func(next http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        fn := func(w http.ResponseWriter, r *http.Request) {
            uri := r.RequestURI
            method := r.Method

            lw := loggingResponseWriter{
                ResponseWriter: w,
            }

            start := time.Now()
            next.ServeHTTP(&lw, r)
            duration := time.Since(start)

            var status int
            if lw.data.status == 0 {
                status = http.StatusOK
            } else {
                status = lw.data.status
            }

            l.Debug(
                "request handled",
                zap.String("uri", uri),
                zap.String("method", method),
                zap.Duration("duration", duration),
                zap.Int("status", status),
                zap.Int("size", lw.data.size),
            )
        }

        return http.HandlerFunc(fn)
    }
}

type (
    responseData struct {
        status int
        size   int
    }

    loggingResponseWriter struct {
        http.ResponseWriter
        data responseData
    }
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
    size, err := r.ResponseWriter.Write(b)
    r.data.size += size
    return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
    r.ResponseWriter.WriteHeader(statusCode)
    r.data.status = statusCode
}
