package logger

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestInfof tests by info logging to a bytes.Buffer an empty string, normal string, and a string
// with unicode characters.
func TestInfof(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, Settings{Info: true})

	cases := []string{"", "this is a test", "😀 😃 😄 😁 😆 😅 😂"}

	for i, c := range cases {
		logger.Infof(c)
		out, err := buf.ReadString('\n')
		out = out[20:] // skip timestamp
		assert.Equal(t, err, nil, fmt.Sprintf("case: %d", i))
		assert.Equal(t, fmt.Sprintf("%s %s\n", infoPrefix, c), out, fmt.Sprintf("case: %d", i))
	}
}

// TestDebugf tests by debug logging to a bytes.Buffer an empty string, normal string, and a string
// with unicode characters.
func TestDebugf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, Settings{Debug: true})

	cases := []string{"", "this is a test", "😀 😃 😄 😁 😆 😅 😂"}

	for i, c := range cases {
		logger.Debugf(c)
		out, err := buf.ReadString('\n')
		out = out[20:] // skip timestamp
		assert.Equal(t, err, nil, fmt.Sprintf("case: %d", i))
		assert.Equal(t, fmt.Sprintf("%s %s\n", debugPrefix, c), out, fmt.Sprintf("case: %d", i))
	}
}

// TestErrorf tests by error logging to a bytes.Buffer an empty string, normal string, and a string
// with unicode characters.
func TestErrorf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, Settings{Error: true})

	cases := []string{"", "this is a test", "😀 😃 😄 😁 😆 😅 😂"}

	for i, c := range cases {
		logger.Errorf(c)
		out, err := buf.ReadString('\n')
		out = out[20:] // skip timestamp
		assert.Equal(t, err, nil, fmt.Sprintf("case: %d", i))
		assert.Equal(t, fmt.Sprintf("%s %s\n", errorPrefix, c), out, fmt.Sprintf("case: %d", i))
	}
}

// TestMiddleware tests by creating an http handler to echo the sent body, wrapping that with the logger, and
// making sure that: the request is logged properly, the echo handler is still functional.
func TestMiddleware(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, Settings{Debug: true})

	h := logger.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(w, r.Body)
	}))

	echo := []byte("test")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", bytes.NewBuffer(echo))

	h.ServeHTTP(w, r)

	out, err := buf.ReadString('\n')
	out = out[20:]
	assert.Equal(t, err, nil)
	assert.Equal(t, strings.HasPrefix(out, fmt.Sprintf("%s %s\t%s\t", debugPrefix, "GET", "/")), true)
	assert.Equal(t, w.Body.Bytes(), echo)
}
