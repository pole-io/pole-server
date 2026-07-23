package otel

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pole-io/pole-server/pkg/common/localdb"
)

const (
	defaultLogBatchSize     = 100
	defaultLogFlushInterval = time.Second
)

type LogRecord struct {
	Timestamp    time.Time
	SeverityText string
	Body         string
	Attributes   map[string]string
}

type LogExporter struct {
	endpoint      string
	resourceAttrs map[string]string
	scopeName     string
	client        *http.Client
	queue         chan LogRecord
	stopCh        chan struct{}
	wakeCh        chan struct{}
	wg            sync.WaitGroup
	dropped       atomic.Uint64

	spoolDir      string
	spoolMaxBytes int64
	spoolDBPath   string
	spoolDB       *localdb.PebbleDB
	spoolPrefix   []byte
	spoolMu       sync.Mutex
	spoolSeq      atomic.Uint64
}

func NewLogExporter(conf *Config) (*LogExporter, error) {
	if conf == nil {
		return nil, fmt.Errorf("otel config is nil")
	}
	next := *conf
	next.setDefault()
	endpoint := normalizeLogsEndpoint(next.LogsEndpoint, next.Endpoint)
	if endpoint == "" {
		return nil, fmt.Errorf("otel logs endpoint is empty")
	}
	exporter := &LogExporter{
		endpoint: endpoint,
		resourceAttrs: map[string]string{
			"service.name":                next.ServiceName,
			"deployment.environment.name": next.Environment,
			"pole.cluster":                next.Cluster,
			"pole.node.role":              next.NodeRole,
		},
		scopeName: next.ServiceName,
		client:    &http.Client{Timeout: next.Timeout},
		queue:     make(chan LogRecord, next.LogQueueSize),
		stopCh:    make(chan struct{}),
		wakeCh:    make(chan struct{}, 1),
	}
	if next.LogSpoolDir != "" {
		exporter.spoolDir = next.LogSpoolDir
		exporter.spoolMaxBytes = next.LogSpoolMaxBytes
		exporter.spoolDBPath = filepath.Join(next.LogSpoolDir, "otel-events.pebble")
		spoolBucket := next.LogSpoolBucket
		if spoolBucket == "" {
			spoolBucket = "otel_logs"
		}
		exporter.spoolPrefix = []byte("otel-queue/" + spoolBucket + "/")
		if err := os.MkdirAll(next.LogSpoolDir, 0o755); err != nil {
			return nil, err
		}
		db, err := localdb.OpenPebble(exporter.spoolDBPath)
		if err != nil {
			return nil, err
		}
		exporter.spoolDB = db
	}
	if next.ServiceVersion != "" {
		exporter.resourceAttrs["service.version"] = next.ServiceVersion
	}
	exporter.wg.Add(1)
	go exporter.run()
	return exporter, nil
}

func (e *LogExporter) Emit(record LogRecord) {
	if e == nil {
		return
	}
	if e.spoolEnabled() {
		if e.appendSpool(record) {
			e.wake()
			return
		}
	}
	select {
	case e.queue <- record:
	default:
		e.dropped.Add(1)
	}
}

func (e *LogExporter) Dropped() uint64 {
	if e == nil {
		return 0
	}
	return e.dropped.Load()
}

