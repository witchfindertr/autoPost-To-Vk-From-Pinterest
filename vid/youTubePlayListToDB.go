package vid

import (
	"../base"
	"../helpers"
	"../initial"
	"../models"
	"errors"
	"fmt"
	"strconv"
	"time"
)

func YouTubePlayListToDB() error {

	var brk bool
	var err error
	var next string
	var req  base.DB
	var page  models.Pages
	var iniYT models.IniYouTube
	var resYT models.YouTubeResp

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	// Чтение переменных из файла ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	i := initial.InitS{File: "autoPost.ini", Sect: "YouTube"}
	if err = i.InitF(&iniYT); err == nil {

		// Подключение к базе данных  ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
		if err = base.Connect("autoPost.ini", "DB", &req); err == nil {
			defer func() {
				if err := req.Db.Close(); err != nil {
					err = errors.New("error req.DB.Close(): " + fmt.Sprint(err))
				}
			}()
			if err = req.Db.Raw("SELECT `cursor` FROM `curs` WHERE `id` = 9;").Find(&page).Error; err == nil {
				next = page.Cursor
				for {
					paramsYTVideos := map[string]string{
						"part":       "snippet",
						"playlistId": iniYT.PlayListId,
						"pageToken":  next,
						"key":        iniYT.YouTubeApiKey,
					}
					// Запрос к youtube: список видео в альбоме  ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
					if response, err := helpers.Request("y", "playlistItems", paramsYTVideos, &resYT); err == nil {

						for i := len(resYT.Items) - 1; i > -1; i-- {
							//log.Printf("iter %d", i)
							helpers.Add2Log("youtubePlaylistToDB", fmt.Sprint(resYT.Items[i].Snippet.Position))
							// полусекундная задержка ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
							time.Sleep(800 * time.Millisecond)
							// Если в базе уже есть такой код видео - не записывать ~~~~~~~~~~~~~~~~~~~~~~~~~~~~
							page.Exists = false
							if err = req.Db.Raw("SELECT EXISTS(SELECT * FROM `pinTut` " +
								"WHERE `link` = '" + resYT.Items[i].ID + "')  " +
								"AS 'exists' ;").Find(&page).Error; err == nil {

								// Если такого видео в базе нет  ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
								if !page.Exists {
									media := "youtube"
									link := resYT.Items[i].ID
									h := "1920"
									w := "1080"
									note := resYT.Items[i].Snippet.ResourceID.VideoID
									orLink := "https://www.youtube.com/watch?v=" + note
									idPin := "0"
									timestamp := strconv.FormatInt(resYT.Items[i].Snippet.PublishedAt.UTC().UnixNano(), 10)

									fields := "(`id`, `media`, `link`, `H`, `W`, `originlink`, `note`, `public`, `idPin`, `timestamp`)"
									if err = req.Db.Table("pinTut").Exec("INSERT INTO `pinTut` " + fields + "  VALUES (NULL, '" + media + "', '" + link + "', " + h + ", " + w + ", '" + orLink + "', '" + note + "', 0, " + idPin + ",'" + timestamp + "');").Error; err == nil {
										helpers.Add2Log("youtubePlaylistToDB", "new video insert to DB - ", "success")
									} else {
										err = errors.New("DB error INSERT INTO pinTut: " + fmt.Sprint(err))
									}
								} else {
									helpers.Add2Log("youtubePlaylistToDB", "Exists")
									helpers.Add2Log("youtubePlaylistToDB", "Snippet.Position", fmt.Sprint(resYT.Items[i].Snippet.Position))

									if resYT.Items[i].Snippet.Position == resYT.PageInfo.TotalResults-1 {
										helpers.Add2Log("youtubePlaylistToDB", "end")
										brk = true
									}
									//break
								}
							} else {
								err = errors.New("DB error SELECT EXISTS videoID: " + fmt.Sprint(err))
								break
							}
						} // end for ~~
						if brk {
							break
						}
						if resYT.NextPageToken != "" {
							next = resYT.NextPageToken
							if err = req.Db.Exec("UPDATE `admin_autoPost`.`curs` t SET t.`cursor` = '" + resYT.NextPageToken + "' WHERE t.`id` = 9").Error; err == nil {
								helpers.Add2Log("youtubePlaylistToDB", "UPDATE", "cursor page success")
							} else {
								err = errors.New("DB error UPDATE cursor page: " + fmt.Sprint(err))
								break
							}
						}
					} else {
						err = errors.New("playlistItems failed: " + fmt.Sprint(err) + " ; response " + fmt.Sprint(response) + ": ")
						break
					}
				} // end for ~~
			} else {
				err = errors.New("DB error Read cursor page: " + fmt.Sprint(err))
			}
		} else {
			err = errors.New("fail base connect: " + err.Error())
		}
	} else {
		err = errors.New("fail to read ini: " + err.Error())
	}
	if err != nil {
		helpers.Add2Log("youtubePlaylistToDB", "ERR", fmt.Sprint(err))
	}
	return err
}
