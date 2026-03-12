package main

import (
	"encoding/base64"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/klauspost/compress/zstd"
)

var zstdEncoder *zstd.Encoder
var zstdDecoder *zstd.Decoder

func initCompression() {
	ddpdict, err := os.ReadFile("ddpdict")
	if err != nil {
		slog.Warn("failed to read ddpdict file. Using default dictionary.", "err", err)
		zstdEncoder, _ = zstd.NewWriter(nil)
		zstdDecoder, _ = zstd.NewReader(nil)
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

func serve_compress(c *gin.Context) {
	code, exists := c.GetQuery("code")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No code parameter present"})
		return
	}

	encoded := zstdEncoder.EncodeAll([]byte(code), nil)
	encodedStr := base64.StdEncoding.EncodeToString(encoded)
	encodedStr = url.QueryEscape(encodedStr)
	c.JSON(http.StatusOK, gin.H{"compressed": encodedStr})
}

func serve_decompress(c *gin.Context) {
	encodedStr, exists := c.GetQuery("code")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No code parameter present"})
		return
	}

	encoded, err := base64.StdEncoding.DecodeString(encodedStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid base64 input"})
		return
	}

	decompressed, err := zstdDecoder.DecodeAll(encoded, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid compressed data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": string(decompressed)})
}
