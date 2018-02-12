package util


import (
	"reflect"
	"github.com/astaxie/beego/logs"
	"sync"
	"fmt"
)

func TestNewPool(t *testing.T) {
	//LoggerInit()
	const chanNum = 5
	idGenerator := util.NewIdGenerator()
	eType := reflect.TypeOf(&TestEntity{})
	pool, err := NewPool(chanNum, eType, func() Entity {
		return &TestEntity{id: idGenerator.GetUint32Id()}
	})
	if err != nil {
		logs.Error("The error is ", err)
	}
	var wg sync.WaitGroup

	for i := 0; i < chanNum+1; i++ {
		wg.Add(1)
		time.Sleep(1*time.Second)

		go func(i int) {
			defer wg.Done()
			entity,err:=pool.Take()
			if err!=nil {
				fmt.Println("The error is",err)
			}
			logs.Info("Gorouting %d Take the entity %d, pool used %d\n",i,entity.Id(),pool.Used())
			time.Sleep(8*time.Second)
			err=pool.Return(entity)
			if err!=nil {
				logs.Error("The error is",err)
			}
		}(i)
	}
	wg.Wait()
}