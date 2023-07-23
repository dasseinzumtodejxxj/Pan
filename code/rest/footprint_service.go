package rest

import (
	"encoding/json"
	"tankmaster/code/core"
	"tankmaster/code/tool/util"

	"net/http"
	"time"
)

// @Service
type FootprintService struct {
	BaseBean
	footprintDao *FootprintDao
	userDao      *UserDao
}

func (this *FootprintService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.footprintDao)
	if b, ok := b.(*FootprintDao); ok {
		this.footprintDao = b
	}

	b = core.CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

}

func (this *FootprintService) Detail(uuid string) *Footprint {

	footprint := this.footprintDao.CheckByUuid(uuid)

	return footprint
}

// log a request.
func (this *FootprintService) Trace(request *http.Request, duration time.Duration, success bool) {

	params := make(map[string][]string)

	//POST params
	values := request.PostForm
	for key, val := range values {
		params[key] = val
	}
	//GET params
	values1 := request.URL.Query()
	for key, val := range values1 {
		params[key] = val
	}

	//ignore password.
	for key, _ := range params {
		if key == core.PASSWORD_KEY || key == "password" || key == "adminPassword" {
			params[key] = []string{"******"}
		}
	}

	paramsString := "{}"
	paramsData, err := json.Marshal(params)
	if err == nil {
		paramsString = string(paramsData)
	}

	footprint := &Footprint{
		Ip:      util.GetIpAddress(request),
		Host:    request.Host,
		Uri:     request.URL.Path,
		Params:  paramsString,
		Cost:    int64(duration / time.Millisecond),
		Success: success,
	}

	//if db not config just print content.
	if core.CONFIG.Installed() {
		user := this.findUser(request)
		userUuid := ""
		if user != nil {
			userUuid = user.Uuid
		}
		footprint.UserUuid = userUuid
		footprint = this.footprintDao.Create(footprint)
	}

	this.logger.Info("Ip:%s Cost:%d Uri:%s Params:%s", footprint.Ip, int64(duration/time.Millisecond), footprint.Uri, paramsString)

}

func (this *FootprintService) Bootstrap() {

	this.logger.Info("Immediately delete Footprint data of 8 days ago.")

	go core.RunWithRecovery(this.CleanOldData)
}

func (this *FootprintService) CleanOldData() {

	day8Ago := time.Now()
	day8Ago = day8Ago.AddDate(0, 0, -8)
	day8Ago = util.FirstSecondOfDay(day8Ago)

	this.logger.Info("Delete footprint data before %s", util.ConvertTimeToDateTimeString(day8Ago))

	this.footprintDao.DeleteByCreateTimeBefore(day8Ago)
}
