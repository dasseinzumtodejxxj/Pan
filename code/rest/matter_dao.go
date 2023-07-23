package rest

import (
	"gorm.io/gorm"
	"math"
	"os"
	"tankmaster/code/core"
	"tankmaster/code/tool/builder"
	"tankmaster/code/tool/result"
	"tankmaster/code/tool/util"
	"tankmaster/code/tool/uuid"
	"time"
)

type MatterDao struct {
	BaseDao
	imageCacheDao *ImageCacheDao
	bridgeDao     *BridgeDao
}

func (this *MatterDao) Init() {
	this.BaseDao.Init()

	b := core.CONTEXT.GetBean(this.imageCacheDao)
	if b, ok := b.(*ImageCacheDao); ok {
		this.imageCacheDao = b
	}

	b = core.CONTEXT.GetBean(this.bridgeDao)
	if b, ok := b.(*BridgeDao); ok {
		this.bridgeDao = b
	}

}

func (this *MatterDao) FindByUuid(uuid string) *Matter {
	var entity = &Matter{}
	db := core.CONTEXT.GetDB().Where("uuid = ?", uuid).First(entity)
	if db.Error != nil {
		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			return nil
		} else {
			panic(db.Error)
		}
	}
	return entity
}

// find by uuid. if not found panic NotFound error
func (this *MatterDao) CheckByUuid(uuid string) *Matter {
	entity := this.FindByUuid(uuid)
	if entity == nil {
		panic(result.NotFound("not found record with uuid = %s", uuid))
	}
	return entity
}

// find by uuid. if uuid=root, then return the Root Matter
func (this *MatterDao) CheckWithRootByUuid(uuid string, user *User) *Matter {

	if uuid == "" {
		panic(result.BadRequest("uuid cannot be null."))
	}

	var matter *Matter
	if uuid == MATTER_ROOT {
		if user == nil {
			panic(result.BadRequest("user cannot be null."))
		}
		matter = NewRootMatter(user)
	} else {
		matter = this.CheckByUuid(uuid)
	}

	return matter
}

// find by path. if path=/, then return the Root Matter
func (this *MatterDao) CheckWithRootByPath(path string, user *User) *Matter {

	var matter *Matter

	if user == nil {
		panic(result.BadRequest("user cannot be null."))
	}

	if path == "" || path == "/" {
		matter = NewRootMatter(user)
	} else {
		matter = this.checkByUserUuidAndPath(user.Uuid, path)
	}

	return matter
}

// find by path. if path=/, then return the Root Matter
func (this *MatterDao) FindWithRootByPath(path string, user *User) *Matter {

	var matter *Matter

	if user == nil {
		panic(result.BadRequest("user cannot be null."))
	}

	if path == "" || path == "/" {
		matter = NewRootMatter(user)
	} else {
		matter = this.findByUserUuidAndPath(user.Uuid, path)
	}

	return matter
}

func (this *MatterDao) FindByUserUuidAndPuuidAndDirTrue(userUuid string, puuid string) []*Matter {

	var wp = &builder.WherePair{}

	if userUuid != "" {
		wp = wp.And(&builder.WherePair{Query: "user_uuid = ?", Args: []interface{}{userUuid}})
	}

	if puuid != "" {
		wp = wp.And(&builder.WherePair{Query: "puuid = ?", Args: []interface{}{puuid}})
	}

	wp = wp.And(&builder.WherePair{Query: "dir = ?", Args: []interface{}{1}})

	var matters []*Matter
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).First(&matters)

	if db.Error != nil {
		return nil
	}

	return matters
}

func (this *MatterDao) CheckByUuidAndUserUuid(uuid string, userUuid string) *Matter {

	var matter = &Matter{}
	db := core.CONTEXT.GetDB().Where(&Matter{Uuid: uuid, UserUuid: userUuid}).First(matter)
	this.PanicError(db.Error)

	return matter

}

func (this *MatterDao) CountByUserUuidAndPuuidAndDirAndName(userUuid string, puuid string, dir bool, name string) int {

	var matter Matter
	var count int64

	var wp = &builder.WherePair{}

	if puuid != "" {
		wp = wp.And(&builder.WherePair{Query: "puuid = ?", Args: []interface{}{puuid}})
	}

	if userUuid != "" {
		wp = wp.And(&builder.WherePair{Query: "user_uuid = ?", Args: []interface{}{userUuid}})
	}

	if name != "" {
		wp = wp.And(&builder.WherePair{Query: "name = ?", Args: []interface{}{name}})
	}

	wp = wp.And(&builder.WherePair{Query: "dir = ?", Args: []interface{}{dir}})

	db := core.CONTEXT.GetDB().
		Model(&matter).
		Where(wp.Query, wp.Args...).
		Count(&count)
	this.PanicError(db.Error)

	return int(count)
}

