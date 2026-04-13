package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/localinsights/portal/internal/pkg/database"
)

// responseWriter is a custom gin.ResponseWriter that captures the response
// status code and body so they can be stored for idempotency replay.
type responseWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Idempotency returns a middleware that ensures POST requests with an
// X-Idempotency-Key header are executed at most once within the given TTL.
//
// When a key is seen for the first time a placeholder row is inserted into
// the idempotency_keys table. After the handler completes the row is
// updated with the response code and body. Subsequent requests with the
// same key replay the stored response.
//
// POST requests without the header are rejected with 400.
func Idempotency(db *database.DB, ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only applies to POST requests.
		if c.Request.Method != http.MethodPost {
			c.Next()
			return
		}

		idempotencyKey := c.GetHeader("X-Idempotency-Key")
		if idempotencyKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "X-Idempotency-Key header is required for this endpoint",
			})
			c.Abort()
			return
		}

		// This middleware requires an authenticated user. All routes using
		// it must be behind RequireAuth. Auth routes are exempt by design.
		userID := GetUserID(c)
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "Authentication required for idempotent endpoints",
			})
			c.Abort()
			return
		}

		// Build a hash of (key + user_id + endpoint) to scope keys per
		// user AND endpoint, preventing cross-endpoint replay collisions.
		endpoint := c.Request.Method + " " + c.FullPath()
		raw := fmt.Sprintf("%s:%d:%s", idempotencyKey, userID, endpoint)
		hash := sha256.Sum256([]byte(raw))
		keyHash := hex.EncodeToString(hash[:])

		ctx := c.Request.Context()

		// Check for an existing entry (key_hash is the unique index).
		var responseCode int
		var responseBody string
		var expiresAt time.Time

		err := db.QueryRowContext(ctx,
			"SELECT response_code, response_body, expires_at FROM idempotency_keys WHERE key_hash = ?",
			keyHash,
		).Scan(&responseCode, &responseBody, &expiresAt)

		if err == nil {
			// Entry exists — check expiry.
			if time.Now().Before(expiresAt) {
				c.Data(responseCode, "application/json; charset=utf-8", []byte(responseBody))
				c.Abort()
				return
			}
			// Expired — delete and proceed as new.
			_, _ = db.ExecContext(ctx,
				"DELETE FROM idempotency_keys WHERE key_hash = ?", keyHash)
		} else if err != sql.ErrNoRows {
			slog.Error("idempotency key lookup failed", "error", err, "key_hash", keyHash)
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"code": http.StatusServiceUnavailable,
				"msg":  "Idempotency check unavailable, please retry",
			})
			c.Abort()
			return
		}

		// Insert placeholder row.
		expiresAt = time.Now().Add(ttl)
		_, insertErr := db.ExecContext(ctx,
			"INSERT INTO idempotency_keys (key_hash, user_id, endpoint, response_code, response_body, created_at, expires_at) VALUES (?, ?, ?, 0, '', NOW(3), ?)",
			keyHash, userID, endpoint, expiresAt,
		)
		if insertErr != nil {
			slog.Error("idempotency key insert failed", "error", insertErr, "key_hash", keyHash)
			// Possible duplicate insert race — try to replay.
			if replayErr := db.QueryRowContext(ctx,
				"SELECT response_code, response_body, expires_at FROM idempotency_keys WHERE key_hash = ?",
				keyHash,
			).Scan(&responseCode, &responseBody, &expiresAt); replayErr == nil && time.Now().Before(expiresAt) && responseCode != 0 {
				c.Data(responseCode, "application/json; charset=utf-8", []byte(responseBody))
				c.Abort()
				return
			}
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"code": http.StatusServiceUnavailable,
				"msg":  "Idempotency check unavailable, please retry",
			})
			c.Abort()
			return
		}

		// Wrap the response writer to capture output.
		w := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
			statusCode:     http.StatusOK,
		}
		c.Writer = w

		c.Next()

		// Update the row with the actual response.
		capturedStatus := w.statusCode
		capturedBody := w.body.String()

		updateCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, updateErr := db.ExecContext(updateCtx,
			"UPDATE idempotency_keys SET response_code = ?, response_body = ? WHERE key_hash = ?",
			capturedStatus, capturedBody, keyHash,
		)
		if updateErr != nil {
			slog.Error("idempotency key update failed", "error", updateErr, "key_hash", keyHash)
		}
	}
}
