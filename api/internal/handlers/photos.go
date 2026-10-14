package handlers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/png" // PNG uploads
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
)

const (
	// PhotoIdPathValue is the photo route's path value (the user id).
	PhotoIdPathValue = "photoUserId"
	photoMaxBytes    = 5 << 20
	photoSize        = 256
	photoURLTTL      = time.Hour
)

func photoPath(cfg *ApiConfig, userID int64) string {
	return filepath.Join(cfg.PhotoDir, fmt.Sprintf("%d.jpg", userID))
}

func photoSignature(cfg *ApiConfig, userID, expires int64) string {
	mac := hmac.New(sha256.New, []byte(cfg.Secret))
	fmt.Fprintf(mac, "photo:%d:%d", userID, expires)
	return hex.EncodeToString(mac.Sum(nil))
}

// PhotoURL is a user's photo address, signed and short-lived so images load
// without an auth header (and can't be fetched forever); "" without a photo.
// The version (`v`) changes with each upload, so caches don't serve the old
// picture.
func PhotoURL(cfg *ApiConfig, user database.User) string {
	return photoURL(cfg, user.ID, user.PhotoUpdatedAt)
}

func photoURL(cfg *ApiConfig, userID int64, updatedAt sql.NullTime) string {
	if !updatedAt.Valid {
		return ""
	}
	// Rounded so URLs (and browser caches) stay stable for a while.
	expires := time.Now().Add(photoURLTTL).Truncate(10 * time.Minute).Add(10 * time.Minute).Unix()
	return fmt.Sprintf("/v1/api/photos/%d.jpg?v=%d&exp=%d&sig=%s",
		userID, updatedAt.Time.Unix(), expires, photoSignature(cfg, userID, expires))
}

// OwnerPhotoURLs maps the association's owners to their account's photo
// (owners without an account, a photo or an enabled account are left out).
// For the staff lists: one query per response, not per owner.
func OwnerPhotoURLs(ctx context.Context, cfg *ApiConfig, associationID int64) map[int64]string {
	rows, err := cfg.Db.ListAssociationOwnerAccounts(ctx, associationID)
	photos := map[int64]string{}
	if err != nil {
		logging.Logger.Log(zap.WarnLevel, "Error reading owner photos", zap.Error(err))
		return photos
	}
	for _, r := range rows {
		if url := photoURL(cfg, r.UserID, r.PhotoUpdatedAt); url != "" && !r.DisabledAt.Valid {
			photos[r.OwnerID] = url
		}
	}
	return photos
}

// HandleGetPhoto: GET /photos/{userId}.jpg?exp&sig — no session needed; the
// signature and expiry are the access check.
func HandleGetPhoto(cfg *ApiConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		name := req.PathValue(PhotoIdPathValue)
		idStr := name
		if ext := filepath.Ext(name); ext == ".jpg" {
			idStr = name[:len(name)-len(ext)]
		}
		userID, ok := userIDParam(idStr)
		expires, err := strconv.ParseInt(req.URL.Query().Get("exp"), 10, 64)
		if !ok || err != nil || time.Now().Unix() > expires ||
			!hmac.Equal([]byte(req.URL.Query().Get("sig")), []byte(photoSignature(cfg, userID, expires))) {
			RespondWithError(rw, http.StatusForbidden, "Invalid or expired photo link")
			return
		}
		f, err := os.Open(photoPath(cfg, userID))
		if err != nil {
			RespondWithError(rw, http.StatusNotFound, "No photo")
			return
		}
		defer f.Close()
		rw.Header().Set("Content-Type", "image/jpeg")
		rw.Header().Set("Cache-Control", "private, max-age=3600")
		io.Copy(rw, f)
	}
}

// squarePhoto crops the centre square and scales it to size×size by
// averaging the source pixels under each target pixel (good for shrinking).
func squarePhoto(src image.Image, size int) *image.RGBA {
	b := src.Bounds()
	side := min(b.Dx(), b.Dy())
	x0 := b.Min.X + (b.Dx()-side)/2
	y0 := b.Min.Y + (b.Dy()-side)/2
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	scale := float64(side) / float64(size)
	for y := 0; y < size; y++ {
		sy0, sy1 := y0+int(float64(y)*scale), y0+max(int(float64(y+1)*scale), int(float64(y)*scale)+1)
		for x := 0; x < size; x++ {
			sx0, sx1 := x0+int(float64(x)*scale), x0+max(int(float64(x+1)*scale), int(float64(x)*scale)+1)
			var r, g, bl, n uint64
			for sy := sy0; sy < sy1; sy++ {
				for sx := sx0; sx < sx1; sx++ {
					cr, cg, cb, _ := src.At(sx, sy).RGBA()
					r, g, bl, n = r+uint64(cr), g+uint64(cg), bl+uint64(cb), n+1
				}
			}
			dst.Set(x, y, color.RGBA64{R: uint16(r / n), G: uint16(g / n), B: uint16(bl / n), A: 0xffff})
		}
	}
	return dst
}

// HandleUploadMyPhoto: PUT /me/photo (multipart "photo": JPEG or PNG, ≤ 5 MB)
// — stored as a 256 px square JPEG.
func HandleUploadMyPhoto(cfg *ApiConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		user, ok := currentUser(cfg, rw, req)
		if !ok {
			return
		}
		req.Body = http.MaxBytesReader(rw, req.Body, photoMaxBytes+1<<20)
		file, _, err := req.FormFile("photo")
		if err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Send the picture as the \"photo\" field (at most 5 MB)")
			return
		}
		defer file.Close()
		data, err := io.ReadAll(io.LimitReader(file, photoMaxBytes+1))
		if err != nil || len(data) > photoMaxBytes {
			RespondWithError(rw, http.StatusBadRequest, "The picture must be at most 5 MB")
			return
		}
		img, format, err := image.Decode(bytes.NewReader(data))
		if err != nil || (format != "jpeg" && format != "png") {
			RespondWithError(rw, http.StatusBadRequest, "The picture must be a JPEG or PNG image")
			return
		}
		var out bytes.Buffer
		if err := jpeg.Encode(&out, squarePhoto(img, photoSize), &jpeg.Options{Quality: 85}); err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error saving the picture")
			return
		}
		path := photoPath(cfg, user.ID)
		if err := os.MkdirAll(cfg.PhotoDir, 0o755); err == nil {
			tmp := path + ".tmp"
			if err = os.WriteFile(tmp, out.Bytes(), 0o644); err == nil {
				err = os.Rename(tmp, path)
			}
			if err == nil {
				err = cfg.Db.SetUserPhoto(req.Context(), database.SetUserPhotoParams{
					PhotoPath:      sql.NullString{String: filepath.Base(path), Valid: true},
					PhotoUpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
					ID:             user.ID,
				})
			}
		}
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error saving photo", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Error saving the picture")
			return
		}
		HandleGetMe(cfg)(rw, req)
	}
}

// HandleDeleteMyPhoto: DELETE /me/photo.
func HandleDeleteMyPhoto(cfg *ApiConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		user, ok := currentUser(cfg, rw, req)
		if !ok {
			return
		}
		if err := os.Remove(photoPath(cfg, user.ID)); err != nil && !errors.Is(err, os.ErrNotExist) {
			RespondWithError(rw, http.StatusInternalServerError, "Error removing the picture")
			return
		}
		if err := cfg.Db.SetUserPhoto(req.Context(), database.SetUserPhotoParams{ID: user.ID}); err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error removing the picture")
			return
		}
		HandleGetMe(cfg)(rw, req)
	}
}
