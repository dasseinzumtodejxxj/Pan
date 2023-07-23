package test

import (
	"fmt"
	"strings"
	"tankmaster/code/tool/util"
	"testing"
	"time"
)

func TestHello(t *testing.T) {

	split := strings.Split("good", "/")
	fmt.Printf("%v", split)

	var i int
	for i = 1; i < 10; i++ {
		fmt.Printf("i=%d\n", i)
	}

}

func TestDayAgo(t *testing.T) {

	dayAgo := time.Now()
	dayAgo = dayAgo.AddDate(0, 0, -8)

	thenDay := util.FirstSecondOfDay(dayAgo)

	fmt.Printf("%s\n", util.ConvertTimeToDateTimeString(thenDay))

}