func (e *LogExporter) Shutdown(ctx context.Context) error {
	if e == nil {
		return nil
	}
	close(e.stopCh)
	done := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		if e.spoolDB != nil {
			return e.spoolDB.Close()
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (e *LogExporter) run() {
	defer e.wg.Done()
	if e.spoolEnabled() {
		e.runSpool()
		return
	}
	ticker := time.NewTicker(defaultLogFlushInterval)
	defer ticker.Stop()

	batch := make([]LogRecord, 0, defaultLogBatchSize)
	flush := func() {
		if len(batch) == 0 {
			return
		}
		e.export(batch)
		batch = batch[:0]
	}

	for {
		select {
		case record := <-e.queue:
			batch = append(batch, record)
			if len(batch) >= defaultLogBatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-e.stopCh:
			drain := true
			for drain {
				select {
				case record := <-e.queue:
					batch = append(batch, record)
					if len(batch) >= defaultLogBatchSize {
						flush()
					}
				default:
					drain = false
				}
			}
			flush()
			return
		}
	}
}

func (e *LogExporter) runSpool() {
	ticker := time.NewTicker(defaultLogFlushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			e.flushSpool()
		case <-e.wakeCh:
			e.flushSpool()
		case <-e.stopCh:
			e.flushSpool()
			return
		}
	}
}

func (e *LogExporter) export(records []LogRecord) bool {
	payload, err := json.Marshal(e.toOTLP(records))
	if err != nil {
		return false
	}
	req, err := http.NewRequest(http.MethodPost, e.endpoint, bytes.NewReader(payload))
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := e.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode < http.StatusBadRequest
}

func (e *LogExporter) spoolEnabled() bool {
	return e != nil && e.spoolDB != nil
}

func (e *LogExporter) appendSpool(record LogRecord) bool {
	e.spoolMu.Lock()
	defer e.spoolMu.Unlock()
	if e.spoolMaxBytes > 0 {
		if size, err := dirSize(e.spoolDBPath); err == nil && size >= e.spoolMaxBytes {
			e.dropped.Add(1)
			return true
		}
	}
	payload, err := json.Marshal(record)
	if err != nil {
		e.dropped.Add(1)
		return true
	}
	if err := e.spoolDB.Set(e.nextSpoolKey(), payload, localdb.WriteSync); err != nil {
		e.dropped.Add(1)
		return false
	}
	return true
}

func (e *LogExporter) flushSpool() {
	keys, records := e.readSpoolBatch()
	if len(records) == 0 {
		if len(keys) > 0 {
			e.deleteSpoolBatch(keys)
		}
		return
	}
	if !e.export(records) {
		return
	}
	e.deleteSpoolBatch(keys)
}

func (e *LogExporter) readSpoolBatch() ([][]byte, []LogRecord) {
	e.spoolMu.Lock()
	defer e.spoolMu.Unlock()
	keys := make([][]byte, 0, defaultLogBatchSize)
	records := make([]LogRecord, 0, defaultLogBatchSize)
	if err := e.spoolDB.ScanPrefix(e.spoolPrefix, defaultLogBatchSize, func(key, value []byte) error {
		var record LogRecord
		if err := json.Unmarshal(value, &record); err != nil {
			keys = append(keys, key)
			return nil
		}
		keys = append(keys, key)
		records = append(records, record)
		return nil
	}); err != nil {
		return nil, nil
	}
	return keys, records
}

func (e *LogExporter) deleteSpoolBatch(keys [][]byte) {
	if len(keys) == 0 {
		return
	}
	e.spoolMu.Lock()
	defer e.spoolMu.Unlock()
	for _, key := range keys {
		_ = e.spoolDB.Delete(key, localdb.WriteSync)
	}
}

func (e *LogExporter) wake() {
	select {
	case e.wakeCh <- struct{}{}:
	default:
	}
}

func (e *LogExporter) nextSpoolKey() []byte {
	key := make([]byte, 0, len(e.spoolPrefix)+16)
	key = append(key, e.spoolPrefix...)
	var buf [16]byte
	binary.BigEndian.PutUint64(buf[:8], uint64(time.Now().UnixNano()))
	binary.BigEndian.PutUint64(buf[8:], e.spoolSeq.Add(1))
	key = append(key, buf[:]...)
	return key
}

func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.WalkDir(path, func(_ string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		size += info.Size()
		return nil
	})
	return size, err
}

func (e *LogExporter) toOTLP(records []LogRecord) map[string]any {
	logRecords := make([]map[string]any, 0, len(records))
	for _, record := range records {
		timestamp := record.Timestamp
		if timestamp.IsZero() {
			timestamp = time.Now()
		}
		severity := record.SeverityText
		if severity == "" {
			severity = "INFO"
		}
		logRecords = append(logRecords, map[string]any{
			"timeUnixNano":         strconv.FormatInt(timestamp.UnixNano(), 10),
			"observedTimeUnixNano": strconv.FormatInt(time.Now().UnixNano(), 10),
			"severityText":         severity,
			"body":                 map[string]any{"stringValue": record.Body},
			"attributes":           stringAttributes(record.Attributes),
		})
	}
	return map[string]any{
		"resourceLogs": []map[string]any{
			{
				"resource": map[string]any{
					"attributes": stringAttributes(e.resourceAttrs),
				},
				"scopeLogs": []map[string]any{
					{
						"scope": map[string]any{
							"name": e.scopeName,
						},
						"logRecords": logRecords,
					},
				},
			},
		},
	}
}

func stringAttributes(attrs map[string]string) []map[string]any {
	if len(attrs) == 0 {
		return []map[string]any{}
	}
	out := make([]map[string]any, 0, len(attrs))
	for key, value := range attrs {
		out = append(out, map[string]any{
			"key": key,
			"value": map[string]any{
				"stringValue": value,
			},
		})
	}
	return out
}

func normalizeLogsEndpoint(logsEndpoint, endpoint string) string {
	if logsEndpoint != "" {
		return strings.TrimRight(logsEndpoint, "/")
	}
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return ""
	}
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return strings.TrimRight(endpoint, "/") + "/v1/logs"
	}
	host := endpoint
	if strings.HasSuffix(host, ":4317") {
		host = strings.TrimSuffix(host, ":4317") + ":4318"
	}
	if parsed, err := url.Parse("http://" + host); err == nil && parsed.Host != "" {
		return "http://" + parsed.Host + "/v1/logs"
	}
	return ""
}
