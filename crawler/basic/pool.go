package basic

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

type Entity interface {
	Id() uint32
}

type Pool interface {
	Take() (Entity, error)
	Return(entity Entity) error
	Total() uint32
	Used() uint32
}

type poolImpl struct {
	total         uint32
	entityType    reflect.Type
	getEntity     func() Entity
	entityChan    chan Entity
	entityUsedMap map[uint32]bool
	sync.Mutex
}

func NewPool(total uint32, entityType reflect.Type, getEntity func() Entity) (Pool, error) {
	if total <= 0 || entityType == nil || getEntity == nil {
		errMsg :=
			fmt.Sprintf("The pool can not be initialized!")
		return nil, errors.New(errMsg)
	}
	entityChan := make(chan Entity, total)
	entityUsedMap := make(map[uint32]bool)

	for i := 0; i < int(total); i++ {
		entity := getEntity()
		if entityType != reflect.TypeOf(entity) {
			errMsg :=
				fmt.Sprintf("The type of result of function genEntity() is NOT %s!\n!", entityType)
			return nil, errors.New(errMsg)
		}
		entityChan <- entity
		entityUsedMap[entity.Id()] = true
	}
	return &poolImpl{
		total:         total,
		entityType:    entityType,
		getEntity:     getEntity,
		entityChan:    entityChan,
		entityUsedMap: entityUsedMap}, nil
}

func (pl *poolImpl) Take() (Entity, error) {
	entity, ok := <-pl.entityChan
	if !ok {
		return nil, errors.New("Get entity from chan error!")
	}
	pl.Lock()
	defer pl.Unlock()
	pl.entityUsedMap[entity.Id()] = true
	return entity, nil
}

func (pl *poolImpl) Return(entity Entity) error {
	if entity == nil {
		return errors.New("The return entity cannot be nil.")
	}

	if pl.entityType != reflect.TypeOf(entity) {
		return errors.New("The return entity type is not the pool type.")
	}
	result:=pl.compareAndSetUsedMap(entity.Id(),true,false)

	if result == CompareSuccess {
		pl.entityChan <- entity
		return nil
	}else if result == ValueChanged{
		errMsg := fmt.Sprintf("The entity (id=%d) is already in the pool!\n", entity.Id())
		return errors.New(errMsg)
	}else {
		errMsg := fmt.Sprintf("The entity (id=%d) is not in the pool!\n", entity.Id())
		return errors.New(errMsg)
	}
}

type CompareResult int

const(
	EntityNotFound CompareResult = iota
	ValueChanged
	CompareSuccess
)

func (pl *poolImpl) compareAndSetUsedMap(entityId uint32, oldValue bool, newValue bool) CompareResult {
	pl.Lock()
	defer pl.Unlock()
	value, ok := pl.entityUsedMap[entityId]
	if !ok {
		return EntityNotFound
	}
	if value != oldValue {
		return ValueChanged
	}
	pl.entityUsedMap[entityId] = newValue
	return CompareSuccess

}
func (pl *poolImpl) Total() uint32 {
	return pl.total
}
func (pl *poolImpl) Used() uint32 {
	return pl.total - uint32(len(pl.entityChan))
}
