package success_cases

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ohnomail00/super-duper-s3/tests/utils"
)

func TestAddServerIntegration(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	tsStorage1, _ := utils.StartStorageServerWithStorageDir(t, dir1)
	tsGateway, serviceGateway1 := utils.StartTestGatewayServer(t, []string{tsStorage1.URL})

	time.Sleep(100 * time.Millisecond)

	tsStorage2, _ := utils.StartStorageServerWithStorageDir(t, dir2)

	addURL := fmt.Sprintf("%s/server", tsGateway.URL)
	body := fmt.Sprintf(`{"addr": "%s"}`, tsStorage2.URL)
	reqAdd, err := http.NewRequest("POST", addURL, bytes.NewReader([]byte(body)))
	if err != nil {
		t.Fatalf("Failed to create request to add server: %v", err)
	}
	respAdd, err := http.DefaultClient.Do(reqAdd)
	if err != nil {
		t.Fatalf("Error calling server add endpoint: %v", err)
	}
	if respAdd.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(respAdd.Body)
		t.Fatalf("Expected OK status when adding server, got %d: %s", respAdd.StatusCode, string(body))
	}
	respAdd.Body.Close()

	bucket := "bucket"
	objectName := "object.txt"
	originalContent := "Hello, test put/get"
	uploadURL := fmt.Sprintf("%s/%s/%s", tsGateway.URL, bucket, objectName)
	reqUpload, err := http.NewRequest("PUT", uploadURL, strings.NewReader(originalContent))
	if err != nil {
		t.Fatalf("Failed to create PUT request: %v", err)
	}

	respUpload, err := http.DefaultClient.Do(reqUpload)
	if err != nil {
		t.Fatalf("Error executing PUT request: %v", err)
	}
	if respUpload.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(respUpload.Body)
		t.Fatalf("PUT request failed with status %d: %s", respUpload.StatusCode, string(body))
	}
	respUpload.Body.Close()

	downloadURL := fmt.Sprintf("%s/%s/%s", tsGateway.URL, bucket, objectName)
	respDownload, err := http.Get(downloadURL)
	if err != nil {
		t.Fatalf("Error executing GET request: %v", err)
	}
	defer respDownload.Body.Close()
	if respDownload.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(respDownload.Body)
		t.Fatalf("GET request failed with status %d: %s", respDownload.StatusCode, string(body))
	}
	downloadedContent, err := io.ReadAll(respDownload.Body)
	if err != nil {
		t.Fatalf("Failed to read GET response body: %v", err)
	}
	if !bytes.Equal(downloadedContent, []byte(originalContent)) {
		t.Fatalf("Downloaded file content does not match: expected %q, got %q", originalContent, string(downloadedContent))
	}

	time.Sleep(200 * time.Millisecond)

	expectedParts := serviceGateway1.Cfg.PartCount

	partsDir1 := countPartFiles(dir1)
	partsDir2 := countPartFiles(dir2)
	totalParts := partsDir1 + partsDir2

	if totalParts != expectedParts {
		t.Fatalf("Incorrect total number of part files: expected %d, got %d (dir1: %d, dir2: %d)",
			expectedParts, totalParts, partsDir1, partsDir2)
	}
	if partsDir1 == 0 || partsDir2 == 0 {
		t.Fatalf("Part files were not distributed among servers (dir1: %d, dir2: %d)", partsDir1, partsDir2)
	}

	t.Logf("Success: total number of part files %d (dir1: %d, dir2: %d)", totalParts, partsDir1, partsDir2)
}

func countPartFiles(root string) int {
	count := 0
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasPrefix(d.Name(), "part_") {
			count++
		}
		return nil
	})
	return count
}
