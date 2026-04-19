package middlewares

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"slices"
	"strings"
)

type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
	compressed bool
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
    return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	log.Printf("compressWriter.Write: len=%d, compressed=%v\n", len(p), c.compressed)
	log.Printf("Content-Type: %s\n", c.w.Header().Get("Content-Type"))

	if c.compressed {
        return c.zw.Write(p)
    }

	contentType := c.w.Header().Get("Content-Type")
	supportTypes := []string{"text/html", "application/json"}

	if slices.Contains(supportTypes, contentType) {
		c.compressed = true
		c.w.Header().Set("Content-Encoding", "gzip")
		c.w.Header().Del("Content-Length")
		return c.zw.Write(p)
	}
	return c.w.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
    c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	if c.compressed {
		return c.zw.Close()
	}
    return nil
}

type compressReader struct {
    r  io.ReadCloser
    zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
    zr, err := gzip.NewReader(r)
    if err != nil {
        return nil, err
    }

    return &compressReader{
        r:  r,
        zr: zr,
    }, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
    return c.zr.Read(p)
}

func (c *compressReader) Close() error {
    if err := c.r.Close(); err != nil {
        return err
    }
    return c.zr.Close()
} 

func Gzip(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Accept-Encoding:", r.Header.Get("Accept-Encoding"))
        log.Println("Content-Encoding:", r.Header.Get("Content-Encoding"))

		if r.URL.Path == "/api/shorten" && r.Method == "POST" {
            h.ServeHTTP(w, r)
            return
        }
		
        ow := w

        acceptEncoding := r.Header.Get("Accept-Encoding")
        supportsGzip := strings.Contains(strings.ToLower(acceptEncoding), "gzip")
		log.Println("supportsGzip:", supportsGzip)
        if supportsGzip {
            cw := newCompressWriter(w)
            ow = cw
            defer func(){
				if cw.compressed {
					cw.Close()
				}
			}()
        }

        contentEncoding := r.Header.Get("Content-Encoding")
        sendsGzip := strings.Contains(contentEncoding, "gzip")
        if sendsGzip {
            cr, err := newCompressReader(r.Body)
            if err != nil {
                w.WriteHeader(http.StatusInternalServerError)
                return
            }
            r.Body = cr
            defer cr.Close()
        }

        h.ServeHTTP(ow, r)
    })
}