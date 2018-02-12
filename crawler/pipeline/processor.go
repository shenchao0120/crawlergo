package itemproc

import "chaoshen.com/crawlergo/crawler/basic"

type ProcessItem func(item basic.ItemMap) (result basic.ItemMap, err error)
