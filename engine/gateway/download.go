package gateway

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sort"

	"github.com/ohnomail00/super-duper-s3/engine"
	"github.com/ohnomail00/super-duper-s3/http/clients"
	"golang.org/x/sync/errgroup"
)

// Downloader handles downloading of file parts.
type Downloader struct {
	f clients.Factory
}

// NewDownloader creates a new Downloader instance.
func NewDownloader(f clients.Factory) *Downloader {
	return &Downloader{f: f}
}

// Do streams file parts directly to the provided writer without buffering the entire file in memory.
// This version uses errgroup to perform parallel downloads, cancelling all operations upon the first error.
func (d *Downloader) Do(ctx context.Context, bucket, object string, plan engine.FileUploadPlan, w io.Writer) error {
	// Sort parts by index for sequential assembly.
	sort.Slice(plan.Parts, func(i, j int) bool { return plan.Parts[i].Index < plan.Parts[j].Index })

	// Structure to store the result of a part download.
	type partResult struct {
		index  int
		reader io.ReadCloser
	}

	// Prepare a slice to store results.
	results := make([]*partResult, len(plan.Parts))

	eg, ctx := errgroup.WithContext(ctx)

	// Launch parallel downloads for each part.
	for i, part := range plan.Parts {
		idx := i
		p := part
		eg.Go(func() error {
			client := d.f.New(p.Server.Address)
			rc, err := client.DownloadPart(ctx, bucket, object, p.Index)
			if err != nil {
				return fmt.Errorf("failed to download part %d: %w", p.Index, err)
			}
			results[idx] = &partResult{
				index:  p.Index,
				reader: rc,
			}
			return nil
		})
	}

	// Wait for all goroutines to complete.
	if err := eg.Wait(); err != nil {
		// In case of error, close all already open readers.
		for _, res := range results {
			if res != nil && res.reader != nil {
				res.reader.Close()
			}
		}
		return err
	}

	// If the order in the slice does not match the part indices, sort them.
	sort.Slice(results, func(i, j int) bool {
		return results[i].index < results[j].index
	})

	// Create a MultiReader to sequentially stream all parts.
	readers := make([]io.Reader, len(results))
	for i, res := range results {
		readers[i] = res.reader
	}
	multiReader := io.MultiReader(readers...)

	// Copy data directly from the MultiReader to the output writer.
	_, err := io.Copy(w, multiReader)

	// Close all readers.
	for _, res := range results {
		res.reader.Close()
	}

	if err != nil {
		slog.Error(fmt.Sprintf("failed to stream file: %v", err))
		return fmt.Errorf("failed to stream file: %w", err)
	}
	return nil
}
