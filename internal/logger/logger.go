package logger

import (
    log "github.com/sirupsen/logrus"
    "net/http"
    "time"
)

func New(system string) *log.Entry {
    return log.WithField("system", system)
}

func Middleware(h http.HandlerFunc, logger *log.Entry) http.HandlerFunc {
    return http.HandlerFunc(
        func(w http.ResponseWriter, r *http.Request) {
            uri := r.RequestURI
            method := r.Method

            lw := loggingResponseWriter{
                ResponseWriter: w,
            }

            start := time.Now()
            h(&lw, r)
            duration := time.Since(start)

            var status int
            if lw.data.status == 0 {
                status = http.StatusOK
            } else {
                status = lw.data.status
            }

            logger.WithFields(
                log.Fields{
                    "uri":      uri,
                    "method":   method,
                    "duration": duration,
                    "status":   status,
                    "size":     lw.data.size,
                },
            ).Info("request handled")
        },
    )
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
