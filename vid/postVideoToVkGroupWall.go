package vid

import (
	"../helpers"
	"../initial"
	"../models"
	"errors"
	"fmt"
	"time"
)

func  PostVideoToVkGroupWall() error {

	v := "5.101"
	var err error
	var iniVk models.IniVK
	var wallPost models.WallPost
	var videoGet models.VkVideoGet
	//var videoGetErr models.VkVideoGetErr

		// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
		// Чтение переменных из файла ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
		i := initial.InitS{File: "autoPost.ini", Sect: "VkVid"}
		if err = i.InitF(&iniVk); err == nil {

			// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
			// Запрос к API VK  video.get ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
			paramsVideoGet := map[string]string{
				"owner_id":     	"-" + iniVk.VkGroupId,
				"count":   		"1",
				"offset":         "0",
				"access_token": 	iniVk.VkToken,
				"v":            	v,
			}
			if response, err := helpers.Request("v", "video.get", paramsVideoGet, &videoGet); err == nil {
				// Pause for 1 seconds
				duration := 1 * time.Second
				time.Sleep(duration)

				if videoGet.Response.Count > 0 {
					// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
					// Запрос к API VK  wall.post ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
					paramsWallPost := map[string]string{
						"owner_id":     	"-"+iniVk.VkGroupId,
						"from_group":   	"1",
						"message":        videoGet.Response.Items[0].Title,
						"attachments":  	"video-" + iniVk.VkGroupId + "_" + fmt.Sprint(videoGet.Response.Items[0].ID),
						"signed":     	"1",
						"access_token": 	iniVk.VkToken,
						"v":            	v,
					}

					if response, err = helpers.Request("v", "wall.post", paramsWallPost, &wallPost); err == nil {
						helpers.Add2Log("wall.post", wallPost.Response.PostId)
					} else {
						err = errors.New("error Request wall.post failed: " + fmt.Sprint(err) + " ; response: " + fmt.Sprint(response))
					}
				} else {
					err = errors.New("video.get invalid response: " + fmt.Sprint(response))
				}
			} else {
				err = errors.New("error Request wall.post failed: " + fmt.Sprint(err) + " ; response: " + fmt.Sprint(response))
			}
		} else {
			err = errors.New("error read ini file: " + fmt.Sprint(err))
		}
	return err
}



