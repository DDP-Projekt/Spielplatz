package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/klauspost/compress/zstd"
)

var zstdEncoder *zstd.Encoder
var zstdDecoder *zstd.Decoder
var shareLinksDB *sql.DB

type ShareLink struct {
	UUID           string
	CompressedCode []byte
	CreatedAt      time.Time
}

const createShareLinksTableSQL = `
CREATE TABLE IF NOT EXISTS share_links (
	uuid TEXT PRIMARY KEY,
	compressed_code BLOB NOT NULL,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`

func initCompression() {
	ddpdict, err := os.ReadFile("ddpdict")
	if err != nil {
		slog.Warn("failed to read ddpdict file. Using default dictionary.", "err", err)
		zstdEncoder, err = zstd.NewWriter(nil)
		if err != nil {
			fatal("failed to create zstd encoder", "err", err)
		}
		zstdDecoder, err = zstd.NewReader(nil)
		if err != nil {
			fatal("failed to create zstd decoder", "err", err)
		}
	} else {
		zstdEncoder, err = zstd.NewWriter(nil, zstd.WithEncoderDict(ddpdict))
		if err != nil {
			fatal("failed to create zstd encoder with dict", "err", err)
		}
		zstdDecoder, err = zstd.NewReader(nil, zstd.WithDecoderDicts(ddpdict))
		if err != nil {
			fatal("failed to create zstd decoder with dict", "err", err)
		}
	}
}

func initShareLinksStorage(path string) error {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return err
	}

	if _, err := db.Exec(createShareLinksTableSQL); err != nil {
		db.Close()
		return err
	}

	shareLinksDB = db
	return nil
}

func closeShareLinksStorage() {
	if shareLinksDB != nil {
		if err := shareLinksDB.Close(); err != nil {
			slog.Warn("failed to close share links database", "err", err)
		}
	}
}

func storeShareData(id string, compressedCode []byte) error {
	if shareLinksDB == nil {
		return fmt.Errorf("share links database is not initialized")
	}

	_, err := shareLinksDB.Exec(
		"INSERT INTO share_links (uuid, compressed_code) VALUES (?, ?)",
		id,
		compressedCode,
	)
	return err
}

func getShareData(id string) (ShareLink, error) {
	if shareLinksDB == nil {
		return ShareLink{}, fmt.Errorf("share links database is not initialized")
	}

	var link ShareLink
	err := shareLinksDB.QueryRow(
		"SELECT uuid, compressed_code, created_at FROM share_links WHERE uuid = ?",
		id,
	).Scan(&link.UUID, &link.CompressedCode, &link.CreatedAt)
	if err != nil {
		return ShareLink{}, err
	}

	return link, nil
}

type CreateShareCodeRequest struct {
	Code string `json:"code"`
}

func serve_create_share_code(c *gin.Context) {
	logger := getLogger(c)

	var req CreateShareCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("failed to bind json", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	encoded := zstdEncoder.EncodeAll([]byte(req.Code), nil)
	id := uuid.NewString()
	if err := storeShareData(id, encoded); err != nil {
		logger.Error("failed to store share data", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate share code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"share_code": id})
}

func serve_get_share_data(c *gin.Context) {
	shareID, exists := c.GetQuery("code")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No code parameter present"})
		return
	}

	link, err := getShareData(shareID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unknown share code"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load shared code"})
		return
	}

	decompressed, err := zstdDecoder.DecodeAll(link.CompressedCode, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid compressed data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": string(decompressed)})
}
