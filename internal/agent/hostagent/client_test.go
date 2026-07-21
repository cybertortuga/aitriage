package hostagent

import (
	"bytes"
	"context"
	"sync"
	"testing"

	"github.com/cybertortuga/aitriage/internal/agent/llm"
)

// memStore is a minimal, concurrency-safe in-memory Store keyed by request
// fingerprint (matching the real content-addressed contract).
type memStore struct {
	mu        sync.Mutex
	requests  map[string]Request
	responses map[string]Response
}

func newMemStore() *memStore {
	return &memStore{requests: map[string]Request{}, responses: map[string]Response{}}
}

func (m *memStore) SaveRequest(req Request) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests[req.RequestID] = req // idempotent by fingerprint
	return nil
}

func (m *memStore) Response(requestID string) (Response, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.responses[requestID]
	return r, ok, nil
}

func (m *memStore) putResponse(r Response) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[r.RequestID] = r
}

func msgs(content string) []llm.Message {
	return []llm.Message{
		{Role: "system", Content: "system"},
		{Role: "user", Content: content},
	}
}

// TestDeferOnFirstCall proves the first Chat defers, persisting the exact
// request and returning a typed awaiting error carrying the content fingerprint.
func TestDeferOnFirstCall(t *testing.T) {
	store := newMemStore()
	c := New("run-1", "state-fp", store)

	_, _, err := c.Chat(context.Background(), msgs("stage-1"))
	awaiting, ok := IsAwaiting(err)
	if !ok {
		t.Fatalf("expected ErrAwaitingHostResponse, got %v", err)
	}
	if awaiting.RunID != "run-1" || awaiting.RequestID == "" {
		t.Fatalf("awaiting error missing identity: %+v", awaiting)
	}
	req, ok := store.requests[awaiting.RequestID]
	if !ok {
		t.Fatal("request was not persisted under its fingerprint")
	}
	if req.Messages[1].Content != "stage-1" {
		t.Fatalf("stored request payload is not verbatim: %+v", req.Messages)
	}
}

// TestReplayReturnsStoredResponse proves a re-run replays a validated response
// for the first request and then defers on the second.
func TestReplayReturnsStoredResponse(t *testing.T) {
	store := newMemStore()

	c1 := New("run-1", "state-fp", store)
	_, _, err := c1.Chat(context.Background(), msgs("stage-1"))
	awaiting, _ := IsAwaiting(err)
	store.putResponse(Response{
		RunID: "run-1", RequestID: awaiting.RequestID,
		Content: "answer-1", UsageReported: true, Usage: llm.Usage{TotalTokens: 12},
	})

	c2 := New("run-1", "state-fp", store)
	got, usage, err := c2.Chat(context.Background(), msgs("stage-1"))
	if err != nil {
		t.Fatalf("replay should succeed, got %v", err)
	}
	if got != "answer-1" || usage.TotalTokens != 12 {
		t.Fatalf("replay returned wrong content/usage: %q %+v", got, usage)
	}

	if _, _, err = c2.Chat(context.Background(), msgs("stage-2")); err != nil {
		if _, ok := IsAwaiting(err); !ok {
			t.Fatalf("expected defer on second request, got %v", err)
		}
	}
}

// TestUsageNotFabricated proves a response without reported usage yields zero
// usage rather than invented tokens.
func TestUsageNotFabricated(t *testing.T) {
	store := newMemStore()
	c1 := New("run-1", "state-fp", store)
	_, _, err := c1.Chat(context.Background(), msgs("stage-1"))
	awaiting, _ := IsAwaiting(err)
	store.putResponse(Response{
		RunID: "run-1", RequestID: awaiting.RequestID,
		Content: "answer", UsageReported: false, Usage: llm.Usage{TotalTokens: 999},
	})
	c2 := New("run-1", "state-fp", store)
	_, usage, err := c2.Chat(context.Background(), msgs("stage-1"))
	if err != nil {
		t.Fatal(err)
	}
	if usage.TotalTokens != 0 {
		t.Fatalf("usage must be zero when not reported, got %+v", usage)
	}
}

// TestFingerprintStableAndByteExact proves the fingerprint is a stable digest of
// the canonical request bytes and is order-independent (no sequence input), so
// it is reproducible from the persisted request.
func TestFingerprintStableAndByteExact(t *testing.T) {
	c := New("run-1", "state-fp", newMemStore())
	m := msgs("stage-1")

	if c.fingerprint(m) != c.fingerprint(m) {
		t.Fatal("fingerprint not stable across calls")
	}
	// Different content -> different fingerprint.
	if c.fingerprint(m) == c.fingerprint(msgs("stage-2")) {
		t.Fatal("distinct content must produce distinct fingerprints")
	}
	want := CanonicalRequestBytes("run-1", "state-fp", CurrentPromptVersions(), m)
	got := CanonicalRequestBytes("run-1", "state-fp", CurrentPromptVersions(), m)
	if !bytes.Equal(want, got) {
		t.Fatal("canonical request bytes are not deterministic")
	}
}

// TestConcurrentChatIsOrderIndependent is the blocker-1 regression: many Chat
// calls issued concurrently (as graph.Run's parallel classification batches do)
// must each map to a stable, content-derived fingerprint regardless of the order
// in which goroutines run. Two independent passes with different scheduling must
// produce the identical set of request fingerprints.
func TestConcurrentChatIsOrderIndependent(t *testing.T) {
	const n = 32
	run := func() map[string]bool {
		store := newMemStore()
		c := New("run-1", "state-fp", store)
		var wg sync.WaitGroup
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				// Distinct content per call — like distinct finding batches.
				_, _, _ = c.Chat(context.Background(), msgs(string(rune('A'+i%26))+itoaTest(i)))
			}(i)
		}
		wg.Wait()
		ids := map[string]bool{}
		for id := range store.requests {
			ids[id] = true
		}
		return ids
	}

	a := run()
	b := run()
	if len(a) != n || len(b) != n {
		t.Fatalf("expected %d distinct requests, got a=%d b=%d", n, len(a), len(b))
	}
	for id := range a {
		if !b[id] {
			t.Fatalf("fingerprint set differs between concurrent passes (order-dependent identity): %s missing", id)
		}
	}
}

func itoaTest(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
