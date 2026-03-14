package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// BotAPIClient is a lightweight HTTP client for Telegram Bot API
type BotAPIClient struct {
	token  string
	client *http.Client
}

func NewBotAPIClient(token string) *BotAPIClient {
	return &BotAPIClient{
		token: token,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Bot API response types

type BotAPIStickerSetResponse struct {
	Ok          bool              `json:"ok"`
	Result      *BotAPIStickerSet `json:"result"`
	Description string            `json:"description"`
}

type BotAPIStickerSet struct {
	Name        string           `json:"name"`
	Title       string           `json:"title"`
	StickerType string           `json:"sticker_type"`
	Stickers    []BotAPISticker  `json:"stickers"`
	Thumbnail   *BotAPIPhotoSize `json:"thumbnail,omitempty"`
}

type BotAPISticker struct {
	FileId           string           `json:"file_id"`
	FileUniqueId     string           `json:"file_unique_id"`
	FileSize         int64            `json:"file_size"`
	Width            int32            `json:"width"`
	Height           int32            `json:"height"`
	Emoji            string           `json:"emoji"`
	SetName          string           `json:"set_name"`
	IsAnimated       bool             `json:"is_animated"`
	IsVideo          bool             `json:"is_video"`
	Type             string           `json:"type"`
	Thumbnail        *BotAPIPhotoSize `json:"thumbnail,omitempty"`
	PremiumAnimation *BotAPIFile      `json:"premium_animation,omitempty"`
}

type BotAPIPhotoSize struct {
	FileId       string `json:"file_id"`
	FileUniqueId string `json:"file_unique_id"`
	FileSize     int64  `json:"file_size"`
	Width        int32  `json:"width"`
	Height       int32  `json:"height"`
}

type BotAPIFileResponse struct {
	Ok          bool        `json:"ok"`
	Result      *BotAPIFile `json:"result"`
	Description string      `json:"description"`
}

type BotAPIFile struct {
	FileId       string `json:"file_id"`
	FileUniqueId string `json:"file_unique_id"`
	FileSize     int64  `json:"file_size"`
	FilePath     string `json:"file_path"`
}

// GetStickerSet calls the Bot API getStickerSet method
func (b *BotAPIClient) GetStickerSet(ctx context.Context, name string) (*BotAPIStickerSet, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getStickerSet?name=%s",
		b.token, url.QueryEscape(name))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("botapi: create request error: %w", err)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("botapi: getStickerSet request error: %w", err)
	}
	defer resp.Body.Close()

	var result BotAPIStickerSetResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("botapi: decode getStickerSet response error: %w", err)
	}

	if !result.Ok {
		logx.WithContext(ctx).Errorf("botapi: getStickerSet failed: %s", result.Description)
		return nil, fmt.Errorf("botapi: getStickerSet failed: %s", result.Description)
	}

	return result.Result, nil
}

// GetFile calls the Bot API getFile method
func (b *BotAPIClient) GetFile(ctx context.Context, fileId string) (*BotAPIFile, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getFile?file_id=%s",
		b.token, url.QueryEscape(fileId))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("botapi: create request error: %w", err)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("botapi: getFile request error: %w", err)
	}
	defer resp.Body.Close()

	var result BotAPIFileResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("botapi: decode getFile response error: %w", err)
	}

	if !result.Ok {
		return nil, fmt.Errorf("botapi: getFile failed: %s", result.Description)
	}

	return result.Result, nil
}

// DownloadFile downloads a file from the Bot API file server
func (b *BotAPIClient) DownloadFile(ctx context.Context, filePath string) ([]byte, error) {
	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", b.token, filePath)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return nil, fmt.Errorf("botapi: create download request error: %w", err)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("botapi: download file error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("botapi: download file returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("botapi: read file body error: %w", err)
	}

	return data, nil
}
