package main

import (
	"./helpers"
	"./pic"
	"./vid"
	"fmt"
	"os"
)

func main() {
	var err error
	if len(os.Args) > 1 {
		arg := os.Args[1]
		switch arg {
		case "pinVidToDB":
			if err = vid.PinterestVidToDB(); err != nil {
				helpers.Add2Log("PinterestVidToDataBase", fmt.Sprint(err))
			}
		case "pinPicToDB":
			if err = pic.PinterestPicToDB(); err != nil {
				helpers.Add2Log("PinterestPicToDB", fmt.Sprint(err))
			}
		case "youTubeToDB":
			if err = vid.YouTubePlayListToDB(); err != nil {
				helpers.Add2Log("YouTubePlayListToDataBase", fmt.Sprint(err))
			}
		case "postVidToVkAlbum":
			if err = vid.PostVideoToVkAlbum(); err != nil {
				helpers.Add2Log("PostVideoToVkAlbum", fmt.Sprint(err))
			}
		case "postVidToVkGroupWall":
			if err = vid.PostVideoToVkGroupWall(); err != nil {
				helpers.Add2Log("PostVideoToVkGroupWall", fmt.Sprint(err))
			}
		case "postPicToVk":
			if err = pic.PostPicToVk(); err != nil {
				helpers.Add2Log("PicPostToVk", fmt.Sprint(err))
			}
		case "downloadPic":
			if err = pic.DownloadPic(); err != nil {
				helpers.Add2Log("ImageDownload", fmt.Sprint(err))
			}
		default:
			helpers.Add2Log("main", "argument - ", arg)
		}
	} else {
		helpers.Add2Log("main", "please enter argument")
	}

}
