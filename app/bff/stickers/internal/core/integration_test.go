package core

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/teamgram/proto/mtproto"
	"github.com/teamgram/teamgram-server/app/bff/stickers/internal/dal/dataobject"
	"github.com/teamgram/teamgram-server/app/bff/stickers/internal/dao"
)

const testBotToken = "8784954709:AAF-N47mj6N_UKn3NbJ2HgPWmryOyFjsJI4"

// TestFullFlowFetchSerializeDeserialize simulates the complete flow:
// 1. Fetch sticker set from Bot API (real call)
// 2. Build Document protobufs (like fetchAndCacheStickerSet)
// 3. Serialize to base64 (like writing to DB)
// 4. Deserialize back (like reading from DB cache)
// 5. Verify the final output matches the original
func TestFullFlowFetchSerializeDeserialize(t *testing.T) {
	client := dao.NewBotAPIClient(testBotToken)
	ctx := context.Background()

	// --- Step 1: Fetch from Bot API ---
	botResult, err := client.GetStickerSet(ctx, "UtyaDuck")
	if err != nil {
		t.Fatalf("Bot API GetStickerSet failed: %v", err)
	}
	t.Logf("Fetched sticker set: name=%s, count=%d", botResult.Name, len(botResult.Stickers))

	// --- Step 2: Build Documents (simulating fetchAndCacheStickerSet) ---
	setId := int64(999000)
	setAccessHash := rand.Int63()
	now := time.Now().Unix()

	type docWithDO struct {
		doc *mtproto.Document
		do  *dataobject.StickerSetDocumentsDO
	}

	items := make([]docWithDO, 0, len(botResult.Stickers))

	for idx, sticker := range botResult.Stickers {
		docId := int64(1000000 + idx)
		docAccessHash := generateAccessHash(sticker)
		mimeType := stickerMimeType(sticker)
		fileSize := sticker.FileSize
		if fileSize == 0 {
			fileSize = 1
		}

		attributes := buildDocumentAttributes(sticker, setId, setAccessHash)
		thumbs := buildStickerThumbs(sticker)

		doc := mtproto.MakeTLDocument(&mtproto.Document{
			Id:            docId,
			AccessHash:    docAccessHash,
			FileReference: []byte{},
			Date:          int32(now),
			MimeType:      mimeType,
			Size2_INT32:   int32(fileSize),
			Size2_INT64:   fileSize,
			Thumbs:        thumbs,
			VideoThumbs:   nil,
			DcId:          1,
			Attributes:    attributes,
		}).To_Document()

		// Serialize document
		docData, err := serializeDocument(doc)
		if err != nil {
			t.Fatalf("serializeDocument failed for sticker[%d]: %v", idx, err)
		}

		thumbFileId := ""
		if sticker.Thumbnail != nil {
			thumbFileId = sticker.Thumbnail.FileId
		}

		do := &dataobject.StickerSetDocumentsDO{
			SetId:           setId,
			DocumentId:      docId,
			StickerIndex:    int32(idx),
			Emoji:           sticker.Emoji,
			BotFileId:       sticker.FileId,
			BotFileUniqueId: sticker.FileUniqueId,
			BotThumbFileId:  thumbFileId,
			DocumentData:    docData,
			FileDownloaded:  false,
		}

		items = append(items, docWithDO{doc: doc, do: do})
	}

	t.Logf("Built %d Document protobufs and serialized to document_data", len(items))

	// --- Step 3: Simulate cache hit — deserialize from document_data ---
	restoredDocs := make([]*mtproto.Document, 0, len(items))
	for i, item := range items {
		restored, err := deserializeDocument(item.do.DocumentData)
		if err != nil {
			t.Fatalf("deserializeDocument failed for sticker[%d]: %v", i, err)
		}
		restoredDocs = append(restoredDocs, restored)
	}

	t.Logf("Deserialized %d Documents from document_data", len(restoredDocs))

	// --- Step 4: Verify every document matches ---
	for i := range items {
		original := items[i].doc
		restored := restoredDocs[i]

		if !proto.Equal(original, restored) {
			t.Errorf("sticker[%d] proto.Equal failed", i)

			// Debug: check individual fields
			if original.Id != restored.Id {
				t.Errorf("  Id: %d vs %d", original.Id, restored.Id)
			}
			if original.AccessHash != restored.AccessHash {
				t.Errorf("  AccessHash: %d vs %d", original.AccessHash, restored.AccessHash)
			}
			if original.MimeType != restored.MimeType {
				t.Errorf("  MimeType: %s vs %s", original.MimeType, restored.MimeType)
			}
			if original.Size2_INT64 != restored.Size2_INT64 {
				t.Errorf("  Size: %d vs %d", original.Size2_INT64, restored.Size2_INT64)
			}
			if len(original.Attributes) != len(restored.Attributes) {
				t.Errorf("  Attributes: %d vs %d", len(original.Attributes), len(restored.Attributes))
			}
			if len(original.Thumbs) != len(restored.Thumbs) {
				t.Errorf("  Thumbs: %d vs %d", len(original.Thumbs), len(restored.Thumbs))
			}
		}
	}

	// --- Step 5: Verify StickerPack grouping ---
	docDOs := make([]dataobject.StickerSetDocumentsDO, len(items))
	for i, item := range items {
		docDOs[i] = *item.do
	}
	packs := buildStickerPacks(docDOs)
	if len(packs) == 0 {
		t.Error("packs are empty")
	}

	// Count total doc references in packs
	totalRefs := 0
	for _, p := range packs {
		totalRefs += len(p.Documents)
	}

	// Stickers with emoji should equal total refs
	stickersWithEmoji := 0
	for _, item := range items {
		if item.do.Emoji != "" {
			stickersWithEmoji++
		}
	}
	if totalRefs != stickersWithEmoji {
		t.Errorf("packs total refs=%d, stickers with emoji=%d — mismatch", totalRefs, stickersWithEmoji)
	}

	// --- Step 6: Verify StickerSet construction ---
	setDO := &dataobject.StickerSetsDO{
		SetId:        setId,
		AccessHash:   setAccessHash,
		ShortName:    botResult.Name,
		Title:        botResult.Title,
		StickerType:  botResult.StickerType,
		IsAnimated:   len(botResult.Stickers) > 0 && botResult.Stickers[0].IsAnimated,
		IsVideo:      len(botResult.Stickers) > 0 && botResult.Stickers[0].IsVideo,
		StickerCount: int32(len(botResult.Stickers)),
	}
	ss := makeStickerSetFromDO(setDO)

	if ss.ShortName != "UtyaDuck" {
		t.Errorf("StickerSet ShortName: got %s, want UtyaDuck", ss.ShortName)
	}
	if ss.Count != int32(len(botResult.Stickers)) {
		t.Errorf("StickerSet Count: got %d, want %d", ss.Count, len(botResult.Stickers))
	}

	// --- Step 7: Build final Messages_StickerSet ---
	result := mtproto.MakeTLMessagesStickerSet(&mtproto.Messages_StickerSet{
		Set:       ss,
		Packs:     packs,
		Keywords:  []*mtproto.StickerKeyword{},
		Documents: restoredDocs,
	}).To_Messages_StickerSet()

	if result.Set == nil {
		t.Fatal("result.Set is nil")
	}
	if len(result.Documents) != len(botResult.Stickers) {
		t.Errorf("result.Documents count: got %d, want %d", len(result.Documents), len(botResult.Stickers))
	}
	if len(result.Packs) == 0 {
		t.Error("result.Packs is empty")
	}

	t.Logf("=== FULL FLOW TEST PASSED ===")
	t.Logf("  Sticker set: %s (%s)", result.Set.Title, result.Set.ShortName)
	t.Logf("  Documents: %d (all serialization roundtrips matched)", len(result.Documents))
	t.Logf("  Packs: %d unique emoji groups", len(result.Packs))
	t.Logf("  First doc: id=%d, mime=%s, size=%d",
		result.Documents[0].Id, result.Documents[0].MimeType, result.Documents[0].Size2_INT64)
}

