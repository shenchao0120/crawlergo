package config

import (
	"testing"
	"fmt"
)

func TestGetConfig(t *testing.T) {
	instance:=GetConfig()
	fmt.Println(instance.Database.ConnectPool)
	fmt.Println(GetDBConnectString())
}
