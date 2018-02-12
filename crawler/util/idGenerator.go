package util

import (
	"sync"
	"math"
)


// ID生成器的接口类型。
type IdGenerator interface {
	GetUint32Id() uint32 // 获得一个uint32类型的ID。
}

// 创建ID生成器。
func NewIdGenerator() IdGenerator {
	return &cyclicIdGenerator{}
}

// ID生成器的实现类型。
type cyclicIdGenerator struct {
	sn    uint32     // 当前的ID。
	ended bool       // 前一个ID是否已经为其类型所能表示的最大值。
	sync.Mutex // 互斥锁。
}

func (gen *cyclicIdGenerator) GetUint32Id() uint32 {
	gen.Lock()
	defer gen.Unlock()
	if gen.ended {
		defer func() { gen.ended = false }()
		gen.sn = 0
		return gen.sn
	}
	id := gen.sn
	if id < math.MaxUint32 {
		gen.sn++
	} else {
		gen.ended = true
	}
	return id
}