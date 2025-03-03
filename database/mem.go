package database

import (
	"sync"

	"github.com/ohnomail00/super-duper-s3/engine"
)

type Store interface {
	Save(key string, plan engine.FileUploadPlan)
	Get(key string) (engine.FileUploadPlan, bool)
}

type Mem struct {
	plans map[string]engine.FileUploadPlan
	mu    sync.RWMutex
}

func New() *Mem {
	return &Mem{
		plans: make(map[string]engine.FileUploadPlan),
	}
}

func (ms *Mem) Save(key string, plan engine.FileUploadPlan) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.plans[key] = plan
}

func (ms *Mem) Get(key string) (engine.FileUploadPlan, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	plan, exists := ms.plans[key]
	return plan, exists
}
