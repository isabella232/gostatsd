package statsd

import (
	"context"
	"sync"

	"github.com/atlassian/gostatsd"
)

type capturingHandler struct {
	m  []*gostatsd.Metric
	mm []*gostatsd.MetricMap
	e  []*gostatsd.Event
}

func (tch *capturingHandler) EstimatedTags() int {
	return 0
}

func (tch *capturingHandler) DispatchMetrics(ctx context.Context, metrics []*gostatsd.Metric) {
	tch.m = append(tch.m, metrics...)
}

func (tch *capturingHandler) DispatchMetricMap(ctx context.Context, metrics *gostatsd.MetricMap) {
	tch.mm = append(tch.mm, metrics)
}

func (tch *capturingHandler) DispatchEvent(ctx context.Context, e *gostatsd.Event) {
	tch.e = append(tch.e, e)
}

func (tch *capturingHandler) WaitForEvents() {
}

type nopHandler struct{}

func (nh *nopHandler) EstimatedTags() int {
	return 0
}

func (nh *nopHandler) DispatchMetrics(ctx context.Context, m []*gostatsd.Metric) {
}

func (nh *nopHandler) DispatchMetricMap(ctx context.Context, mm *gostatsd.MetricMap) {
}

func (nh *nopHandler) DispatchEvent(ctx context.Context, e *gostatsd.Event) {
}

func (nh *nopHandler) WaitForEvents() {
}

type expectingHandler struct {
	countingHandler

	wgMetrics    sync.WaitGroup
	wgMetricMaps sync.WaitGroup
	wgEvents     sync.WaitGroup
}

func (e *expectingHandler) DispatchMetrics(ctx context.Context, m []*gostatsd.Metric) {
	e.countingHandler.DispatchMetrics(ctx, m)
	e.wgMetrics.Add(-len(m))
}

func (e *expectingHandler) DispatchMetricMap(ctx context.Context, mm *gostatsd.MetricMap) {
	e.countingHandler.DispatchMetricMap(ctx, mm)
	e.wgMetricMaps.Done()
}

func (e *expectingHandler) DispatchEvent(ctx context.Context, event *gostatsd.Event) {
	e.countingHandler.DispatchEvent(ctx, event)
	e.wgEvents.Done()
}

func (e *expectingHandler) Expect(ms, mms, es int) {
	e.wgMetrics.Add(ms)
	e.wgMetricMaps.Add(mms)
	e.wgEvents.Add(es)
}

func (e *expectingHandler) WaitAll() {
	e.wgMetrics.Wait()
	e.wgMetricMaps.Wait()
	e.wgEvents.Wait()
}

type countingHandler struct {
	mu      sync.Mutex
	metrics []*gostatsd.Metric
	events  gostatsd.Events
}

func (ch *countingHandler) Metrics() []*gostatsd.Metric {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	result := make([]*gostatsd.Metric, len(ch.metrics))
	copy(result, ch.metrics)
	return result
}

func (ch *countingHandler) Events() gostatsd.Events {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	result := make(gostatsd.Events, len(ch.events))
	copy(result, ch.events)
	return result
}

func (ch *countingHandler) EstimatedTags() int {
	return 0
}

// Wrapper until we can remove it
func (ch *countingHandler) DispatchMetrics(ctx context.Context, metrics []*gostatsd.Metric) {
	ch.dispatchMetrics(ctx, metrics)
}

func (ch *countingHandler) dispatchMetrics(ctx context.Context, metrics []*gostatsd.Metric) {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	for _, m := range metrics {
		m.DoneFunc = nil // Clear DoneFunc because it contains non-predictable variable data which interferes with the tests
		ch.metrics = append(ch.metrics, m)
	}
}

// DispatchMetricMap re-dispatches a metric map through BackendHandler.DispatchMetrics
func (ch *countingHandler) DispatchMetricMap(ctx context.Context, mm *gostatsd.MetricMap) {
	ch.dispatchMetrics(ctx, mm.AsMetrics())
}

func (ch *countingHandler) DispatchEvent(ctx context.Context, e *gostatsd.Event) {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	ch.events = append(ch.events, e)
}

func (ch *countingHandler) WaitForEvents() {
}

/*
func testContext(t *testing.T) (context.Context, func()) {
	ctxTest, completeTest := context.WithTimeout(context.Background(), 1100*time.Millisecond)
	go func() {
		after := time.NewTimer(1 * time.Second)
		select {
		case <-ctxTest.Done():
			after.Stop()
		case <-after.C:
			require.Fail(t, "test timed out")
		}
	}()
	return ctxTest, completeTest
}
*/
