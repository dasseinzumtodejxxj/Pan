package core

import (
	"gorm.io/gorm"
	"net/http"
	"tankmaster/code/tool/cache"
)

type Context interface {
	http.Handler

	//get the gorm.DB. all the db connection will use this
	GetDB() *gorm.DB

	GetBean(bean Bean) Bean

	//get the global session cache
	GetSessionCache() *cache.Table

	GetControllerMap() map[string]Controller

	//when application installed. this method will invoke every bean's Bootstrap method
	InstallOk()

	//this method will invoke every bean's Cleanup method
	Cleanup()
}
