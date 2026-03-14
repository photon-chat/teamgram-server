package core

import (
	"math/rand"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/teamgram/proto/mtproto"
	"github.com/teamgram/teamgram-server/app/bff/stickers/internal/dal/dataobject"
	"github.com/teamgram/teamgram-server/app/bff/stickers/internal/dao"
)

// TestDocumentSerializationRoundtrip tests that Document protobuf survives serialize/deserialize
func TestDocumentSerializationRoundtrip(t *testing.T) {
	now := time.Now().Unix()

	// Build a realistic Document like fetchAndCacheStickerSet does
	sticker := dao.BotAPISticker{
		FileId:       "CAACAgIAAxkBAAIBdGF...",
		FileUniqueId: "AgADYQAD",
		FileSize:     5272,
		Width:        512,
		Height:       512,
		Emoji:        "😂",
		SetName:      "UtyaDuck",
		IsAnimated:   true,
		IsVideo:      false,
		Type:         "regular",
		Thumbnail: &dao.BotAPIPhotoSize{
			FileId:       "AAMCAgADFQABabU1uG4...",
			FileUniqueId: "AQADYQADthumb",
			FileSize:     5038,
			Width:        128,
			Height:       128,
		},
	}

	docId := int64(7654321)
	setId := int64(1234567)
	setAccessHash := rand.Int63()
	docAccessHash := generateAccessHash(sticker)
	mimeType := stickerMimeType(sticker)
	attributes := buildDocumentAttributes(sticker, setId, setAccessHash)
	thumbs := buildStickerThumbs(sticker)

	original := mtproto.MakeTLDocument(&mtproto.Document{
		Id:            docId,
		AccessHash:    docAccessHash,
		FileReference: []byte{},
		Date:          int32(now),
		MimeType:      mimeType,
		Size2_INT32:   int32(sticker.FileSize),
		Size2_INT64:   sticker.FileSize,
		Thumbs:        thumbs,
		VideoThumbs:   nil,
		DcId:          1,
		Attributes:    attributes,
	}).To_Document()

	// Serialize
	serialized, err := serializeDocument(original)
	if err != nil {
		t.Fatalf("serializeDocument failed: %v", err)
	}
	if serialized == "" {
		t.Fatal("serialized document is empty string")
	}
	t.Logf("Serialized document: %d chars (base64)", len(serialized))

	// Deserialize
	restored, err := deserializeDocument(serialized)
	if err != nil {
		t.Fatalf("deserializeDocument failed: %v", err)
	}

	// Verify all fields
	if restored.Id != original.Id {
		t.Errorf("Id: got %d, want %d", restored.Id, original.Id)
	}
	if restored.AccessHash != original.AccessHash {
		t.Errorf("AccessHash: got %d, want %d", restored.AccessHash, original.AccessHash)
	}
	if restored.MimeType != original.MimeType {
		t.Errorf("MimeType: got %s, want %s", restored.MimeType, original.MimeType)
	}
	if restored.DcId != original.DcId {
		t.Errorf("DcId: got %d, want %d", restored.DcId, original.DcId)
	}
	if restored.Date != original.Date {
		t.Errorf("Date: got %d, want %d", restored.Date, original.Date)
	}
	if restored.Size2_INT64 != original.Size2_INT64 {
		t.Errorf("Size2_INT64: got %d, want %d", restored.Size2_INT64, original.Size2_INT64)
	}
	if restored.Size2_INT32 != original.Size2_INT32 {
		t.Errorf("Size2_INT32: got %d, want %d", restored.Size2_INT32, original.Size2_INT32)
	}

	// Verify attributes count
	if len(restored.Attributes) != len(original.Attributes) {
		t.Fatalf("Attributes count: got %d, want %d", len(restored.Attributes), len(original.Attributes))
	}

	// Verify thumbs
	if len(restored.Thumbs) != len(original.Thumbs) {
		t.Fatalf("Thumbs count: got %d, want %d", len(restored.Thumbs), len(original.Thumbs))
	}
	if len(restored.Thumbs) > 0 {
		if restored.Thumbs[0].W != original.Thumbs[0].W || restored.Thumbs[0].H != original.Thumbs[0].H {
			t.Errorf("Thumb dimensions: got %dx%d, want %dx%d",
				restored.Thumbs[0].W, restored.Thumbs[0].H,
				original.Thumbs[0].W, original.Thumbs[0].H)
		}
	}

	// Verify proto-level equality
	if !proto.Equal(original, restored) {
		t.Error("proto.Equal returned false: original and restored documents differ")
	}

	t.Log("Document serialization roundtrip: PASS (all fields preserved)")
}

