package http

import (
	"encoding/json"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/httpheader"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestIDInjectorMiddleware(t *testing.T) {
	const testRequestID = "test-req-123"

	newRequest := func() *http.Request {
		r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		ctx := ctxutils.NewHttpContext(r.Context(), &ctxutils.HttpRequestContext{
			RequestId: testRequestID,
		})
		return r.WithContext(ctx)
	}

	t.Run("injects requestId when absent", func(t *testing.T) {
		handler := requestIDInjectorMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(httpheader.ContentType, httpheader.ContentTypeApplicationJSON)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"memberships":[]}`))
		}))

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, newRequest())

		var body map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, http.StatusOK, rec.Code)

		var gotID string
		require.NoError(t, json.Unmarshal(body["requestId"], &gotID))
		assert.Equal(t, testRequestID, gotID)
	})

	t.Run("does not overwrite existing non-empty requestId", func(t *testing.T) {
		handler := requestIDInjectorMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(httpheader.ContentType, httpheader.ContentTypeApplicationJSON)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"requestId":"already-set","memberships":[]}`))
		}))

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, newRequest())

		var body map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

		var gotID string
		require.NoError(t, json.Unmarshal(body["requestId"], &gotID))
		assert.Equal(t, "already-set", gotID)
	})

	t.Run("overwrites empty requestId", func(t *testing.T) {
		handler := requestIDInjectorMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(httpheader.ContentType, httpheader.ContentTypeApplicationJSON)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"requestId":"","memberships":[]}`))
		}))

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, newRequest())

		var body map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

		var gotID string
		require.NoError(t, json.Unmarshal(body["requestId"], &gotID))
		assert.Equal(t, testRequestID, gotID)
	})

	t.Run("passes through non-JSON response unchanged", func(t *testing.T) {
		const htmlBody = "<html>ok</html>"
		handler := requestIDInjectorMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(httpheader.ContentType, "text/html")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(htmlBody))
		}))

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, newRequest())

		assert.Equal(t, htmlBody, rec.Body.String())
	})

	t.Run("passes through empty body unchanged", func(t *testing.T) {
		handler := requestIDInjectorMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(httpheader.ContentType, httpheader.ContentTypeApplicationJSON)
			w.WriteHeader(http.StatusNoContent)
		}))

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, newRequest())

		assert.Empty(t, rec.Body.String())
	})
}
