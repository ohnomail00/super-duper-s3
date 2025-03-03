package gateway

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"sort"

	"github.com/ohnomail00/super-duper-s3/engine"
	"github.com/ohnomail00/super-duper-s3/engine/hash"
	"github.com/ohnomail00/super-duper-s3/http/clients"
	"golang.org/x/sync/errgroup"
)

// Uploader handles the uploading of file parts.
type Uploader struct {
	hashRing  *hash.Ring
	partCount int
	f         clients.Factory
}

// NewUploader creates a new Uploader instance.
func NewUploader(hr *hash.Ring, partCount int, clientFactory clients.Factory) (*Uploader, error) {
	return &Uploader{
		hashRing:  hr,
		partCount: partCount,
		f:         clientFactory,
	}, nil
}

// uploadPart uploads a single file part to the designated server.
func (u *Uploader) uploadPart(ctx context.Context, bucket, object string, server engine.Server, partIndex int, r io.Reader) error {
	// Create a new HTTP client for the storage server.
	client := u.f.New(server.Address)

	// Use the HTTP client to upload the part.
	err := client.UploadPart(ctx, bucket, object, partIndex, r)
	if err != nil {
		return fmt.Errorf("failed to upload part %d: %w", partIndex, err)
	}

	slog.Debug(fmt.Sprintf("Part %d successfully uploaded to server %s via HTTP", partIndex, server.Address))
	return nil
}

// Do uploads file parts concurrently with improved error handling using errgroup.
// If an error occurs, all parallel operations are cancelled and the error is returned.
func (u *Uploader) Do(ctx context.Context, bucket, object string, reader io.ReadCloser, fileSize int64) (engine.FileUploadPlan, error) {
	defer reader.Close()

	partSize := fileSize / int64(u.partCount)
	var offset int64

	eg, ctx := errgroup.WithContext(ctx)
	partPlanCh := make(chan engine.PartPlan, u.partCount)

	// Acquire a read lock to prevent changes (adding server) to the hash ring while reading from it.
	u.hashRing.RLock()
	defer u.hashRing.RUnlock()

	// Read data for each part sequentially and launch goroutines for uploading.
	for i := 0; i < u.partCount; i++ {
		currentPartSize := partSize
		if i == u.partCount-1 {
			currentPartSize = fileSize - offset
		}

		// Read part data into a buffer.
		partData := make([]byte, currentPartSize)
		if _, err := io.ReadFull(reader, partData); err != nil {
			return engine.FileUploadPlan{}, fmt.Errorf("failed to read part %d: %w", i, err)
		}

		// Capture variables for the goroutine.
		partIndex := i
		off := offset
		data := make([]byte, len(partData))
		copy(data, partData)
		key := hash.GeneratePartKey(partIndex, int(off))
		server := u.hashRing.GetNode(key)

		eg.Go(func() error {
			// Create a new reader for the part data.
			partReader := bytes.NewReader(data)
			if err := u.uploadPart(ctx, bucket, object, server, partIndex, partReader); err != nil {
				return err
			}
			// Send information about the successfully uploaded part.
			partPlanCh <- engine.PartPlan{
				Index:  partIndex,
				Server: server,
				Offset: off,
				Length: currentPartSize,
			}
			return nil
		})

		offset += currentPartSize
	}

	// Wait for all goroutines to finish.
	if err := eg.Wait(); err != nil {
		close(partPlanCh)
		return engine.FileUploadPlan{}, err
	}
	close(partPlanCh)

	// Collect results from the channel.
	var plan engine.FileUploadPlan
	for partPlan := range partPlanCh {
		plan.Parts = append(plan.Parts, partPlan)
	}

	// Sort parts by index.
	sort.Slice(plan.Parts, func(i, j int) bool { return plan.Parts[i].Index < plan.Parts[j].Index })

	return plan, nil
}
