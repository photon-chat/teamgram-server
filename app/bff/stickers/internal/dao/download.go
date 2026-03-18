package dao

import (
	"context"
	"encoding/base64"
	"fmt"
	"path"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/teamgram/proto/mtproto"
	"github.com/teamgram/teamgram-server/app/service/media/media"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	downloadWorkers = 1  // sequential download to minimize memory
	downloadBatch   = 10 // process stickers in small batches to limit memory
)

// StickerDownloadInput holds the info needed to download one sticker file and upload it to DFS.
type StickerDownloadInput struct {
	BotFileId       string
	BotFileUniqueId string
	MimeType        string
	Attributes      []*mtproto.DocumentAttribute
	ThumbFileId     string // Bot API thumbnail file_id (optional)
	ThumbWidth      int32
	ThumbHeight     int32
}

// DownloadAndUploadStickerFiles downloads sticker files from Telegram Bot API and uploads them
// to DFS (MinIO) synchronously. Returns DFS-backed Documents in the same order as inputs.
// If any file fails, returns an error (caller should not cache partial results).
// For large sets, processes in batches to limit peak memory usage.
func (d *Dao) DownloadAndUploadStickerFiles(ctx context.Context, inputs []StickerDownloadInput) ([]*mtproto.Document, error) {
	if len(inputs) == 0 {
		return nil, nil
	}

	log := logx.WithContext(ctx)
	startAll := time.Now()
	log.Infof("DownloadAndUploadStickerFiles - start: %d stickers, workers=%d, batchSize=%d", len(inputs), downloadWorkers, downloadBatch)

	results := make([]*mtproto.Document, len(inputs))

	// Process in batches to limit peak memory
	for batchStart := 0; batchStart < len(inputs); batchStart += downloadBatch {
		batchEnd := batchStart + downloadBatch
		if batchEnd > len(inputs) {
			batchEnd = len(inputs)
		}

		batch := inputs[batchStart:batchEnd]
		log.Infof("DownloadAndUploadStickerFiles - batch [%d, %d) of %d", batchStart, batchEnd, len(inputs))

		var (
			mu       sync.Mutex
			firstErr error
		)

		sem := make(chan struct{}, downloadWorkers)
		var wg sync.WaitGroup

		for i := range batch {
			idx := batchStart + i
			input := batch[i]

			wg.Add(1)
			sem <- struct{}{} // per-request concurrency limit
			go func() {
				defer wg.Done()
				defer func() { <-sem }()
				defer func() {
					if r := recover(); r != nil {
						mu.Lock()
						if firstErr == nil {
							firstErr = fmt.Errorf("panic downloading sticker %d: %v", idx, r)
						}
						mu.Unlock()
						logx.WithContext(ctx).Errorf("downloadAndUploadOne panic: %v", r)
					}
				}()

				// Acquire global semaphore to cap total memory usage across all requests
				globalDownloadSem <- struct{}{}
				defer func() { <-globalDownloadSem }()

				doc, err := d.downloadAndUploadOne(ctx, &input)
				mu.Lock()
				defer mu.Unlock()
				if err != nil {
					if firstErr == nil {
						firstErr = fmt.Errorf("sticker[%d] (%s): %w", idx, input.BotFileId, err)
					}
				} else {
					results[idx] = doc
				}
			}()
		}

		wg.Wait()

		if firstErr != nil {
			log.Errorf("DownloadAndUploadStickerFiles - batch FAILED after %v: %v", time.Since(startAll), firstErr)
			return nil, firstErr
		}

		// Force GC between batches to reclaim memory from completed downloads
		if batchEnd < len(inputs) {
			runtime.GC()
			debug.FreeOSMemory()
		}
	}

	log.Infof("DownloadAndUploadStickerFiles - SUCCESS: %d stickers in %v", len(inputs), time.Since(startAll))
	return results, nil
}

