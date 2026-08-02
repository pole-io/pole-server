package ratelimitv2

import (
	"context"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"

	apiv2 "github.com/pole-io/specification/source/go/api/v1/traffic_manage/ratelimiter"

	limiterapi "github.com/pole-io/pole-server/pkg/limiter/internal/api/v2"
	"github.com/pole-io/pole-server/pkg/limiter/internal/utils"
)

const (
	leaseCleanupInterval = time.Second
	leaseRecordRetention = 5 * time.Minute
)

type leaseStatus uint8

const (
	leaseActive leaseStatus = iota
	leaseSettled
	leaseExpired
)

type leaseCounter struct {
	counterKey uint32
	amount     uint32
	consumed   uint32
	accounting apiv2.QuotaAccounting
	counter    CounterV2
}

type quotaLease struct {
	id             string
	clientKey      uint32
	idempotencyKey string
	ttlSeconds     uint32
	expiresAt      int64
	terminalAt     int64
	status         leaseStatus
	sequence       uint64
	lastUpdate     map[uint32]uint32
	counters       map[uint32]*leaseCounter
	counterKeys    []uint32
	reserveLefts   []*apiv2.QuotaLeft
	settleResponse *apiv2.QuotaSettleResponse
}

type resolvedReservation struct {
	counterKey uint32
	amount     uint32
	counter    CounterV2
	left       *apiv2.QuotaLeft
}

func (s *Server) startLeaseCleanup(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(leaseCleanupInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				s.ExpireQuotaLeases(now)
			}
		}
	}()
}

// ReserveQuota 原子预留请求中的全部 counter。
func (s *Server) ReserveQuota(client Client, request *apiv2.QuotaReserveRequest) *apiv2.QuotaReserveResponse {
	nowMs := utils.CurrentMillisecond()
	s.leaseMutex.Lock()
	defer s.leaseMutex.Unlock()
	s.cleanupQuotaLeasesLocked(nowMs)

	if client == nil || request == nil || request.GetClientKey() == 0 || request.GetClientKey() != client.ClientKey() {
		return newQuotaReserveResponse(limiterapi.InvalidClientKey, nowMs)
	}
	if request.GetIdempotencyKey() == "" {
		return newQuotaReserveResponse(limiterapi.InvalidIdempotencyKey, nowMs)
	}
	if request.GetTtlSeconds() == 0 || len(request.GetReservations()) == 0 {
		return newQuotaReserveResponse(limiterapi.InvalidReservation, nowMs)
	}

	idempotencyKey := leaseIdempotencyKey(request.GetClientKey(), request.GetIdempotencyKey())
	if leaseID, ok := s.leaseByIdempotency[idempotencyKey]; ok {
		lease := s.quotaLeases[leaseID]
		if lease == nil || !s.sameReservationRequest(lease, request) {
			return newQuotaReserveResponse(limiterapi.IdempotencyConflict, nowMs)
		}
		if lease.status == leaseExpired {
			response := newQuotaReserveResponse(limiterapi.LeaseExpired, nowMs)
			response.LeaseId = lease.id
			response.ExpiresAt = lease.expiresAt
			return response
		}
		return lease.reserveResponse(nowMs)
	}

	reservations, code := s.resolveReservations(request, nowMs)
	if code != limiterapi.ExecuteSuccess {
		return newQuotaReserveResponse(code, nowMs)
	}

	quotaLefts := make([]*apiv2.QuotaLeft, 0, len(reservations))
	for _, reservation := range reservations {
		reserved := s.reservedByCounter[reservation.counterKey]
		available := reservation.left.GetLeft() - int64(reserved)
		quotaLefts = append(quotaLefts, cloneQuotaLeft(reservation.left, s.boxCounterKey(reservation.counterKey), available))
		if uint64(reservation.amount) > uint64(maxInt64(available, 0)) {
			response := newQuotaReserveResponse(limiterapi.QuotaExceeded, nowMs)
			response.QuotaLefts = quotaLefts
			return response
		}
	}

	lease := &quotaLease{
		id:             uuid.NewString(),
		clientKey:      request.GetClientKey(),
		idempotencyKey: request.GetIdempotencyKey(),
		ttlSeconds:     request.GetTtlSeconds(),
		expiresAt:      nowMs + int64(request.GetTtlSeconds())*int64(time.Second/time.Millisecond),
		status:         leaseActive,
		counters:       make(map[uint32]*leaseCounter, len(reservations)),
		counterKeys:    make([]uint32, 0, len(reservations)),
		reserveLefts:   make([]*apiv2.QuotaLeft, 0, len(reservations)),
	}
	for idx, reservation := range reservations {
		s.reservedByCounter[reservation.counterKey] += uint64(reservation.amount)
		reservation.counter.PinLease()
		lease.counters[reservation.counterKey] = &leaseCounter{
			counterKey: reservation.counterKey,
			amount:     reservation.amount,
			accounting: reservation.counter.Accounting(),
			counter:    reservation.counter,
		}
		lease.counterKeys = append(lease.counterKeys, reservation.counterKey)
		left := quotaLefts[idx]
		left.Left -= int64(reservation.amount)
		lease.reserveLefts = append(lease.reserveLefts, left)
	}
	s.quotaLeases[lease.id] = lease
	s.leaseByIdempotency[idempotencyKey] = lease.id
	return lease.reserveResponse(nowMs)
}