func (this *MatterDao) FindByUserUuidAndPuuidAndDirAndName(userUuid string, puuid string, dir string, name string) *Matter {

	var matter = &Matter{}

	var wp = &builder.WherePair{}

	if puuid != "" {
		wp = wp.And(&builder.WherePair{Query: "puuid = ?", Args: []interface{}{puuid}})
	}

	if userUuid != "" {
		wp = wp.And(&builder.WherePair{Query: "user_uuid = ?", Args: []interface{}{userUuid}})
	}

	if name != "" {
		wp = wp.And(&builder.WherePair{Query: "name = ?", Args: []interface{}{name}})
	}

	if dir == TRUE {
		wp = wp.And(&builder.WherePair{Query: "dir = ?", Args: []interface{}{true}})
	} else if dir == FALSE {
		wp = wp.And(&builder.WherePair{Query: "dir = ?", Args: []interface{}{false}})
	}

	db := core.CONTEXT.GetDB().Where(wp.Query, wp.Args...).First(matter)

	if db.Error != nil {
		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			return nil
		} else {
			this.PanicError(db.Error)
		}
	}

	return matter
}

func (this *MatterDao) FindByPuuidAndUserUuid(puuid string, userUuid string, sortArray []builder.OrderPair) []*Matter {
	return this.FindByPuuidAndUserUuidAndDeleted(puuid, userUuid, "", sortArray)
}

func (this *MatterDao) FindByPuuidAndUserUuidAndDeleted(puuid string, userUuid string, deleted string, sortArray []builder.OrderPair) []*Matter {
	var matters []*Matter

	var wp = &builder.WherePair{}
	wp = wp.And(&builder.WherePair{Query: "puuid = ? AND user_uuid = ?", Args: []interface{}{puuid, userUuid}})
	if deleted == TRUE {
		wp = wp.And(&builder.WherePair{Query: "deleted = 1", Args: []interface{}{}})
	} else if deleted == FALSE {
		wp = wp.And(&builder.WherePair{Query: "deleted = 0", Args: []interface{}{}})
	}

	if sortArray == nil {

		sortArray = []builder.OrderPair{
			{
				Key:   "dir",
				Value: DIRECTION_DESC,
			},
			{
				Key:   "create_time",
				Value: DIRECTION_DESC,
			},
		}
	}

	db := core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Order(this.GetSortString(sortArray)).Find(&matters)
	this.PanicError(db.Error)

	return matters
}

func (this *MatterDao) FindByUuids(uuids []string, sortArray []builder.OrderPair) []*Matter {
	var matters []*Matter

	db := core.CONTEXT.GetDB().Where(uuids).Order(this.GetSortString(sortArray)).Find(&matters)
	this.PanicError(db.Error)

	return matters
}
func (this *MatterDao) PlainPage(
	page int,
	pageSize int,
	puuid string,
	userUuid string,
	name string,
	dir string,
	deleted string,
	deleteTimeBefore *time.Time,
	extensions []string,
	sortArray []builder.OrderPair) (int, []*Matter) {

	var wp = &builder.WherePair{}

	if puuid != "" {
		wp = wp.And(&builder.WherePair{Query: "puuid = ?", Args: []interface{}{puuid}})
	}

	if userUuid != "" {
		wp = wp.And(&builder.WherePair{Query: "user_uuid = ?", Args: []interface{}{userUuid}})
	}

	if name != "" {
		wp = wp.And(&builder.WherePair{Query: "name LIKE ?", Args: []interface{}{"%" + name + "%"}})
	}

	if deleteTimeBefore != nil {
		wp = wp.And(&builder.WherePair{Query: "delete_time < ?", Args: []interface{}{&deleteTimeBefore}})
	}

	if dir == TRUE {
		wp = wp.And(&builder.WherePair{Query: "dir = ?", Args: []interface{}{1}})
	} else if dir == FALSE {
		wp = wp.And(&builder.WherePair{Query: "dir = ?", Args: []interface{}{0}})
	}

	if deleted == TRUE {
		wp = wp.And(&builder.WherePair{Query: "deleted = ?", Args: []interface{}{1}})
	} else if deleted == FALSE {
		wp = wp.And(&builder.WherePair{Query: "deleted = ?", Args: []interface{}{0}})
	}

	var conditionDB *gorm.DB
	if extensions != nil && len(extensions) > 0 {
		var orWp = &builder.WherePair{}

		for _, v := range extensions {
			orWp = orWp.Or(&builder.WherePair{Query: "name LIKE ?", Args: []interface{}{"%." + v}})
		}

		conditionDB = core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Where(orWp.Query, orWp.Args...)
	} else {
		conditionDB = core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...)
	}

	var count int64 = 0
	db := conditionDB.Count(&count)
	this.PanicError(db.Error)

	var matters []*Matter
	db = conditionDB.Order(this.GetSortString(sortArray)).Offset(page * pageSize).Limit(pageSize).Find(&matters)
	this.PanicError(db.Error)

	return int(count), matters
}
func (this *MatterDao) Page(page int, pageSize int, puuid string, userUuid string, name string, dir string, deleted string, extensions []string, sortArray []builder.OrderPair) *Pager {

	count, matters := this.PlainPage(page, pageSize, puuid, userUuid, name, dir, deleted, nil, extensions, sortArray)
	pager := NewPager(page, pageSize, count, matters)

	return pager
}

