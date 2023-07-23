package rest

import (
	"tankmaster/code/core"
)

// @Service
type BridgeService struct {
	BaseBean
	bridgeDao *BridgeDao
	userDao   *UserDao
}

func (this *BridgeService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.bridgeDao)
	if b, ok := b.(*BridgeDao); ok {
		this.bridgeDao = b
	}

	b = core.CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

}

func (this *BridgeService) Detail(uuid string) *Bridge {

	bridge := this.bridgeDao.CheckByUuid(uuid)

	return bridge
}