// UpdateQuota 更新各 counter 的累计消费量。
func (s *Server) UpdateQuota(client Client, request *apiv2.QuotaUpdateRequest) *apiv2.QuotaUpdateResponse {
	nowMs := utils.CurrentMillisecond()
	s.leaseMutex.Lock()
	defer s.leaseMutex.Unlock()
	s.cleanupQuotaLeasesLocked(nowMs)

	lease, code := s.activeLease(client, request.GetClientKey(), request.GetLeaseId())
	if code != limiterapi.ExecuteSuccess {
		return newQuotaUpdateResponse(code, nowMs)
	}
	consumptions, code := s.normalizeConsumptions(lease, request.GetConsumptions())
	if code != limiterapi.ExecuteSuccess {
		return newQuotaUpdateResponse(code, nowMs)
	}
	if request.GetSequence() < lease.sequence {
		return newQuotaUpdateResponse(limiterapi.InvalidSequence, nowMs)
	}
	if request.GetSequence() == lease.sequence {
		if lease.lastUpdate != nil && equalConsumptions(lease.lastUpdate, consumptions) {
			return s.updateResponse(lease, nowMs)
		}
		return newQuotaUpdateResponse(limiterapi.InvalidSequence, nowMs)
	}
	if code = validateMonotonicConsumptions(lease, consumptions); code != limiterapi.ExecuteSuccess {
		return newQuotaUpdateResponse(code, nowMs)
	}

	s.applyConsumptionsLocked(lease, consumptions, nowMs)
	lease.sequence = request.GetSequence()
	lease.lastUpdate = consumptions
	return s.updateResponse(lease, nowMs)
}

