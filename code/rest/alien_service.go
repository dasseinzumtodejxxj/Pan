package rest

import (
	"fmt"
	"tankmaster/code/core"
	"tankmaster/code/tool/result"

	"net/http"
	"time"
)

// @Service
type AlienService struct {
	BaseBean
	matterDao         *MatterDao
	matterService     *MatterService
	userDao           *UserDao
	uploadTokenDao    *UploadTokenDao
	downloadTokenDao  *DownloadTokenDao
	shareService      *ShareService
	imageCacheDao     *ImageCacheDao
	imageCacheService *ImageCacheService
}

func (this *AlienService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}

	b = core.CONTEXT.GetBean(this.matterService)
	if b, ok := b.(*MatterService); ok {
		this.matterService = b
	}

	b = core.CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

	b = core.CONTEXT.GetBean(this.uploadTokenDao)
	if c, ok := b.(*UploadTokenDao); ok {
		this.uploadTokenDao = c
	}

	b = core.CONTEXT.GetBean(this.downloadTokenDao)
	if c, ok := b.(*DownloadTokenDao); ok {
		this.downloadTokenDao = c
	}

	b = core.CONTEXT.GetBean(this.shareService)
	if c, ok := b.(*ShareService); ok {
		this.shareService = c
	}

	b = core.CONTEXT.GetBean(this.imageCacheDao)
	if c, ok := b.(*ImageCacheDao); ok {
		this.imageCacheDao = c
	}

	b = core.CONTEXT.GetBean(this.imageCacheService)
	if c, ok := b.(*ImageCacheService); ok {
		this.imageCacheService = c
	}
}

func (this *AlienService) PreviewOrDownload(
	writer http.ResponseWriter,
	request *http.Request,
	uuid string,
	filename string,
	withContentDisposition bool) {

	matter := this.matterDao.CheckByUuid(uuid)

	if matter.Name != filename {
		panic(result.BadRequest("filename in url incorrect"))
	}

	//only private file need auth.
	if matter.Privacy {

		//1.use downloadToken to auth.
		downloadTokenUuid := request.FormValue("downloadTokenUuid")
		if downloadTokenUuid != "" {
			downloadToken := this.downloadTokenDao.CheckByUuid(downloadTokenUuid)
			if downloadToken.ExpireTime.Before(time.Now()) {
				panic(result.BadRequest("downloadToken has expired"))
			}

			if downloadToken.MatterUuid != uuid {
				panic(result.BadRequest("token and file info not match"))
			}

			tokenUser := this.userDao.CheckByUuid(downloadToken.UserUuid)
			if matter.UserUuid != tokenUser.Uuid {
				panic(result.UNAUTHORIZED)
			}

			//TODO: expire the download token. If download by chunk, do this later.
			downloadToken.ExpireTime = time.Now()
			this.downloadTokenDao.Save(downloadToken)

		} else {

			operator := this.findUser(request)

			//use share code to auth.
			shareUuid := request.FormValue("shareUuid")
			shareCode := request.FormValue("shareCode")
			shareRootUuid := request.FormValue("shareRootUuid")

			this.shareService.ValidateMatter(request, shareUuid, shareCode, operator, shareRootUuid, matter)

		}
	}

	//download directory
	if matter.Dir {

		this.matterService.DownloadZip(writer, request, []*Matter{matter})

	} else {

		//handle the image operation.
		needProcess, imageResizeM, imageResizeW, imageResizeH := this.imageCacheService.ResizeParams(request)
		if needProcess {

			//if image, try to use cache.
			mode := fmt.Sprintf("%s_%d_%d", imageResizeM, imageResizeW, imageResizeH)
			imageCache := this.imageCacheDao.FindByMatterUuidAndMode(matter.Uuid, mode)
			if imageCache == nil {
				imageCache = this.imageCacheService.cacheImage(writer, request, matter)
			}

			//download the cache image file.
			this.matterService.DownloadFile(writer, request, GetUserCacheRootDir(imageCache.Username)+imageCache.Path, imageCache.Name, withContentDisposition)

		} else {
			this.matterService.DownloadFile(writer, request, matter.AbsolutePath(), matter.Name, withContentDisposition)
		}

	}

	//async increase the download times.
	go core.RunWithRecovery(func() {
		this.matterDao.TimesIncrement(uuid)
	})

}
