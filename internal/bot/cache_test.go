package bot

import "testing"

func TestBoundedCacheEvictsOldestEntries(t *testing.T) {
	cache := newBoundedCache[int](2)
	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.Set("c", 3)

	if cache.Len() != 2 {
		t.Fatalf("expected cache len 2, got %d", cache.Len())
	}
	if _, ok := cache.Get("a"); ok {
		t.Fatal("expected oldest entry to be evicted")
	}
	if v, ok := cache.Get("b"); !ok || v != 2 {
		t.Fatalf("expected to keep b=2, got %v ok=%v", v, ok)
	}
	if v, ok := cache.Get("c"); !ok || v != 3 {
		t.Fatalf("expected to keep c=3, got %v ok=%v", v, ok)
	}
}
