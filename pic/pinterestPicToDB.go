package pic

import (
	"../base"
	"../helpers"
	"../initial"
	"../models"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func PinterestPicToDB() error {

	// Инициализация переменных ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	var err error
	var cur string
	var writes bool
	var req = base.DB{}
	var page = models.Pages{}
	var res = models.Pinterest{}
	var ini = models.PinterestIni{}

	if err = base.Connect("autoPost.ini", "DB", &req); err == nil {
		defer func() {
			if err = req.Db.Close(); err != nil {
				err = errors.New("error req.DB.Close(): " + fmt.Sprint(err))
			}
		}()
		// Чтение переменных из файла настроек ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
		i := initial.InitS{File: "autoPost.ini", Sect: "Pinterest"}
		if err = i.InitF(&ini); err == nil {

			// Основной цикл запроса к API ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
			for {
				// Существует ли курсор для этой доски ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
				page.Exists = false // TODO удалить если ни на что не влияет
				if err = req.Db.Raw("SELECT EXISTS(SELECT * FROM `curs`  WHERE `group` = '" + ini.BoardPic + "')  AS 'exists' ;").Find(&page).Error; err == nil {

					if !page.Exists {// Если нет - создать записть  ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
						if err = req.Db.Table("curs").Exec("INSERT INTO `curs` (`group`, `cursor`)  VALUES ('" + ini.BoardPic + "', '');").Error; err != nil {
							err = errors.New("DB error INSERT INTO curs: " + fmt.Sprint(err))
							break
						}
					} else { // Если есть - считать  ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
						if err = req.Db.Raw("SELECT * FROM `curs`  WHERE `group` = '" + ini.BoardPic + "';").Find(&page).Error; err == nil {
							cur = page.Cursor // обновить локальный курсор ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
							helpers.Add2Log("PinterestPicToDB","cursor exists")
							helpers.Add2Log("PinterestPicToDB","cursor "+cur)
						} else {
							err = errors.New("DB error SELECT FROM curs: " + fmt.Sprint(err))
							break
						}
					}
				} else {
					err = errors.New("DB error SELECT EXISTS curs: " + fmt.Sprint(err))
					break
				}


				// API запрос к Pinterest ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
				if requests, err := http.Get(helpers.UrlEncoded("https://api.pinterest.com/v1/boards/" + ini.PinUser + "/" + ini.BoardPic + "/pins/?access_token=" + ini.Token + "&fields=" + models.PIN_FIELDS + "&cursor=" + cur)); err == nil {

					// Узнать количество оставшихся запросов ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
					limit := models.GetRatelimit(requests)

					helpers.Add2Log("PinterestPicToDB","Limit "+fmt.Sprint(limit.Limit))
					helpers.Add2Log("PinterestPicToDB","Refresh "+fmt.Sprint(limit.Refresh))
					helpers.Add2Log("PinterestPicToDB","Remaining "+fmt.Sprint(limit.Remaining))

					// выход из цикла если лимит исчерпан  ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
					if limit.Remaining == 0 {
						helpers.Add2Log("PinterestPicToDB","limit break")
						break
					}

					helpers.Add2Log("PinterestPicToDB","limit: " + fmt.Sprint(limit.Remaining))
					writes = false

					// Считать тело возвращенного ответа ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
					if jss, err := ioutil.ReadAll(requests.Body); err == nil {
						// Распарсить ответ в структуру ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
						if err = json.Unmarshal([]byte(jss), &res); err == nil {
							if limit.Remaining != 0 {

								cur = res.Page.Cursor

								for i := 0; i < len(res.Data); i++ {
									firstPin := res.Data[i]

									idPin := firstPin.ID
									//note := firstPin.Note
									note := ""
									orLink := firstPin.OriginalLink
									media := firstPin.Media.Type
									link := firstPin.Image.Original.URL
									//createAt := firstPin.CreatedAt
									h := fmt.Sprint(firstPin.Image.Original.Height)
									w := fmt.Sprint(firstPin.Image.Original.Width)

									//layout := "2006-01-02T15:04:05"
									//t, err := time.Parse(layout, createAt)
									//if err != nil {
									//	fmt.Println(err)
									//}

									page.Exists = false // TODO удалить если ни на что не влияет
									if len(orLink) > 0 {
										if err = req.Db.Raw("SELECT EXISTS(SELECT * FROM `pintArtPic` " +
											"WHERE `originlink` = '" + orLink + "')  AS 'exists' ;").Find(&page).Error; err == nil {

											if !page.Exists && len(orLink) > 10 {
												time.Sleep(1000 * time.Millisecond) // секундная задержка
												helpers.Add2Log("PinterestPicToDB","write")
												writes = true
												fields := "(`id`, `media`, `link`, `H`, `W`, `originlink`, `note`, `public`, `idPin`)"
												if err = req.Db.Table("pintArtPic").Exec("INSERT INTO `pintArtPic` " + fields + "  VALUES (NULL, '" + media + "', '" + link + "', " + h + ", " + w + ", '" + orLink + "', '" + note + "', 0, " + idPin + ");").Error; err != nil {
													err = errors.New("DB error INSERT INTO pintArtPic: " + fmt.Sprint(err))
													break
												}
											} else {
												helpers.Add2Log("PinterestPicToDB","skip write")
											}
										} else {
											err = errors.New("DB error SELECT EXISTS pintArtPic: " + fmt.Sprint(err))
											break
										}
									}
								} // end for ~~~~~~~~~~~~~~
							}
						} else {
							err = errors.New("error JSON unmarshal failed: " + fmt.Sprint(err))
							break
						}
					} else {
						err = errors.New("error ReadAll(requests.Body): " + fmt.Sprint(err))
						break
					}
				} else {
					err = errors.New("fail requests v1/boards: " + fmt.Sprint(err))
					break
				}



				if !writes {
					if err = req.Db.Table("curs").Exec("UPDATE `curs` SET `writes` = `writes` + 1 WHERE `curs`.`group` = '" + ini.BoardPic + "';").Error; err == nil {
						if err = req.Db.Raw("SELECT * FROM `curs`  WHERE `group` = '" + ini.BoardPic + "';").Find(&page).Error; err != nil {
							err = errors.New("DB error SELECT FROM curs: " + fmt.Sprint(err))
						}
					} else {
						err = errors.New("DB error UPDATE curs SET writes = writes + 1: " + fmt.Sprint(err))
						break
					}
				} else {
					if err = req.Db.Table("curs").Exec("UPDATE `curs` SET `writes` = 0 WHERE `curs`.`group` = '" + ini.BoardPic + "';").Error; err != nil {
						err = errors.New("DB error UPDATE curs SET writes = 0: " + fmt.Sprint(err))
						break
					}
				}

				if page.Writes >= ini.Iter { // update for writes > 5 and break
					if err = req.Db.Table("curs").Exec("UPDATE `curs` SET `cursor` = '' WHERE `curs`.`group` = '" + ini.BoardPic + "';").Error; err == nil {
						helpers.Add2Log("PinterestPicToDB","update for writes >5 and break")
						break
					} else {
						err = errors.New("DB error UPDATE curs SET cursor: " + fmt.Sprint(err))
						break
					}
				} else { // update
					if err = req.Db.Table("curs").Exec("UPDATE `curs` SET `cursor` = '" + cur + "' WHERE `curs`.`group` = '" + ini.BoardPic + "';").Error; err == nil {
						helpers.Add2Log("PinterestPicToDB","cursor update")
					} else {
						err = errors.New("DB error UPDATE curs SET cursor: " + fmt.Sprint(err))
						break
					}
				}

				if (page.Cursor != "" && res.Page.Cursor == "") || (page.Cursor == "" && res.Page.Cursor == "") {
					// cursor break
					if err = req.Db.Table("curs").Exec("UPDATE `curs` SET `writes` = 0 WHERE `curs`.`group` = '" + ini.BoardPic + "';").Error; err == nil {
						helpers.Add2Log("PinterestPicToDB","cursor break")
						break
					} else {
						err = errors.New("DB error UPDATE curs SET writes = 0: " + fmt.Sprint(err))
						break
					}
				}
			} // end for
		} else {
			err = errors.New("fail to read ini: " + fmt.Sprint(err))
		}
	} else {
		err = errors.New("fail connect to DB: " + fmt.Sprint(err))
	}
	return err
}