// SettleQuota 结算租约并归还未消费配额或全部占用型配额。
func (s *Server) SettleQuota(client Client, request *apiv2.QuotaSettleRequest) *apiv2.QuotaSettleResponse {
	nowMs := utils.CurrentMillisecond()
	s.leaseMutex.Lock()
	defer s.leaseMutex.Unlock()
	s.cleanupQuotaLeasesLocked(nowMs)

	if client == nil || request == nil || request.GetClientKey() == 0 || request.GetClientKey() != client.ClientKey() {
		return newQuotaSettleResponse(limiterapi.InvalidClientKey, nowMs)
	}
	lease := s.quotaLeases[request.GetLeaseId()]
	if lease == nil || lease.clientKey != request.GetClientKey() {
		return newQuotaSettleResponse(limiterapi.LeaseNotFound, nowMs)
	}
	if lease.status == leaseExpired {
		return newQuotaSettleResponse(limiterapi.LeaseExpired, nowMs)
	}
	consumptions, code := s.normalizeConsumptions(lease, request.GetConsumptions())
	if code != limiterapi.ExecuteSuccess {
		return newQuotaSettleResponse(code, nowMs)
	}
	if lease.status == leaseSettled {
		if request.GetSequence() == lease.sequence && equalConsumptions(lease.lastUpdate, consumptions) {
			return cloneSettleResponse(lease.settleResponse, nowMs)
		}
		return newQuotaSettleResponse(limiterapi.InvalidSequence, nowMs)
	}
	if request.GetSequence() <= lease.sequence {
		return newQuotaSettleResponse(limiterapi.InvalidSequence, nowMs)
	}
	if code = validateMonotonicConsumptions(lease, consumptions); code != limiterapi.ExecuteSuccess {
		return newQuotaSettleResponse(code, nowMs)
	}

	s.applyConsumptionsLocked(lease, consumptions, nowMs)
	settlements := make([]*apiv2.QuotaSettlement, 0, len(lease.counterKeys))
	for _, counterKey := range lease.counterKeys {
		entry := lease.counters[counterKey]
		returned := entry.amount
		if entry.accounting == apiv2.QuotaAccounting_CONSUMABLE {
			returned -= entry.consumed
		}
		s.releaseReservedLocked(counterKey, uint64(returned))
		entry.counter.UnpinLease()
		settlements = append(settlements, &apiv2.QuotaSettlement{
			CounterKey:     s.boxCounterKey(counterKey),
			ConsumedTotal:  entry.consumed,
			ReturnedAmount: returned,
		})
	}
	lease.status = leaseSettled
	lease.sequence = request.GetSequence()
	lease.lastUpdate = consumptions
	lease.terminalAt = nowMs
	lease.settleResponse = &apiv2.QuotaSettleResponse{
		Code:        uint32(limiterapi.ExecuteSuccess),
		Settlements: settlements,
		Timestamp:   nowMs,
	}
	return cloneSettleResponse(lease.settleResponse, nowMs)
}

// ExpireQuotaLeases 释放给定时间点之前过期的活跃租约。
func (s *Server) ExpireQuotaLeases(now time.Time) int {
	s.leaseMutex.Lock()
	defer s.leaseMutex.Unlock()
	return s.cleanupQuotaLeasesLocked(now.UnixMilli())
}

func (s *Server) resolveReservations(request *apiv2.QuotaReserveRequest, nowMs int64) ([]resolvedReservation, limiterapi.Code) {
	reservations := make([]resolvedReservation, 0, len(request.GetReservations()))
	seen := make(map[uint32]struct{}, len(request.GetReservations()))
	for _, reservation := range request.GetReservations() {
		if reservation == nil || reservation.GetAmount() == 0 {
			return nil, limiterapi.InvalidReservation
		}
		code, counterKey := s.unboxCounterKey(reservation.GetCounterKey())
		if code != limiterapi.ExecuteSuccess {
			return nil, code
		}
		if _, ok := seen[counterKey]; ok {
			return nil, limiterapi.InvalidReservation
		}
		seen[counterKey] = struct{}{}
		code, counter := s.counterMng.GetCounter(counterKey)
		if code != limiterapi.ExecuteSuccess {
			return nil, code
		}
		counter.Update()
		reservations = append(reservations, resolvedReservation{
			counterKey: counterKey,
			amount:     reservation.GetAmount(),
			counter:    counter,
			left:       counter.SumQuota(nowMs),
		})
	}
	sort.Slice(reservations, func(i, j int) bool {
		return reservations[i].counterKey < reservations[j].counterKey
	})
	return reservations, limiterapi.ExecuteSuccess
}

