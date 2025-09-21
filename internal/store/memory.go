package store

import (
	"sync"

	"admira/internal/model"
)

type Memory struct {
	mu       sync.RWMutex
	ads      map[string]model.AdsPerf // key: AdsPerf.Key()
	crm      map[string]model.CRMOpp  // key: CRMOpp.Key()
}

func NewMemory() *Memory {
	return &Memory{
		ads: make(map[string]model.AdsPerf),
		crm: make(map[string]model.CRMOpp),
	}
}

func (m *Memory) UpsertAds(a model.AdsPerf) {
	m.mu.Lock(); defer m.mu.Unlock()
	m.ads[a.Key()] = a
}
func (m *Memory) UpsertCRM(o model.CRMOpp) {
	m.mu.Lock(); defer m.mu.Unlock()
	m.crm[o.Key()] = o
}

func (m *Memory) AllAds() []model.AdsPerf {
	m.mu.RLock(); defer m.mu.RUnlock()
	out := make([]model.AdsPerf, 0, len(m.ads))
	for _, v := range m.ads { out = append(out, v) }
	return out
}
func (m *Memory) AllCRM() []model.CRMOpp {
	m.mu.RLock(); defer m.mu.RUnlock()
	out := make([]model.CRMOpp, 0, len(m.crm))
	for _, v := range m.crm { out = append(out, v) }
	return out
}
