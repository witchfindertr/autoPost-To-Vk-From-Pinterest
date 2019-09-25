package pic

import (
	"../base"
	"../helpers"
	"../models"
	"errors"
	"fmt"
)

func  DownloadPic() error {

	var err error
	var req 	base.DB
	var im models.ImageDownLoad

	// Подключение к базе данных  ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	if err = base.Connect("autoPost.ini", "DB", &req); err == nil {
		defer func() {
			if err := req.Db.Close(); err != nil {
				err = errors.New("error req.DB.Close(): " + fmt.Sprint(err))
			}
		}()
		// Выборка рандомной ссылки из списка в базе данных  ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
		if err = req.Db.Raw("SELECT t.id,link,originLink " +
			"FROM pintArtPic t WHERE t.public = 0 " +
			"AND t.originLink != '' " +
			"AND (t.H*t.W) < 6000000 " +
			"AND t.link LIKE '%.jpg' ORDER BY RAND() LIMIT 1;").Find(&im).Error;  err == nil {
			// Пометить строку, чтобы не выбрать ее повторно  ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
			if err = req.Db.Table("pintArtPic").Exec("UPDATE `pintArtPic` t " +
				"SET t.`public` = 1 " +
				"WHERE t.`id` = '" + fmt.Sprint(im.Id) + "';").Error;  err == nil {

				// Скачать картинку и положить в текущий каталог  ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
				if err = helpers.DownloadFile(fmt.Sprint(im.Id) + ".jpg", im.Link); err != nil {
					err = errors.New("failed to Download File: " + fmt.Sprint(err))
				}
			} else {
				err = errors.New("DB error Update public = 1: " + fmt.Sprint(err))
			}
		} else {
			err = errors.New("DB error rnd Select: " + fmt.Sprint(err))
		}
	} else {
		err = errors.New("fail base connect: " +fmt.Sprint(err))
	}

	return err
}
