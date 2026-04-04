package rawsample

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeSampleID(t *testing.T) {
	got := NormalizeSampleID("  Hello, World!  ")
	if got != "hello-world" {
		t.Fatalf("expected hello-world, got %q", got)
	}
}

func TestPersistWritesSampleFilesAndMeta(t *testing.T) {
	root := t.TempDir()
	saved, err := Persist(PersistOptions{
		RootDir:  root,
		SampleID: "My Sample! 01",
		Source:   "unit-test",
		Request: map[string]any{
			"model":  "deepseek-chat",
			"stream": true,
			"messages": []any{
				map[string]any{"role": "user", "content": "广州天气"},
			},
		},
		Capture: CaptureSummary{
			Label:      "deepseek_completion",
			URL:        "https://chat.deepseek.com/api/v0/chat/completion",
			StatusCode: 200,
		},
		UpstreamBody: []byte("data: {\"v\":\"hello [reference:1]\"}\n\n" +
			"data: {\"v\":\"FINISHED\",\"p\":\"response/status\"}\n\n"),
		ProcessedBody:        []byte("data: {\"choices\":[{\"delta\":{\"content\":\"hello\"},\"index\":0}],\"created\":1,\"id\":\"id\",\"model\":\"m\",\"object\":\"chat.completion.chunk\"}\n\n"),
		ProcessedStatusCode:  200,
		ProcessedContentType: "text/event-stream",
	})
	if err != nil {
		t.Fatalf("Persist failed: %v", err)
	}

	if saved.SampleID != "my-sample-01" {
		t.Fatalf("expected normalized sample id, got %q", saved.SampleID)
	}
	if _, err := os.Stat(saved.Dir); err != nil {
		t.Fatalf("sample dir missing: %v", err)
	}
	if _, err := os.Stat(saved.UpstreamPath); err != nil {
		t.Fatalf("upstream file missing: %v", err)
	}
	if _, err := os.Stat(saved.ProcessedPath); err != nil {
		t.Fatalf("processed file missing: %v", err)
	}
	if saved.OutputPath != filepath.Join(saved.Dir, "openai.output.txt") {
		t.Fatalf("unexpected processed text path: %s", saved.OutputPath)
	}
	if _, err := os.Stat(saved.OutputPath); err != nil {
		t.Fatalf("processed text file missing: %v", err)
	}

	metaBytes, err := os.ReadFile(saved.MetaPath)
	if err != nil {
		t.Fatalf("read meta: %v", err)
	}
	var meta Meta
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		t.Fatalf("decode meta: %v", err)
	}
	if meta.SampleID != saved.SampleID {
		t.Fatalf("expected meta sample id %q, got %q", saved.SampleID, meta.SampleID)
	}
	if meta.Capture.ReferenceMarkerCount != 1 {
		t.Fatalf("expected one reference marker, got %+v", meta.Capture)
	}
	if meta.Capture.FinishedTokenCount != 1 {
		t.Fatalf("expected one finished token, got %+v", meta.Capture)
	}
	if meta.Processed.File != "openai.stream.sse" {
		t.Fatalf("expected stream processed file, got %+v", meta.Processed)
	}
	if meta.Processed.TextFile != "openai.output.txt" {
		t.Fatalf("expected text file metadata, got %+v", meta.Processed)
	}
	if meta.Processed.ResponseBytes == 0 {
		t.Fatalf("expected processed bytes to be recorded, got %+v", meta.Processed)
	}
	if !strings.HasSuffix(saved.ProcessedPath, filepath.Join(saved.SampleID, "openai.stream.sse")) {
		t.Fatalf("unexpected processed path: %s", saved.ProcessedPath)
	}
}