// handle matter page by page.
func (this *MatterDao) PageHandle(
	puuid string,
	userUuid string,
	name string,
	dir string,
	deleted string,
	deleteTimeBefore *time.Time,
	sortArray []builder.OrderPair,
	fun func(matter *Matter)) {

	pageSize := 1000
	if sortArray == nil || len(sortArray) == 0 {
		sortArray = []builder.OrderPair{
			{
				Key:   "uuid",
				Value: DIRECTION_ASC,
			},
		}
	}

	count, _ := this.PlainPage(0, pageSize, puuid, userUuid, name, dir, deleted, deleteTimeBefore, nil, sortArray)
	if count > 0 {
		var totalPages = int(math.Ceil(float64(count) / float64(pageSize)))

		var page int
		for page = 0; page < totalPages; page++ {
			_, matters := this.PlainPage(0, pageSize, puuid, userUuid, name, dir, deleted, deleteTimeBefore, nil, sortArray)
			for _, matter := range matters {
				fun(matter)
			}
		}
	}
}

func (this *MatterDao) Create(matter *Matter) *Matter {

	timeUUID, _ := uuid.NewV4()
	matter.Uuid = string(timeUUID.String())
	matter.CreateTime = time.Now()
	matter.UpdateTime = time.Now()
	matter.Sort = time.Now().UnixNano() / 1e6
	db := core.CONTEXT.GetDB().Create(matter)
	this.PanicError(db.Error)

	return matter
}

func (this *MatterDao) Save(matter *Matter) *Matter {

	matter.UpdateTime = time.Now()
	db := core.CONTEXT.GetDB().Save(matter)
	this.PanicError(db.Error)

	return matter
}

// download time add 1
func (this *MatterDao) TimesIncrement(matterUuid string) {
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where("uuid = ?", matterUuid).Updates(map[string]interface{}{"times": gorm.Expr("times + 1"), "visit_time": time.Now()})
	this.PanicError(db.Error)
}

func (this *MatterDao) SizeByPuuidAndUserUuid(matterUuid string, userUuid string) int64 {

	var wp = &builder.WherePair{Query: "puuid = ? AND user_uuid = ?", Args: []interface{}{matterUuid, userUuid}}

	var count int64
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Count(&count)
	if count == 0 {
		return 0
	}

	var sumSize int64
	db = core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Select("SUM(size)")
	this.PanicError(db.Error)
	row := db.Row()
	err := row.Scan(&sumSize)
	core.PanicError(err)

	return sumSize
}