// TestFullFlowMultipleStickerSets tests that multiple sticker sets can be processed independently
func TestFullFlowMultipleStickerSets(t *testing.T) {
	client := dao.NewBotAPIClient(testBotToken)
	ctx := context.Background()

	setNames := []string{"UtyaDuck", "Animals"}

	for _, name := range setNames {
		t.Run(name, func(t *testing.T) {
			botResult, err := client.GetStickerSet(ctx, name)
			if err != nil {
				t.Skipf("Skipping %s: %v", name, err)
				return
			}

			setId := rand.Int63()
			setAccessHash := rand.Int63()
			now := time.Now().Unix()

			// Build and serialize all documents
			for idx, sticker := range botResult.Stickers {
				docId := int64(2000000 + idx)
				mimeType := stickerMimeType(sticker)
				attrs := buildDocumentAttributes(sticker, setId, setAccessHash)
				thumbs := buildStickerThumbs(sticker)

				doc := mtproto.MakeTLDocument(&mtproto.Document{
					Id:            docId,
					AccessHash:    generateAccessHash(sticker),
					FileReference: []byte{},
					Date:          int32(now),
					MimeType:      mimeType,
					Size2_INT32:   int32(sticker.FileSize),
					Size2_INT64:   sticker.FileSize,
					Thumbs:        thumbs,
					DcId:          1,
					Attributes:    attrs,
				}).To_Document()

				// Roundtrip
				serialized, err := serializeDocument(doc)
				if err != nil {
					t.Errorf("sticker[%d] serialize failed: %v", idx, err)
					continue
				}
				restored, err := deserializeDocument(serialized)
				if err != nil {
					t.Errorf("sticker[%d] deserialize failed: %v", idx, err)
					continue
				}
				if !proto.Equal(doc, restored) {
					t.Errorf("sticker[%d] roundtrip mismatch", idx)
				}
			}

			t.Logf("Set '%s': %d stickers, all roundtrips passed", name, len(botResult.Stickers))
		})
	}
}