// downloadAndUploadOne downloads a single sticker file from Bot API and uploads it via media service.
// Uses direct mode: file bytes are passed inline via gRPC, bypassing SSDB entirely.
func (d *Dao) downloadAndUploadOne(ctx context.Context, input *StickerDownloadInput) (*mtproto.Document, error) {
	log := logx.WithContext(ctx)
	start := time.Now()

	// 1. Get file path from Bot API
	fileInfo, err := d.BotAPI.GetFile(ctx, input.BotFileId)
	if err != nil {
		return nil, fmt.Errorf("GetFile: %w", err)
	}
	tGetFile := time.Since(start)

	// 2. Download the file bytes
	data, err := d.BotAPI.DownloadFile(ctx, fileInfo.FilePath)
	if err != nil {
		return nil, fmt.Errorf("DownloadFile: %w", err)
	}
	tDownload := time.Since(start)
	dataSize := len(data)

	// 3. Build InputMedia with file name (for extension detection)
	ext := path.Ext(fileInfo.FilePath)
	if ext == "" {
		ext = ".dat"
	}

	inputMedia := mtproto.MakeTLInputMediaUploadedDocument(&mtproto.InputMedia{
		File: mtproto.MakeTLInputFile(&mtproto.InputFile{
			Name: input.BotFileUniqueId + ext,
		}).To_InputFile(),
		MimeType:   input.MimeType,
		Attributes: input.Attributes,
	}).To_InputMedia()

	// 4. Download thumbnail if available
	var thumbData []byte
	if input.ThumbFileId != "" {
		thumbData, err = d.downloadThumbBytes(ctx, input.ThumbFileId)
		if err != nil {
			log.Infof("downloadAndUploadOne - thumb download failed (non-fatal): %v", err)
			thumbData = nil
		}
	}

	// 5. Upload via media service with inline file data (no SSDB)
	messageMedia, err := d.MediaClient.MediaUploadedDocumentMedia(ctx, &media.TLMediaUploadedDocumentMedia{
		OwnerId:   0,
		Media:     inputMedia,
		FileData:  data,
		ThumbData: thumbData,
	})

	// Release data early
	data = nil
	thumbData = nil

	if err != nil {
		return nil, fmt.Errorf("MediaUploadedDocumentMedia: %w", err)
	}

	dfsDoc := messageMedia.GetDocument()
	if dfsDoc == nil {
		return nil, fmt.Errorf("MediaUploadedDocumentMedia returned nil document")
	}

	log.Infof("downloadAndUploadOne - %s → doc %d (%d bytes) getFile=%v download=%v total=%v",
		input.BotFileUniqueId, dfsDoc.GetId(), dataSize,
		tGetFile, tDownload, time.Since(start))

	return dfsDoc, nil
}

// SerializeStickerDoc serializes a Document protobuf to base64 string for DB storage.
func SerializeStickerDoc(doc *mtproto.Document) (string, error) {
	data, err := proto.Marshal(doc)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// DeserializeStickerDoc deserializes a base64-encoded Document protobuf from DB.
func DeserializeStickerDoc(s string) (*mtproto.Document, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	doc := &mtproto.Document{}
	return doc, proto.Unmarshal(data, doc)
}

// downloadThumbBytes downloads a thumbnail from Bot API and returns the raw bytes.
func (d *Dao) downloadThumbBytes(ctx context.Context, thumbFileId string) ([]byte, error) {
	// 1. Get file path from Bot API
	fileInfo, err := d.BotAPI.GetFile(ctx, thumbFileId)
	if err != nil {
		return nil, fmt.Errorf("GetFile(thumb): %w", err)
	}

	// 2. Download the thumb bytes
	data, err := d.BotAPI.DownloadFile(ctx, fileInfo.FilePath)
	if err != nil {
		return nil, fmt.Errorf("DownloadFile(thumb): %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("thumb file is empty")
	}

	return data, nil
}
