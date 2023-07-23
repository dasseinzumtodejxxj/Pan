package main

import (
	"tankmaster/code/core"
	"tankmaster/code/support"
)

func main() {

	core.APPLICATION = &support.TankApplication{}
	core.APPLICATION.Start()

}
