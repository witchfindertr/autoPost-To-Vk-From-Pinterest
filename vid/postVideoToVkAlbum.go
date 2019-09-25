package vid

import (
	"../base"
	"../helpers"
	"../initial"
	"../models"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)


func PostVideoToVkAlbum() error {

	v := "5.101"
	var err error
	var req 	base.DB
	var iniVk 	models.IniVK
	var iniYT 	models.IniYouTube
	var resYT 	models.YouTubeResp
	var dbVidInf models.DbVideoInfo
	var vkUploadUrl	models.UploadUrlVK

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	// Подключение к базе данных  ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	if err = base.Connect("autoPost.ini", "DB", &req); err == nil {
		helpers.Add2Log("PostVideoToVkAlbum", "подключение к базе ок")
		defer func() {
			if err := req.Db.Close(); err != nil {
				err = errors.New("error req.DB.Close(): " + fmt.Sprint(err))
			}
		}()
		// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
		// Чтение настроек ВК ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
		vk := initial.InitS{File: "autoPost.ini", Sect: "VkVid"}
		if err = vk.InitF(&iniVk); err == nil {
			// Чтение настроек Ютуб ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
			yt := initial.InitS{File: "autoPost.ini", Sect: "YouTube"}
			if err = yt.InitF(&iniYT); err == nil {
				helpers.Add2Log("PostVideoToVkAlbum", "чтение настроек ок")
				// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
				// Выборка рандомной записи из таблицы ~~~~~~~~~~~~~~~~~~~~~~~~~~
				if err = req.Db.Raw("SELECT t.id,t.note,t.originlink " +
					"FROM admin_autoPost.pinTut t " +
					"WHERE t.media = 'youtube' " +
					"AND t.public = 0 " +
					"ORDER BY RAND() LIMIT 1").Find(&dbVidInf).Error; err == nil {
					helpers.Add2Log("PostVideoToVkAlbum", "выбор видео из базы ок: " + fmt.Sprint(dbVidInf.Id))
					// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
					// Изменение признака публикации ~~~~~~~~~~~~~~~~~~~~~~~~~~
					if err = req.Db.Table("pinTut").Exec("UPDATE pinTut SET public = 2 WHERE id = '" + fmt.Sprint(dbVidInf.Id) + "';" ).Error; err == nil {
						helpers.Add2Log("PostVideoToVkAlbum", "апдейт признака в базе ок")
						// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
						// Рандомная временная задержка Перед постом ~~~~~~~~~~~~~~~~~~~~
						rand.Seed(time.Now().UnixNano())
						duration := time.Duration(rand.Intn(iniVk.PreTimeOut)) * time.Second // Pause for rnd seconds
						time.Sleep(duration)
						// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
						// Запрос к API YouTube snippet ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
						paramsYTVideos := map[string]string{
							"part": "snippet",
							"id":   dbVidInf.Note,
							"key":  iniYT.YouTubeApiKey,
						}
						if responseYT, err := helpers.Request("y", "videos", paramsYTVideos, &resYT); err == nil {
							duration := 1 * time.Second // Pause for 1 seconds
							time.Sleep(duration)
							helpers.Add2Log("PostVideoToVkAlbum", "запрос youtube videos ок: " + helpers.ByteToString(responseYT))
							// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
							// Запрос к API VK video.save ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
							helpers.Add2Log("VkGroupId",  iniVk.VkGroupId)
							helpers.Add2Log("OriginLink",  dbVidInf.Originlink)
							helpers.Add2Log("VkToken",  iniVk.VkToken)
							paramsVideoSave := map[string]string{
								"group_id":     iniVk.VkGroupId,
								"link":         dbVidInf.Originlink,
								"name":         resYT.Items[0].Snippet.Title,
								"description":  resYT.Items[0].Snippet.Title,
								"wallpost":     "0",
								"access_token": iniVk.VkToken,
								"v":            v,
							}
							if responseVkSave, err := helpers.Request("v", "video.save", paramsVideoSave, &vkUploadUrl); err == nil {
								time.Sleep(duration)
								fmt.Println("url: " + vkUploadUrl.Response.UploadUrl)
								if vkUploadUrl.Error.ErrorCode == 0 {
									helpers.Add2Log("PostVideoToVkAlbum", "запрос в вк video.save ок: " + helpers.ByteToString(responseVkSave))
									// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
									// Переход по ссылке подтверждения пуликации  ~~~~~~~~~~~~~~~~~~~~
									if req, err := http.Get(vkUploadUrl.Response.UploadUrl); err == nil {
										if req.StatusCode == 200 {
											helpers.Add2Log("PostVideoToVkAlbum", "Get(uploadUrlVk) Переход по ссылке подтверждения пуликации ок: " + req.Status )
										} else {
											helpers.Add2Log("PostVideoToVkAlbum", "Get(uploadUrlVk) Ошибка перехода по ссылке. Статус: " + req.Status)
										}

										defer func() {
											if err := req.Body.Close(); err != nil {
												err = errors.New("error req.Body.Close(): " + fmt.Sprint(err) )
											}
										}()
									} else {
										err = errors.New("error  Get(uploadUrlVk): " + fmt.Sprint(err))
									}
								} else {
									err = errors.New("error Request video.save: " + fmt.Sprint(vkUploadUrl.Error.ErrorMsg))
								}
							} else {
								err = errors.New("error Request video.save failed: " + fmt.Sprint(err) + " ; response: " + fmt.Sprint(responseVkSave))
							}
						} else {
							err = errors.New("error Request YT videos failed: " + fmt.Sprint(err) + " ; response: " + fmt.Sprint(responseYT))
						}
					} else {
						err = errors.New("DB error UPDATE pinTut SET public = 2: " + fmt.Sprint(err))
					}
				} else {
					err = errors.New("DB error SELECT admin_autoPost.pinTut RAND: " + fmt.Sprint(err))
				}
			} else {
				err = errors.New("error read iniYT file: " + fmt.Sprint(err))
			}
		} else {
			err = errors.New("error read iniVK file: " + fmt.Sprint(err))
		}
	} else {
		err = errors.New("error base connect: " + fmt.Sprint(err))
	}
return err
}