func (s *Server) sameReservationRequest(lease *quotaLease, request *apiv2.QuotaReserveRequest) bool {
	if lease.clientKey != request.GetClientKey() || lease.ttlSeconds != request.GetTtlSeconds() ||
		len(lease.counters) != len(request.GetReservations()) {
		return false
	}
	seen := make(map[uint32]struct{}, len(request.GetReservations()))
	for _, reservation := range request.GetReservations() {
		code, counterKey := s.unboxCounterKey(reservation.GetCounterKey())
		if code != limiterapi.ExecuteSuccess {
			return false
		}
		entry := lease.counters[counterKey]
		if entry == nil || entry.amount != reservation.GetAmount() {
			return false
		}
		if _, ok := seen[counterKey]; ok {
			return false
		}
		seen[counterKey] = struct{}{}
	}
	return true
}

func (s *Server) activeLease(client Client, clientKey uint32, leaseID string) (*quotaLease, limiterapi.Code) {
	if client == nil || clientKey == 0 || clientKey != client.ClientKey() {
		return nil, limiterapi.InvalidClientKey
	}
	lease := s.quotaLeases[leaseID]
	if lease == nil || lease.clientKey != clientKey {
		return nil, limiterapi.LeaseNotFound
	}
	if lease.status == leaseExpired {
		return nil, limiterapi.LeaseExpired
	}
	if lease.status != leaseActive {
		return nil, limiterapi.LeaseSettled
	}
	return lease, limiterapi.ExecuteSuccess
}

func (s *Server) normalizeConsumptions(
	lease *quotaLease, values []*apiv2.QuotaConsumption) (map[uint32]uint32, limiterapi.Code) {
	consumptions := make(map[uint32]uint32, len(values))
	for _, value := range values {
		if value == nil {
			return nil, limiterapi.InvalidConsumption
		}
		code, counterKey := s.unboxCounterKey(value.GetCounterKey())
		if code != limiterapi.ExecuteSuccess || lease.counters[counterKey] == nil {
			return nil, limiterapi.InvalidCounterKey
		}
		if _, ok := consumptions[counterKey]; ok {
			return nil, limiterapi.InvalidConsumption
		}
		consumptions[counterKey] = value.GetConsumedTotal()
	}
	return consumptions, limiterapi.ExecuteSuccess
}

func validateMonotonicConsumptions(lease *quotaLease, consumptions map[uint32]uint32) limiterapi.Code {
	for counterKey, consumed := range consumptions {
		entry := lease.counters[counterKey]
		if consumed < entry.consumed || consumed > entry.amount {
			return limiterapi.InvalidConsumedTotal
		}
	}
	return limiterapi.ExecuteSuccess
}

func (s *Server) applyConsumptionsLocked(lease *quotaLease, consumptions map[uint32]uint32, nowMs int64) {
	for counterKey, consumed := range consumptions {
		entry := lease.counters[counterKey]
		delta := consumed - entry.consumed
		if delta == 0 {
			continue
		}
		if entry.accounting == apiv2.QuotaAccounting_CONSUMABLE {
			entry.counter.CommitQuota(delta, nowMs)
			s.releaseReservedLocked(counterKey, uint64(delta))
		}
		entry.consumed = consumed
		entry.counter.Update()
	}
}

func (s *Server) cleanupQuotaLeasesLocked(nowMs int64) int {
	expired := 0
	for leaseID, lease := range s.quotaLeases {
		if lease.status == leaseActive && lease.expiresAt <= nowMs {
			for _, counterKey := range lease.counterKeys {
				entry := lease.counters[counterKey]
				reserved := entry.amount
				if entry.accounting == apiv2.QuotaAccounting_CONSUMABLE {
					reserved -= entry.consumed
				}
				s.releaseReservedLocked(counterKey, uint64(reserved))
				entry.counter.UnpinLease()
			}
			lease.status = leaseExpired
			lease.terminalAt = nowMs
			expired++
		}
		if lease.status != leaseActive && nowMs-lease.terminalAt >= leaseRecordRetention.Milliseconds() {
			delete(s.quotaLeases, leaseID)
			idempotencyKey := leaseIdempotencyKey(lease.clientKey, lease.idempotencyKey)
			if s.leaseByIdempotency[idempotencyKey] == leaseID {
				delete(s.leaseByIdempotency, idempotencyKey)
			}
		}
	}
	return expired
}

