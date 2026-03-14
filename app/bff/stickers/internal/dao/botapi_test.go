package dao

import (
	"context"
	"testing"
)

const testBotToken = "8784954709:AAF-N47mj6N_UKn3NbJ2HgPWmryOyFjsJI4"

// TestBotAPIGetStickerSet tests the real Bot API getStickerSet call
func TestBotAPIGetStickerSet(t *testing.T) {
	client := NewBotAPIClient(testBotToken)
	ctx := context.Background()

	set, err := client.GetStickerSet(ctx, "UtyaDuck")
	if err != nil {
		t.Fatalf("GetStickerSet failed: %v", err)
	}

	// Verify basic fields
	if set.Name != "UtyaDuck" {
		t.Errorf("expected name 'UtyaDuck', got '%s'", set.Name)
	}
	if set.Title == "" {
		t.Error("title is empty")
	}
	if set.StickerType == "" {
		t.Error("sticker_type is empty")
	}
	if len(set.Stickers) == 0 {
		t.Fatal("stickers array is empty")
	}

	t.Logf("Sticker set: name=%s, title=%s, type=%s, count=%d",
		set.Name, set.Title, set.StickerType, len(set.Stickers))

	// Verify first sticker
	s := set.Stickers[0]
	if s.FileId == "" {
		t.Error("first sticker file_id is empty")
	}
	if s.FileUniqueId == "" {
		t.Error("first sticker file_unique_id is empty")
	}
	if s.Width == 0 || s.Height == 0 {
		t.Errorf("first sticker dimensions invalid: %dx%d", s.Width, s.Height)
	}
	if s.Emoji == "" {
		t.Error("first sticker emoji is empty")
	}
	if s.Type == "" {
		t.Error("first sticker type is empty")
	}

	t.Logf("First sticker: file_id=%s..., emoji=%s, type=%s, size=%d, %dx%d, animated=%v, video=%v",
		s.FileId[:20], s.Emoji, s.Type, s.FileSize, s.Width, s.Height, s.IsAnimated, s.IsVideo)

	// Verify thumbnail on sticker
	if s.Thumbnail != nil {
		t.Logf("First sticker thumbnail: %dx%d, size=%d", s.Thumbnail.Width, s.Thumbnail.Height, s.Thumbnail.FileSize)
	}

	// Check all stickers have required fields
	for i, sticker := range set.Stickers {
		if sticker.FileId == "" {
			t.Errorf("sticker[%d] has empty file_id", i)
		}
		if sticker.FileUniqueId == "" {
			t.Errorf("sticker[%d] has empty file_unique_id", i)
		}
		if sticker.Emoji == "" {
			t.Errorf("sticker[%d] has empty emoji", i)
		}
	}
	t.Logf("All %d stickers have valid file_id, file_unique_id, and emoji", len(set.Stickers))
}

// TestBotAPIGetStickerSetNotFound tests error handling for non-existent sticker set
func TestBotAPIGetStickerSetNotFound(t *testing.T) {
	client := NewBotAPIClient(testBotToken)
	ctx := context.Background()

	_, err := client.GetStickerSet(ctx, "this_sticker_set_definitely_does_not_exist_12345")
	if err == nil {
		t.Fatal("expected error for non-existent sticker set, got nil")
	}
	t.Logf("Correctly returned error for non-existent set: %v", err)
}

// TestBotAPIGetFile tests the real Bot API getFile call
func TestBotAPIGetFile(t *testing.T) {
	client := NewBotAPIClient(testBotToken)
	ctx := context.Background()

	// First get a sticker set to get a valid file_id
	set, err := client.GetStickerSet(ctx, "UtyaDuck")
	if err != nil {
		t.Fatalf("GetStickerSet failed: %v", err)
	}
	if len(set.Stickers) == 0 {
		t.Fatal("no stickers to test with")
	}

	fileId := set.Stickers[0].FileId
	file, err := client.GetFile(ctx, fileId)
	if err != nil {
		t.Fatalf("GetFile failed: %v", err)
	}

	if file.FileId == "" {
		t.Error("file_id is empty")
	}
	if file.FilePath == "" {
		t.Error("file_path is empty")
	}
	if file.FileSize == 0 {
		t.Error("file_size is 0")
	}

	t.Logf("File: id=%s..., path=%s, size=%d", file.FileId[:20], file.FilePath, file.FileSize)
}

// TestBotAPIDownloadFile tests downloading a real sticker file
func TestBotAPIDownloadFile(t *testing.T) {
	client := NewBotAPIClient(testBotToken)
	ctx := context.Background()

	// Chain: getStickerSet -> getFile -> downloadFile
	set, err := client.GetStickerSet(ctx, "UtyaDuck")
	if err != nil {
		t.Fatalf("GetStickerSet failed: %v", err)
	}

	fileId := set.Stickers[0].FileId
	file, err := client.GetFile(ctx, fileId)
	if err != nil {
		t.Fatalf("GetFile failed: %v", err)
	}

	data, err := client.DownloadFile(ctx, file.FilePath)
	if err != nil {
		t.Fatalf("DownloadFile failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("downloaded file is empty")
	}

	// Verify file size matches
	if int64(len(data)) != file.FileSize {
		t.Errorf("download size mismatch: got %d bytes, API said %d bytes", len(data), file.FileSize)
	}

	t.Logf("Successfully downloaded sticker: %d bytes (path: %s)", len(data), file.FilePath)
}

// TestBotAPIStickerTypes tests that we can handle different sticker types
func TestBotAPIStickerTypes(t *testing.T) {
	client := NewBotAPIClient(testBotToken)
	ctx := context.Background()

	testSets := []struct {
		name       string
		expectType string
	}{
		{"UtyaDuck", "regular"},
	}

	for _, tc := range testSets {
		t.Run(tc.name, func(t *testing.T) {
			set, err := client.GetStickerSet(ctx, tc.name)
			if err != nil {
				t.Fatalf("GetStickerSet(%s) failed: %v", tc.name, err)
			}

			if set.StickerType != tc.expectType {
				t.Errorf("expected sticker_type '%s', got '%s'", tc.expectType, set.StickerType)
			}

			// Check sticker type field on individual stickers
			for i, s := range set.Stickers {
				if s.Type != tc.expectType {
					t.Errorf("sticker[%d].type = '%s', expected '%s'", i, s.Type, tc.expectType)
					break
				}
			}

			t.Logf("Set '%s': type=%s, count=%d, animated=%v, video=%v",
				set.Name, set.StickerType, len(set.Stickers),
				len(set.Stickers) > 0 && set.Stickers[0].IsAnimated,
				len(set.Stickers) > 0 && set.Stickers[0].IsVideo)
		})
	}
}
