package pic

import (
	"../base"
	"../helpers"
	"../initial"
	"../models"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

func  PostPicToVk() error {

	// Инициализация переменных ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	v := "5.101"
	var s string
	var err error
	var tags []string
	var photoId string
	var req = base.DB{}
	var empty = models.Empty{}
	var postPic = models.PostPic{}
	var vkPicIni = models.VkPicIni{}
	var getUrl = models.VkGetWallUploadS{}
	var saveWall = models.VkSaveWallPhoto{}

	// Подключение к базе данных  ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	if err = base.Connect("autoPost.ini", "DB", &req); err == nil {
		defer func() {
			if err = req.Db.Close(); err != nil {
				err = errors.New("error req.DB.Close(): " + fmt.Sprint(err))
			}
		}()
		// Чтение переменных из файла ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
		i := initial.InitS{File: "autoPost.ini", Sect: "VkPic"}
		if err = i.InitF(&vkPicIni); err == nil {

			if vkPicIni.VkHashTags != "" {
				tags = strings.Split(vkPicIni.VkHashTags, ",")
			}

			arr := helpers.RangeInt(0, len(tags), vkPicIni.MaxHashTags)

			for r := 0; r < len(arr); r++ {
				s = s + " #" + fmt.Sprint(tags[arr[r]])
			}

			if err = req.Db.Raw("SELECT t.id,t.originlink FROM pintArtPic t WHERE t.public = 1 LIMIT 1").Find(&postPic).Error; err == nil {

				if err = req.Db.Table("pintArtPic").Exec("UPDATE pintArtPic SET public = 2 WHERE id = " + fmt.Sprint(postPic.Id) + ";").Error; err == nil {
					// ~~~~ Пост картинки с описанием на стену группы ~~~~
					// получить ссылку по которой можно загрузить картинку на сайт ВК
					paramsWallPhotosGetUploadServer := map[string]string{
						"group_id":     vkPicIni.VkGroupId,
						"access_token": vkPicIni.VkToken,
						"v":            v,
					}
					if respon, err := helpers.Request("v", "photos.getWallUploadServer", paramsWallPhotosGetUploadServer, &getUrl); err == nil {
						// путь к файлу
						path := fmt.Sprint(postPic.Id) + ".jpg"

						// подготовительный этап в формировании загрузки на сайт
						if uploaded, err := helpers.PhotoWall(getUrl.Response.UploadUrl, path); err == nil {

							// сохранить картинку для поста на стене группы
							paramsSaveWallPhoto := map[string]string{
								"group_id":     vkPicIni.VkGroupId,
								"photo":        uploaded.Photo,
								"hash":         uploaded.Hash,
								"server":       fmt.Sprint(uploaded.Server),
								"access_token": vkPicIni.VkToken,
								"v":            v,
							}
							if respon, err = helpers.Request("v", "photos.saveWallPhoto", paramsSaveWallPhoto, &saveWall); err == nil {
								if len(saveWall.Response) > 0 {
									// формируем правильное название для загружаемого файла
									photoId = "photo" + fmt.Sprint(saveWall.Response[0].OwnerId) + "_" + fmt.Sprint(saveWall.Response[0].Id)

									//// Рандомная временная задержка Перед постом~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
									rand.Seed(time.Now().UnixNano())
									duration := time.Duration(rand.Intn(vkPicIni.PreTimeOut)) * time.Second // Pause for 10 seconds
									time.Sleep(duration)

									//пост картинки с описанием в группу
									paramsWallPost := map[string]string{
										"owner_id":     "-" + vkPicIni.VkGroupId,
										"from_group":   "1",
										"message":      "Источник: " + postPic.Originlink + "\n\r\n\r" + strings.Trim(s, " "),
										"attachments":  photoId,
										"access_token": vkPicIni.VkToken,
										"v":            v,
									}
									if respon, err = helpers.Request("v", "wall.post", paramsWallPost, &empty); err == nil {
										time.Sleep(120 * time.Second) // задержка перед сохранением картинки в альбом

										// ~~~~ Пост картинки с описанием в альбом группы ~~~~
										// получить ссылку по которой можно загрузить картинку на сайт ВК
										paramsPhotosGetUploadServer := map[string]string{
											"group_id":     vkPicIni.VkGroupId,
											"album_id":     vkPicIni.VkAlbumId,
											"access_token": vkPicIni.VkToken,
											"v":            v,
										}
										if respon, err = helpers.Request("v", "photos.getUploadServer", paramsPhotosGetUploadServer, &getUrl); err == nil {
											// загрузить картинку по ссылке полученной ранее
											if uploadPhoto, err := helpers.PhotoGroup(getUrl.Response.UploadUrl, path); err == nil {
												// сохранить картинку в альбом
												paramsPhotosSave := map[string]string{
													"group_id":     vkPicIni.VkGroupId,
													"album_id":     vkPicIni.VkAlbumId,
													"photos_list":  uploadPhoto.PhotoList,
													"server":       fmt.Sprint(uploadPhoto.Server),
													"hash":         uploadPhoto.Hash,
													"caption":      "Источник: " + postPic.Originlink,
													"access_token": vkPicIni.VkToken,
													"v":            v,
												}
												if respon, err = helpers.Request("v", "photos.save", paramsPhotosSave, &empty); err == nil {
													// удалить файл
													if err = os.Remove(path); err != nil {
														err = errors.New("error Remove file: " + fmt.Sprint(err))
													}
												} else {
													err = errors.New("photos.save failed: " + fmt.Sprint(err) + " ; response: " + fmt.Sprint(respon))
												}
											} else {
												err = errors.New("uploaded to album failed: " + fmt.Sprint(err))
											}
										} else {
											err = errors.New("photos.getUploadServer failed: " + fmt.Sprint(err) + " ; response: " + fmt.Sprint(respon))
										}
									} else {
										err = errors.New("wall.post failed: " + fmt.Sprint(err) + " ; \n response: " + fmt.Sprint(respon))
									}
								} else {
									err = errors.New("method saveWallPhoto failed: Response index = 0" + fmt.Sprint(err))
								}
							} else {
								err = errors.New("photos.saveWallPhoto failed: " + fmt.Sprint(err) + " ; \n response: " + fmt.Sprint(respon))
							}
						} else {
							err = errors.New("error uploaded photoWall: " + fmt.Sprint(err))
						}
					} else {
						err = errors.New("photos.getWallUploadServer failed: " + fmt.Sprint(respon) + "; response: " + fmt.Sprint(err))
					}
				} else {
					err = errors.New("DB error UPDATE pintArtPic SET public = 2: " + fmt.Sprint(err))
				}
			} else {
				err = errors.New("DB error SELECT FROM pintArtPic: " + fmt.Sprint(err))
			}
		} else {
			err = errors.New("error read ini file: " + fmt.Sprint(err))
		}
	} else {
		err = errors.New("fail connect to DB: " + fmt.Sprint(err))
	}

	return err
}