func (s *Server) releaseReservedLocked(counterKey uint32, amount uint64) {
	reserved := s.reservedByCounter[counterKey]
	if amount >= reserved {
		delete(s.reservedByCounter, counterKey)
		return
	}
	s.reservedByCounter[counterKey] = reserved - amount
}

func (lease *quotaLease) reserveResponse(nowMs int64) *apiv2.QuotaReserveResponse {
	quotaLefts := make([]*apiv2.QuotaLeft, 0, len(lease.reserveLefts))
	for _, left := range lease.reserveLefts {
		quotaLefts = append(quotaLefts, cloneQuotaLeft(left, left.GetCounterKey(), left.GetLeft()))
	}
	return &apiv2.QuotaReserveResponse{
		Code:       uint32(limiterapi.ExecuteSuccess),
		LeaseId:    lease.id,
		QuotaLefts: quotaLefts,
		ExpiresAt:  lease.expiresAt,
		Timestamp:  nowMs,
	}
}

func (s *Server) updateResponse(lease *quotaLease, nowMs int64) *apiv2.QuotaUpdateResponse {
	consumptions := make([]*apiv2.QuotaConsumption, 0, len(lease.counterKeys))
	for _, counterKey := range lease.counterKeys {
		consumptions = append(consumptions, &apiv2.QuotaConsumption{
			CounterKey:    s.boxCounterKey(counterKey),
			ConsumedTotal: lease.counters[counterKey].consumed,
		})
	}
	return &apiv2.QuotaUpdateResponse{
		Code:         uint32(limiterapi.ExecuteSuccess),
		Consumptions: consumptions,
		Sequence:     lease.sequence,
		Timestamp:    nowMs,
	}
}

func cloneSettleResponse(response *apiv2.QuotaSettleResponse, nowMs int64) *apiv2.QuotaSettleResponse {
	settlements := make([]*apiv2.QuotaSettlement, 0, len(response.GetSettlements()))
	for _, settlement := range response.GetSettlements() {
		settlements = append(settlements, &apiv2.QuotaSettlement{
			CounterKey:     settlement.GetCounterKey(),
			ConsumedTotal:  settlement.GetConsumedTotal(),
			ReturnedAmount: settlement.GetReturnedAmount(),
		})
	}
	return &apiv2.QuotaSettleResponse{Code: response.GetCode(), Settlements: settlements, Timestamp: nowMs}
}

func newQuotaReserveResponse(code limiterapi.Code, nowMs int64) *apiv2.QuotaReserveResponse {
	return &apiv2.QuotaReserveResponse{Code: uint32(code), Timestamp: nowMs}
}

func newQuotaUpdateResponse(code limiterapi.Code, nowMs int64) *apiv2.QuotaUpdateResponse {
	return &apiv2.QuotaUpdateResponse{Code: uint32(code), Timestamp: nowMs}
}

func newQuotaSettleResponse(code limiterapi.Code, nowMs int64) *apiv2.QuotaSettleResponse {
	return &apiv2.QuotaSettleResponse{Code: uint32(code), Timestamp: nowMs}
}

func cloneQuotaLeft(left *apiv2.QuotaLeft, counterKey uint32, amount int64) *apiv2.QuotaLeft {
	return &apiv2.QuotaLeft{
		CounterKey:  counterKey,
		Left:        amount,
		Mode:        left.GetMode(),
		ClientCount: left.GetClientCount(),
	}
}

func equalConsumptions(left, right map[uint32]uint32) bool {
	if len(left) != len(right) {
		return false
	}
	for counterKey, consumed := range left {
		if right[counterKey] != consumed {
			return false
		}
	}
	return true
}

func leaseIdempotencyKey(clientKey uint32, key string) string {
	return strconv.FormatUint(uint64(clientKey), 10) + "\x00" + key
}

func maxInt64(value, minimum int64) int64 {
	if value < minimum {
		return minimum
	}
	return value
}
