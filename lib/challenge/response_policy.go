package challenge

import (
	"mime"
	"net/http"
	"path"
	"strings"
)

const broadVaryHeaderValue = "Cookie, Accept, Accept-Encoding, Accept-Language, User-Agent"

type responsePolicyWriter struct {
	http.ResponseWriter
	request     *http.Request
	data        *RequestData
	wroteHeader bool
}

// NewResponsePolicyWriter applies cache policy after the backend response
// headers are available but before they are committed to the client.
func NewResponsePolicyWriter(w http.ResponseWriter, r *http.Request, data *RequestData) http.ResponseWriter {
	return &responsePolicyWriter{
		ResponseWriter: w,
		request:        r,
		data:           data,
	}
}

func (w *responsePolicyWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *responsePolicyWriter) WriteHeader(status int) {
	if status >= 100 && status < 200 && status != http.StatusSwitchingProtocols {
		w.ResponseWriter.WriteHeader(status)
		return
	}
	if w.wroteHeader {
		return
	}

	w.wroteHeader = true
	w.apply(status)
	w.ResponseWriter.WriteHeader(status)
}

func (w *responsePolicyWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(data)
}

func (w *responsePolicyWriter) FlushError() error {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return http.NewResponseController(w.ResponseWriter).Flush()
}

func (w *responsePolicyWriter) apply(status int) {
	if w.data == nil || !w.data.HasAnyValidChallenge() {
		return
	}
	if !isImmutableMediaResponse(w.request, w.Header(), status) {
		return
	}

	makeCacheControlPrivate(w.Header())
	if w.data.broadVaryApplied {
		removeFirstHeaderValue(w.Header(), "Vary", broadVaryHeaderValue)
	}
}

func isImmutableMediaResponse(r *http.Request, header http.Header, status int) bool {
	if r == nil || (r.Method != http.MethodGet && r.Method != http.MethodHead) {
		return false
	}
	if (status < 200 || status >= 300) && status != http.StatusNotModified {
		return false
	}
	if !hasCacheControlDirective(header, "immutable") {
		return false
	}

	if mediaType, _, err := mime.ParseMediaType(header.Get("Content-Type")); err == nil {
		if strings.HasPrefix(mediaType, "image/") ||
			strings.HasPrefix(mediaType, "audio/") ||
			strings.HasPrefix(mediaType, "video/") {
			return true
		}
	}

	switch strings.ToLower(path.Ext(r.URL.Path)) {
	case ".avif", ".bmp", ".flac", ".gif", ".ico", ".jpeg", ".jpg", ".jxl",
		".mp3", ".mp4", ".ogg", ".png", ".svg", ".wav", ".webm", ".webp":
		return true
	default:
		return false
	}
}

func hasCacheControlDirective(header http.Header, expected string) bool {
	for _, value := range header.Values("Cache-Control") {
		for _, directive := range strings.Split(value, ",") {
			name, _, _ := strings.Cut(strings.TrimSpace(directive), "=")
			if strings.EqualFold(name, expected) {
				return true
			}
		}
	}
	return false
}

func makeCacheControlPrivate(header http.Header) {
	directives := []string{"private"}
	for _, value := range header.Values("Cache-Control") {
		for _, directive := range strings.Split(value, ",") {
			directive = strings.TrimSpace(directive)
			if directive == "" {
				continue
			}

			name, _, _ := strings.Cut(directive, "=")
			switch strings.ToLower(strings.TrimSpace(name)) {
			case "private", "public", "s-maxage":
				continue
			default:
				directives = append(directives, directive)
			}
		}
	}
	header.Set("Cache-Control", strings.Join(directives, ", "))
}

func removeFirstHeaderValue(header http.Header, name, value string) {
	values := header.Values(name)
	kept := make([]string, 0, len(values))
	removed := false
	for _, candidate := range values {
		if !removed && candidate == value {
			removed = true
			continue
		}
		kept = append(kept, candidate)
	}
	if !removed {
		return
	}

	header.Del(name)
	for _, candidate := range kept {
		header.Add(name, candidate)
	}
}
