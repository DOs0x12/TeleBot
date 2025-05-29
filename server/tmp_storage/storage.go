package tmp_storage

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// TmpStorage stores objects with timestamps.
type TmpStorage struct {
	storage *sync.Map
}

type storageObj struct {
	Obj       any
	timeStamp time.Time
}

// NewTmpStorage creates a new TmpStorage instance.
func NewTmpStorage() TmpStorage {
	return TmpStorage{storage: &sync.Map{}}
}

// StartCleanupOldObjs starts a worker for cleaning up the old objects.
func (s TmpStorage) StartCleanupOldObjs(ctx context.Context, lifeTime, period time.Duration) {
	go s.delOldObjsWithPeriod(ctx, lifeTime, period)
}

func (s TmpStorage) delOldObjsWithPeriod(ctx context.Context, lifeTime, period time.Duration) {
	t := time.NewTicker(period)

	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.delOldObjs(lifeTime)
		}
	}
}

func (s TmpStorage) delOldObjs(lifeTime time.Duration) {
	now := time.Now()
	uuidsForDel := make([]uuid.UUID, 0)

	s.storage.Range(func(key, value any) bool {
		if value.(storageObj).timeStamp.Add(lifeTime).Before(now) {
			uuidsForDel = append(uuidsForDel, key.(uuid.UUID))
		}
		return true
	})

	for _, uuid := range uuidsForDel {
		s.storage.Delete(uuid)
	}
}

// AddObj adds an object to the storage.
func (s TmpStorage) AddObj(obj any) uuid.UUID {
	objUuid := uuid.New()
	sObj := storageObj{Obj: obj, timeStamp: time.Now()}
	s.storage.Store(objUuid, sObj)

	return objUuid
}

// DelObj deletes a single object from the storage by UUID.
func (s TmpStorage) DelObj(objUuid uuid.UUID) {
	s.storage.Delete(objUuid)
}

// GetObj gets an object by UUID.
func (s TmpStorage) GetObj(objUuid uuid.UUID) (storageObj, bool) {
	obj, ok := s.storage.Load(objUuid)
	if !ok {
		return storageObj{}, ok
	}
	return obj.(storageObj), ok
}