// delete a file from db and disk.
func (this *MatterDao) Delete(matter *Matter) {

	// recursive if dir
	if matter.Dir {
		matters := this.FindByPuuidAndUserUuid(matter.Uuid, matter.UserUuid, nil)

		for _, f := range matters {
			this.Delete(f)
		}

		//delete from db.
		db := core.CONTEXT.GetDB().Delete(&matter)
		this.PanicError(db.Error)

		//delete dir from disk.
		util.DeleteEmptyDir(matter.AbsolutePath())

	} else {

		//delete from db.
		db := core.CONTEXT.GetDB().Delete(&matter)
		this.PanicError(db.Error)

		//delete its image cache.
		this.imageCacheDao.DeleteByMatterUuid(matter.Uuid)

		//delete all the share.
		this.bridgeDao.DeleteByMatterUuid(matter.Uuid)

		//delete from disk.
		err := os.Remove(matter.AbsolutePath())
		if err != nil {
			this.logger.Error("occur error when deleting file. %v", err)
		}

	}
}

// soft delete a file or dir
func (this *MatterDao) SoftDelete(matter *Matter) {

	//soft delete from db.
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where("uuid = ?", matter.Uuid).Updates(map[string]interface{}{"deleted": true, "delete_time": time.Now()})
	this.PanicError(db.Error)

}

// recovery a file
func (this *MatterDao) Recovery(matter *Matter) {

	//recovery from db.
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where("uuid = ?", matter.Uuid).Updates(map[string]interface{}{"deleted": false, "delete_time": time.Now()})
	this.PanicError(db.Error)

}

func (this *MatterDao) DeleteByUserUuid(userUuid string) {

	db := core.CONTEXT.GetDB().Where("user_uuid = ?", userUuid).Delete(Matter{})
	this.PanicError(db.Error)

}

func (this *MatterDao) CountBetweenTime(startTime time.Time, endTime time.Time) int64 {
	var count int64
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where("create_time >= ? AND create_time <= ?", startTime, endTime).Count(&count)
	this.PanicError(db.Error)
	return count
}

func (this *MatterDao) SizeBetweenTime(startTime time.Time, endTime time.Time) int64 {

	var wp = &builder.WherePair{Query: "dir = 0 AND create_time >= ? AND create_time <= ?", Args: []interface{}{startTime, endTime}}

	var count int64
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Count(&count)
	if count == 0 {
		return 0
	}

	var size int64
	db = core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Select("SUM(size)")
	this.PanicError(db.Error)
	row := db.Row()
	err := row.Scan(&size)
	this.PanicError(err)
	return size
}

// find by userUuid and path. if not found, return nil
func (this *MatterDao) findByUserUuidAndPath(userUuid string, path string) *Matter {

	var wp = &builder.WherePair{Query: "user_uuid = ? AND path = ?", Args: []interface{}{userUuid, path}}

	var matter = &Matter{}
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).First(matter)

	if db.Error != nil {
		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			return nil
		} else {
			this.PanicError(db.Error)
		}
	}

	return matter
}

// find by userUuid and path. if not found, panic
func (this *MatterDao) checkByUserUuidAndPath(userUuid string, path string) *Matter {

	if path == "" {
		panic(result.BadRequest("path cannot be null"))
	}
	matter := this.findByUserUuidAndPath(userUuid, path)
	if matter == nil {
		panic(result.NotFound("path = %s not exists", path))
	}

	return matter
}

func (this *MatterDao) SumSizeByUserUuidAndPath(userUuid string, path string) int64 {

	var wp = &builder.WherePair{Query: "user_uuid = ? AND path like ?", Args: []interface{}{userUuid, path + "%"}}

	var count int64
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Count(&count)
	if count == 0 {
		return 0
	}

	var sumSize int64
	db = core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Select("SUM(size)")
	this.PanicError(db.Error)
	row := db.Row()
	err := row.Scan(&sumSize)
	core.PanicError(err)

	return sumSize

}

func (this *MatterDao) CountByUserUuidAndPath(userUuid string, path string) int64 {

	var wp = &builder.WherePair{Query: "user_uuid = ? AND path like ?", Args: []interface{}{userUuid, path + "%"}}

	var count int64
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Count(&count)
	core.PanicError(db.Error)

	return count

}

// 统计总共有多少条。
func (this *MatterDao) Count() int64 {

	var count int64
	db := core.CONTEXT.GetDB().Model(&Matter{}).Count(&count)
	core.PanicError(db.Error)

	return count

}

// System cleanup.
func (this *MatterDao) Cleanup() {
	this.logger.Info("[MatterDao] clean up. Delete all Matter record in db and on disk.")
	db := core.CONTEXT.GetDB().Where("uuid is not null").Delete(Matter{})
	this.PanicError(db.Error)

	err := os.RemoveAll(core.CONFIG.MatterPath())
	this.PanicError(err)

}
