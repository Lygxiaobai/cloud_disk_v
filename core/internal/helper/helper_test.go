package helper

import "testing"

func TestBuildMultipartRangesSingleChunk(t *testing.T) {
	ranges, err := buildMultipartRanges(2, multipartChunkSize)
	if err != nil {
		t.Fatalf("buildMultipartRanges returned error: %v", err)
	}
	if len(ranges) != 1 {
		t.Fatalf("expected 1 range, got %d", len(ranges))
	}
	if ranges[0].PartNumber != 1 || ranges[0].Start != 0 || ranges[0].EndExclusive != 2 {
		t.Fatalf("unexpected range: %+v", ranges[0])
	}
}

func TestBuildMultipartRangesSplitByChunkSize(t *testing.T) {
	ranges, err := buildMultipartRanges(11, 5)
	if err != nil {
		t.Fatalf("buildMultipartRanges returned error: %v", err)
	}
	if len(ranges) != 3 {
		t.Fatalf("expected 3 ranges, got %d", len(ranges))
	}

	expected := []multipartRange{
		{PartNumber: 1, Start: 0, EndExclusive: 5},
		{PartNumber: 2, Start: 5, EndExclusive: 10},
		{PartNumber: 3, Start: 10, EndExclusive: 11},
	}
	for i := range expected {
		if ranges[i] != expected[i] {
			t.Fatalf("unexpected range at index %d: got %+v want %+v", i, ranges[i], expected[i])
		}
	}
}

func TestBuildMultipartRangesRejectsEmptyFile(t *testing.T) {
	if _, err := buildMultipartRanges(0, multipartChunkSize); err == nil {
		t.Fatal("expected error for empty file size")
	}
}
