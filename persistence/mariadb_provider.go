package persistence

import (
	"sync"

	"github.com/golang/protobuf/proto"
)

type entry struct {
	eventIndex int // the event index right after snapshot
	snapshot   proto.Message
	events     []proto.Message
}

type MariaDBProvider struct {
	snapshotInterval int
	mu               sync.RWMutex
	store            map[string]*entry // actorName -> a persistence entry
}

func NewMariaDBProvider(snapshotInterval int) *MariaDBProvider {
	return &MariaDBProvider{
		snapshotInterval: snapshotInterval,
		store:            make(map[string]*entry),
	}
}

// loadOrInit returns the existing entry for actorName if present.
// Otherwise, it initializes and returns an empty entry.
// The loaded result is true if the entry was loaded, false if initialized.
func (provider *MariaDBProvider) loadOrInit(actorName string) (e *entry, loaded bool) {
	provider.mu.RLock()
	e, ok := provider.store[actorName]
	provider.mu.RUnlock()

	if !ok {
		provider.mu.Lock()
		e = &entry{}
		provider.store[actorName] = e
		provider.mu.Unlock()
	}

	return e, ok
}

func (provider *MariaDBProvider) Restart() {}

func (provider *MariaDBProvider) GetSnapshotInterval() int {
	return provider.snapshotInterval
}

func (provider *MariaDBProvider) GetSnapshot(actorName string) (snapshot interface{}, eventIndex int, ok bool) {
	entry, loaded := provider.loadOrInit(actorName)
	if !loaded || entry.snapshot == nil {
		return nil, 0, false
	}
	return entry.snapshot, entry.eventIndex, true
}

func (provider *MariaDBProvider) PersistSnapshot(actorName string, eventIndex int, snapshot proto.Message) {
	entry, _ := provider.loadOrInit(actorName)
	entry.eventIndex = eventIndex
	entry.snapshot = snapshot
}

func (provider *MariaDBProvider) GetEvents(actorName string, eventIndexStart int, callback func(e interface{})) {
	entry, _ := provider.loadOrInit(actorName)
	for _, e := range entry.events[eventIndexStart:] {
		callback(e)
	}
}

func (provider *MariaDBProvider) PersistEvent(actorName string, eventIndex int, event proto.Message) {
	entry, _ := provider.loadOrInit(actorName)
	entry.events = append(entry.events, event)
}
