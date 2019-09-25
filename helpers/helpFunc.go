package helpers

import (
	"../models"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"
	"unsafe"
)

func UrlEncoded(str string) string {
	if r, err := regexp.Compile(","); err == nil {
		str = r.ReplaceAllString(str, "%2C")
		if r, err = regexp.Compile("\n"); err == nil {
			str = r.ReplaceAllString(str, "%0A")
			if r, err = regexp.Compile("\r"); err == nil {
				str = r.ReplaceAllString(str, "%0D")
				if u, err := url.Parse(str); err == nil {
					return u.String()
				}
			} else {
				log.Printf("UrlEncoded error regexp.Compile \r: %v", err)
			}
		} else {
			log.Printf("UrlEncoded error regexp.Compile \n: %v", err)
		}
	} else {
		log.Printf("UrlEncoded error regexp.Compile: %v", err)
	}

	return ""
}

func VideoID(link string) string {
	// Parse the URL and ensure there are no errors.
	if strings.Contains(link, "attribution_link") {
		return YTP(YTP(link, "u"), "v")
	} else {
		return YTP(link, "v")
	}
}

func YTP(link string, key string) string {
	if u, err := url.Parse(link); err == nil {
		if fragments, err := url.ParseQuery(u.RawQuery); err == nil {
			if len(fragments[key]) > 0 {
				log.Print(fragments[key][0])
				return fragments[key][0]
			} else {
				log.Print("not found")
			}
		} else {
			log.Printf("error ParseQuery: %v", err)
		}
	} else {
		log.Printf("error Parse: %v", err)
	}
	return ""
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {
	var err error
	// Get the data
	if resp, err := http.Get(url); err == nil {
		defer func() {
			if err = resp.Body.Close(); err != nil {
				err = errors.New("DownloadFile error resp.Body.Close(): " + fmt.Sprint(err))
			}
		}()
		// Create the file
		if out, err := os.Create(filepath); err == nil {
			defer func() {
				if err = out.Close(); err != nil {
					err = errors.New("DownloadFile error out.Close(): " + fmt.Sprint(err))
				}
			}()
			// Write the body to file
			if _, err = io.Copy(out, resp.Body); err != nil {
				err = errors.New("DownloadFile error io.Copy: " + fmt.Sprint(err))
			}
		} else {
			err = errors.New("DownloadFile error os.Create: " + fmt.Sprint(err))
		}
	} else {
		err = errors.New("DownloadFile error http.Get: " + fmt.Sprint(err))
	}
	return err
}

func RangeInt(min int, max int, n int) []int {
	rand.Seed(time.Now().UnixNano())
	arr := make([]int, n)
	var r int
	for r = 0; r <= n-1; r++ {
		arr[r] = rand.Intn(max) + min
	}
	return arr
}

func Add2Log(who string, what string, options ...string) {
	var opt, comma, whats string
	var mu sync.Mutex
	var err error
	path := "log"

	if len(options) > 0 {
		opt = " ("
		for _, op := range options {
			if len(opt) > 2 {
				comma = "; "
			}
			if len(op) > 0 {
				opt += comma + op
			}
		}
		opt += ")"
	}

	whats = strings.Replace(what, ":b:", "\033[97m", 1)
	whats = strings.Replace(whats, ":-:", "\033[0m", 1)
	what = strings.Replace(what, ":b:", "", 1)
	what = strings.Replace(what, ":-:", "", 1)

	if err = os.MkdirAll(path, 0777); err != nil {
		err = errors.New("Add2Log error os.MkdirAll: " + fmt.Sprint(err))
	}

	dateLog := time.Now().Format("2006_01_02")
	dateNow := time.Now().Format("2006.01.02")
	timeNow := time.Now().Format("15:04:05")

	mu.Lock()
	defer func() {
		mu.Unlock()
	}()

	if file, err := os.OpenFile(path+"/"+dateLog+"_log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644); err == nil {
		if _, err = file.WriteString(dateNow + "	" + timeNow + "	" + who + " " + what + opt + "\n"); err == nil {
			if err = file.Close(); err != nil {
				err = errors.New("Add2Log error file.Close: " + fmt.Sprint(err))
			}
		} else {
			err = errors.New("Add2Log error file.WriteString: " + fmt.Sprint(err))
		}
	} else {
		err = errors.New("Add2Log error os.OpenFile: " + fmt.Sprint(err))
	}

	if _, err = os.Stderr.WriteString(dateNow + " " + timeNow + "	\033[92m" + who + "\033[0m " + whats + opt + "\n"); err != nil {
		err = errors.New("Add2Log error os.Stderr.WriteString: " + fmt.Sprint(err))
	}

	if err != nil {
		log.Print(err)
	}
}

// PhotoWall upload file (on filePath) to given url.
// Return info about uploaded photo.
func PhotoWall(url, filePath string) (models.UploadPhotoWallResponse, error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	fileWriter, err := bodyWriter.CreateFormFile("photo", filePath)
	if err != nil {
		return models.UploadPhotoWallResponse{}, err
	}

	fh, err := os.Open(filePath)
	if err != nil {
		return models.UploadPhotoWallResponse{}, err
	}
	defer fh.Close()

	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return models.UploadPhotoWallResponse{}, err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(url, contentType, bodyBuf)
	if err != nil {
		return models.UploadPhotoWallResponse{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return models.UploadPhotoWallResponse{}, err
	}

	var uploaded models.UploadPhotoWallResponse
	err = json.Unmarshal(body, &uploaded)
	if err != nil {
		return models.UploadPhotoWallResponse{}, err
	}

	return uploaded, nil
}

func PhotoGroup(url, filePath string) (models.UploadPhotoResponse, error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	fileWriter, err := bodyWriter.CreateFormFile("file1", filePath)
	if err != nil {
		return models.UploadPhotoResponse{}, err
	}

	fh, err := os.Open(filePath)
	if err != nil {
		return models.UploadPhotoResponse{}, err
	}
	defer fh.Close()

	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return models.UploadPhotoResponse{}, err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(url, contentType, bodyBuf)
	if err != nil {
		return models.UploadPhotoResponse{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return models.UploadPhotoResponse{}, err
	}

	var uploaded models.UploadPhotoResponse
	err = json.Unmarshal(body, &uploaded)
	if err != nil {
		return models.UploadPhotoResponse{}, err
	}

	return uploaded, nil
}

// Request provides access to VK API methods.
func Request(s string, method string, params map[string]string, st interface{}) ([]byte, error) {

	var apiURL string
	switch s {
	case "v":
		apiURL = "https://api.vk.com/method/"
	case "y":
		apiURL = "https://www.googleapis.com/youtube/v3/"
	default:
		apiURL = "https://api.vk.com/method/"
	}

	//apiURL  := "https://api.vk.com/method/"

	u, err := url.Parse(apiURL + method)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	for k, v := range params {
		query.Set(k, v)
	}

	//query.Set("access_token", vk.AccessToken)
	//query.Set("v", vk.Version)
	u.RawQuery = query.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var handler struct {
		Error    *models.Error
		Response json.RawMessage
	}
	err = json.Unmarshal(body, &handler)
	err = json.Unmarshal(body, st)

	if handler.Error != nil {
		return nil, handler.Error
	}

	return handler.Response, nil
}

func ResToStruct(b *http.Response, s interface{}) error {
	jss, err := ioutil.ReadAll(b.Body)
	if err != nil {
		return fmt.Errorf("JSON ReadAll failed: %s", err)
	}

	if err := json.Unmarshal([]byte(jss), s); err != nil {
		return fmt.Errorf("JSON unmarshaling failed: %s", err)
	}

	return nil
}

func RespToStruct(b []byte, s interface{}) error {

	if err := json.Unmarshal(b, s); err != nil {
		return fmt.Errorf("JSON unmarshaling failed: %s", err)
	}

	return nil
}

func ByteToString(b []byte) string {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := reflect.StringHeader{bh.Data, bh.Len}
	return *(*string)(unsafe.Pointer(&sh))
}

func ByteToJson(b *http.Response) string {
	bs := make([]byte, 1014)
	js := ""
	for true {
		n, err := b.Body.Read(bs)
		js = js + string(bs[:n])
		if n == 0 || err != nil {
			break
		}
	}
	return js
}