// TestDocumentSerializationNoThumbs tests Document without thumbnails
func TestDocumentSerializationNoThumbs(t *testing.T) {
	doc := mtproto.MakeTLDocument(&mtproto.Document{
		Id:            999,
		AccessHash:    888,
		FileReference: []byte{},
		Date:          int32(time.Now().Unix()),
		MimeType:      "video/webm",
		Size2_INT32:   1024,
		Size2_INT64:   1024,
		Thumbs:        nil,
		VideoThumbs:   nil,
		DcId:          1,
		Attributes:    []*mtproto.DocumentAttribute{},
	}).To_Document()

	serialized, err := serializeDocument(doc)
	if err != nil {
		t.Fatalf("serializeDocument failed: %v", err)
	}

	restored, err := deserializeDocument(serialized)
	if err != nil {
		t.Fatalf("deserializeDocument failed: %v", err)
	}

	if restored.Id != doc.Id {
		t.Errorf("Id: got %d, want %d", restored.Id, doc.Id)
	}
	if restored.MimeType != "video/webm" {
		t.Errorf("MimeType: got %s, want video/webm", restored.MimeType)
	}

	t.Log("Document without thumbs: PASS")
}

// TestStickerMimeType verifies MIME type determination
func TestStickerMimeType(t *testing.T) {
	tests := []struct {
		name     string
		sticker  dao.BotAPISticker
		wantMime string
		wantExt  string
	}{
		{
			name:     "animated TGS",
			sticker:  dao.BotAPISticker{IsAnimated: true, IsVideo: false},
			wantMime: "application/x-tgsticker",
			wantExt:  ".tgs",
		},
		{
			name:     "video WebM",
			sticker:  dao.BotAPISticker{IsAnimated: false, IsVideo: true},
			wantMime: "video/webm",
			wantExt:  ".webm",
		},
		{
			name:     "static WebP",
			sticker:  dao.BotAPISticker{IsAnimated: false, IsVideo: false},
			wantMime: "image/webp",
			wantExt:  ".webp",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotMime := stickerMimeType(tc.sticker)
			gotExt := stickerExt(tc.sticker)
			if gotMime != tc.wantMime {
				t.Errorf("stickerMimeType: got %s, want %s", gotMime, tc.wantMime)
			}
			if gotExt != tc.wantExt {
				t.Errorf("stickerExt: got %s, want %s", gotExt, tc.wantExt)
			}
		})
	}
}

// TestBuildDocumentAttributes verifies document attributes are correctly built
func TestBuildDocumentAttributes(t *testing.T) {
	sticker := dao.BotAPISticker{
		Emoji:        "🦆",
		Width:        512,
		Height:       512,
		FileUniqueId: "AgADtest",
		IsAnimated:   true,
	}

	setId := int64(100)
	setAccessHash := int64(200)

	attrs := buildDocumentAttributes(sticker, setId, setAccessHash)

	if len(attrs) != 3 {
		t.Fatalf("expected 3 attributes, got %d", len(attrs))
	}

	// Verify documentAttributeSticker
	stickerAttr := attrs[0]
	if stickerAttr.GetAlt() != "🦆" {
		t.Errorf("sticker alt: got '%s', want '🦆'", stickerAttr.GetAlt())
	}
	stickerSetRef := stickerAttr.GetStickerset()
	if stickerSetRef == nil {
		t.Fatal("stickerset reference is nil")
	}
	if stickerSetRef.GetId() != setId {
		t.Errorf("stickerset id: got %d, want %d", stickerSetRef.GetId(), setId)
	}
	if stickerSetRef.GetAccessHash() != setAccessHash {
		t.Errorf("stickerset access_hash: got %d, want %d", stickerSetRef.GetAccessHash(), setAccessHash)
	}

	// Verify documentAttributeImageSize
	imgAttr := attrs[1]
	if imgAttr.GetW() != 512 || imgAttr.GetH() != 512 {
		t.Errorf("image size: got %dx%d, want 512x512", imgAttr.GetW(), imgAttr.GetH())
	}

	// Verify documentAttributeFilename
	fileAttr := attrs[2]
	expectedName := "AgADtest.tgs"
	if fileAttr.GetFileName() != expectedName {
		t.Errorf("filename: got '%s', want '%s'", fileAttr.GetFileName(), expectedName)
	}

	t.Log("Document attributes: PASS")
}

// TestBuildStickerThumbs verifies thumbnail building
func TestBuildStickerThumbs(t *testing.T) {
	// With thumbnail
	sticker := dao.BotAPISticker{
		Thumbnail: &dao.BotAPIPhotoSize{
			Width:    128,
			Height:   128,
			FileSize: 4096,
		},
	}
	thumbs := buildStickerThumbs(sticker)
	if len(thumbs) != 1 {
		t.Fatalf("expected 1 thumb, got %d", len(thumbs))
	}
	if thumbs[0].W != 128 || thumbs[0].H != 128 {
		t.Errorf("thumb size: got %dx%d, want 128x128", thumbs[0].W, thumbs[0].H)
	}

	// Without thumbnail
	sticker2 := dao.BotAPISticker{Thumbnail: nil}
	thumbs2 := buildStickerThumbs(sticker2)
	if thumbs2 != nil {
		t.Errorf("expected nil thumbs, got %v", thumbs2)
	}

	t.Log("Sticker thumbs: PASS")
}

