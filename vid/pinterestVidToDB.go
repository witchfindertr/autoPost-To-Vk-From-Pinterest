package vid

import (
	"../base"
	"../helpers"
	"../initial"
	"../models"
	"errors"
	"time"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func PinterestVidToDB() error {
	// Инициализация переменных ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	var err error
	var cur string
	var req base.DB
	var writes bool
	var media string
	var pages models.Pages
	var res  models.Pinterest
	var pinIni models.PinterestIni

	// Подключение к базе данных  ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	if err = base.Connect("autoPost.ini", "DB", &req); err == nil {
		defer func() {
			if err := req.Db.Close(); err != nil {
				err = errors.New("error req.DB.Close(): " + fmt.Sprint(err))
			}
		}()
		// Чтение переменных из файла настроек ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
		i := initial.InitS{File: "autoPost.ini", Sect: "Pinterest"}
		if err = i.InitF(&pinIni); err == nil {
			// Основной цикл запроса к API ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
			for {
				// Существует ли курсор для этой доски ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
				pages.Exists = false
				if err = req.Db.Raw("SELECT EXISTS(SELECT * FROM `curs`  WHERE `group` = '" + pinIni.BoardVid + "')  AS 'exists' ;").Find(&pages).Error; err == nil {
					// Если нет - создать записть  ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
					if !pages.Exists {
						fields := "(`group`, `cursor`)"
						if err = req.Db.Table("curs").Exec("INSERT INTO `curs` " + fields + "  VALUES ('" + pinIni.BoardVid + "', '');").Error; err != nil {
							err = errors.New("DB error INSERT INTO curs: " + fmt.Sprint(err))
							break
						}
					} else { // Если есть - считать  ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
						if err = req.Db.Raw("SELECT * FROM `curs`  WHERE `group` = '" + pinIni.BoardVid+ "';").Find(&pages).Error; err == nil {
							cur = pages.Cursor // обновить локальный курсор ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
						} else {
							err = errors.New("DB error SELECT FROM curs: " + fmt.Sprint(err))
							break
						}
					}

					// API запрос к Pinterest ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
					if request, err := http.Get(helpers.UrlEncoded("https://api.pinterest.com/v1/boards/" + pinIni.PinUser + "/" + pinIni.BoardVid + "/pins/?access_token=" + pinIni.Token + "&fields=" + models.PIN_FIELDS + "&cursor=" + cur)); err == nil {

						// Узнать количество оставшихся запросов ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
						limit := models.GetRatelimit(request)

						helpers.Add2Log("PinterestPicToDB","Limit "+fmt.Sprint(limit.Limit))
						helpers.Add2Log("PinterestPicToDB","Refresh "+fmt.Sprint(limit.Refresh))
						helpers.Add2Log("PinterestPicToDB","Remaining "+fmt.Sprint(limit.Remaining))

						// выход из цикла если лимит исчерпан  ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
						if limit.Remaining == 0 {
							helpers.Add2Log("limit: ", "limit break")
							break
						}
						helpers.Add2Log("limit: ", fmt.Sprint(limit.Remaining))

						writes = false
						// Считать тело возвращенного ответа ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
						if jss, err := ioutil.ReadAll(request.Body); err == nil {
							// Распарсить ответ в структуру ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
							if err := json.Unmarshal([]byte(jss), &res); err == nil {

								if limit.Remaining != 0 {
									cur = res.Page.Cursor
									l := len(res.Data)

									// если ответ не пустой - запустить цикл для записи значений в базу  ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
									for i := 0; i < l; i++ {

										firstPin := res.Data[i]
										//createtime := res.Data[i].CreatedAt
										idPin := firstPin.ID
										note := helpers.VideoID(firstPin.OriginalLink)
										orLink := "https://www.youtube.com/watch?v=" + note
										media = firstPin.Media.Type
										link := firstPin.Image.Original.URL
										h := fmt.Sprint(firstPin.Image.Original.Height)
										w := fmt.Sprint(firstPin.Image.Original.Width)

										// опеределить тип видео ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
										if strings.Contains(firstPin.OriginalLink, "youtube") {
											media = "youtube"
										} else {
											if strings.Contains(firstPin.OriginalLink, "vimeo") {
												media = "vimeo"
											}
										}

										// Если в базе уже есть такой код видео - не записывать ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
										pages.Exists = false
										if err = req.Db.Raw("SELECT EXISTS(SELECT * FROM `pinTut` WHERE `note` = '" + note + "')  AS 'exists' ;").Find(&pages).Error; err == nil {
											// записать только видео с youtube и vimeo
											// TODO добавить обработку видео с vimeo || media == "vimeo"
											if !pages.Exists && (media == "youtube") {
												time.Sleep(500 * time.Millisecond) // полусекундная задержка
												//timestamp := strconv.FormatInt(createtime.UTC().UnixNano(), 10)
												helpers.Add2Log("write: ", orLink)
												writes = true
												fields := "(`id`, `media`, `link`, `H`, `W`, `originlink`, `note`, `public`, `idPin`)"
												if err = req.Db.Table("pinTut").Exec("INSERT INTO `pinTut` " + fields + "  VALUES (NULL, '" + media + "', '" + link + "', " + h + ", " + w + ", '" + orLink + "', '" + note + "', 0, " + idPin + ");").Error; err != nil {
													err = errors.New("DB error INSERT video to db: " + fmt.Sprint(err))
												}
											} else {
												helpers.Add2Log("don`t write: ", orLink)
											}
										} else {
											err = errors.New("DB error EXISTS pinTut: " + fmt.Sprint(err))
											break
										}

									}
								}

								// Управление курсором для текущей доски ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
								if !writes {
									// если записи не производилось - увеличить переменную writes на 1 ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
									req.Db.Table("curs").Exec("UPDATE `curs` SET `writes` = `writes` + 1 WHERE `curs`.`group` = '" + pinIni.BoardVid + "';")
									req.Db.Raw("SELECT * FROM `curs`  WHERE `group` = '?';", pinIni.BoardVid).Find(&pages)
								} else {
									// если запись произвелась - сбросить переменную writes на 0 ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
									req.Db.Table("curs").Exec("UPDATE `curs` SET `writes` = 0 WHERE `curs`.`group` = '" + pinIni.BoardVid + "';" )
								}

								if pages.Writes >= pinIni.Iter {
									// если счетчик курсора страниц больше либо равно установленного в файле настройки - обнулить курсор - выйти из цикла ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
									req.Db.Table("curs").Exec("UPDATE `curs` SET `cursor` = '' WHERE `curs`.`group` = '" + pinIni.BoardVid + "';")
									break
								} else {
									// иначе обновить сам курсор в базе
									req.Db.Table("curs").Exec("UPDATE `curs` SET `cursor` = '" + cur + "' WHERE `curs`.`group` = '" + pinIni.BoardVid + "';")
								}

								// если курсор в ответе на запрос пуст а курсор в таблице заполнен или оба пусты - выйти из цикла
								if (pages.Cursor != "" && res.Page.Cursor == "") || (pages.Cursor == "" && res.Page.Cursor == "") {
									req.Db.Table("curs").Exec("UPDATE `curs` SET `writes` = 0 WHERE `curs`.`group` = '" + pinIni.BoardVid + "';")
									break
								}
								request.Body.Close()

							} else {
								err = errors.New("error JSON unmarshaling failed: " + fmt.Sprint(err))
								break
							}
						} else {
							err = errors.New("error ioutil.ReadAll: " + fmt.Sprint(err))
							break
						}
					} else {
						err = errors.New("fail requests " + fmt.Sprint(err))
						break
					}
				} else {
					err = errors.New("DB error EXISTS curs: " + fmt.Sprint(err))
					break
				}
			} // end for
		} else {
			err = errors.New("fail to read ini: " + fmt.Sprint(err))
		}
	} else {
		err = errors.New("fail base connect: " + fmt.Sprint(err))
	}

return  err
}
