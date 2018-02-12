package util

import (
	"sync"
	"fmt"
)

type StopSign interface {
	SignStop() bool
	IsSigned() bool
	Reset()
	Record(code string )
	RecordCount(code string )uint
	RecordTotal() uint
	Summary() string
}


type StopSignImpl struct {
	sync.RWMutex
	signed bool
	recodeCountMap map[string]uint
}

func  NewStopSign () StopSign {
	return &StopSignImpl{recodeCountMap:make(map[string]uint)}
}

func (ssi *StopSignImpl)SignStop() bool{
	ssi.Lock()
	defer ssi.Unlock()
	if ssi.IsSigned() {
		return false
	}
	ssi.signed=true
	return true
}

func (ssi *StopSignImpl)IsSigned() bool{
	return ssi.signed
}

func (ssi *StopSignImpl) Reset() {
	ssi.Lock()
	defer ssi.Unlock()
	ssi.signed=true
	ssi.recodeCountMap=make(map[string]uint)
}

func (ssi *StopSignImpl)Record (code string) {
	if !ssi.IsSigned() {
		return
	}
	ssi.Lock()
	defer ssi.Unlock()
	if record,ok:=ssi.recodeCountMap[code] ;!ok{
		ssi.recodeCountMap[code]=1
	}else{
		ssi.recodeCountMap[code]=record+1
	}
}

func (ssi *StopSignImpl)RecordCount (code string)uint {
	ssi.RLock()
	defer ssi.RUnlock()
	return ssi.recodeCountMap[code]
}

func (ssi *StopSignImpl)RecordTotal ()uint {
	ssi.RLock()
	defer ssi.RUnlock()
	var total uint
	for _,v:= range ssi.recodeCountMap{
		total+=v
	}
	return total
}


func (ssi *StopSignImpl) Summary() string {
	if ssi.signed {
		return fmt.Sprintf("signed: true, recodeCountMap: %v", ssi.recodeCountMap)
	} else {
		return "signed: false"
	}
}