// TestBuildStickerPacks verifies emoji -> document_id grouping
func TestBuildStickerPacks(t *testing.T) {
	docDOs := []dataobject.StickerSetDocumentsDO{
		{DocumentId: 1, Emoji: "😂"},
		{DocumentId: 2, Emoji: "😭"},
		{DocumentId: 3, Emoji: "😂"}, // duplicate emoji
		{DocumentId: 4, Emoji: "🦆"},
		{DocumentId: 5, Emoji: ""}, // empty emoji
	}

	packs := buildStickerPacks(docDOs)

	// Should have 3 packs (😂, 😭, 🦆) — empty emoji skipped
	if len(packs) != 3 {
		t.Fatalf("expected 3 packs, got %d", len(packs))
	}

	// Build a map for easier checking
	packMap := make(map[string][]int64)
	for _, p := range packs {
		packMap[p.Emoticon] = p.Documents
	}

	// 😂 should have 2 documents
	if docs, ok := packMap["😂"]; !ok {
		t.Error("missing pack for 😂")
	} else if len(docs) != 2 || docs[0] != 1 || docs[1] != 3 {
		t.Errorf("😂 pack: got %v, want [1, 3]", docs)
	}

	// 😭 should have 1 document
	if docs, ok := packMap["😭"]; !ok {
		t.Error("missing pack for 😭")
	} else if len(docs) != 1 || docs[0] != 2 {
		t.Errorf("😭 pack: got %v, want [2]", docs)
	}

	// 🦆 should have 1 document
	if docs, ok := packMap["🦆"]; !ok {
		t.Error("missing pack for 🦆")
	} else if len(docs) != 1 || docs[0] != 4 {
		t.Errorf("🦆 pack: got %v, want [4]", docs)
	}

	t.Log("Sticker packs: PASS")
}

// TestMakeStickerSetFromDO verifies StickerSet protobuf construction from DO
func TestMakeStickerSetFromDO(t *testing.T) {
	setDO := &dataobject.StickerSetsDO{
		SetId:        12345,
		AccessHash:   67890,
		ShortName:    "TestSet",
		Title:        "Test Title",
		StickerCount: 10,
		Hash:         0,
		IsAnimated:   true,
		IsVideo:      false,
		IsMasks:      false,
		IsEmojis:     false,
		IsOfficial:   false,
		ThumbDocId:   0,
	}

	ss := makeStickerSetFromDO(setDO)

	if ss.Id != 12345 {
		t.Errorf("Id: got %d, want 12345", ss.Id)
	}
	if ss.AccessHash != 67890 {
		t.Errorf("AccessHash: got %d, want 67890", ss.AccessHash)
	}
	if ss.ShortName != "TestSet" {
		t.Errorf("ShortName: got %s, want TestSet", ss.ShortName)
	}
	if ss.Title != "Test Title" {
		t.Errorf("Title: got %s, want Test Title", ss.Title)
	}
	if ss.Count != 10 {
		t.Errorf("Count: got %d, want 10", ss.Count)
	}
	if !ss.Animated {
		t.Error("Animated should be true")
	}
	if ss.Videos {
		t.Error("Videos should be false")
	}
	if ss.ThumbDocumentId != nil {
		t.Error("ThumbDocumentId should be nil when ThumbDocId is 0")
	}

	// Test with thumb
	setDO.ThumbDocId = 999
	ss2 := makeStickerSetFromDO(setDO)
	if ss2.ThumbDocumentId == nil || ss2.ThumbDocumentId.Value != 999 {
		t.Errorf("ThumbDocumentId: got %v, want 999", ss2.ThumbDocumentId)
	}

	t.Log("StickerSet from DO: PASS")
}

// TestGenerateAccessHash verifies access hash generation for different sticker types
func TestGenerateAccessHash(t *testing.T) {
	animated := dao.BotAPISticker{IsAnimated: true}
	video := dao.BotAPISticker{IsVideo: true}
	static := dao.BotAPISticker{}

	hashA := generateAccessHash(animated)
	hashV := generateAccessHash(video)
	hashS := generateAccessHash(static)

	// Verify they encode the storageType in the upper 32 bits
	if (hashA >> 32) != 5 {
		t.Errorf("animated storageType: got %d, want 5", hashA>>32)
	}
	if (hashV >> 32) != 3 {
		t.Errorf("video storageType: got %d, want 3", hashV>>32)
	}
	if (hashS >> 32) != 1 {
		t.Errorf("static storageType: got %d, want 1", hashS>>32)
	}

	t.Logf("Access hashes: animated=0x%x, video=0x%x, static=0x%x", hashA, hashV, hashS)
}
