package service

import (
	"DouyinDownload/model"
	"DouyinDownload/utils"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var httpData string

var startDateTime int64
var endDateTime int64

var Config model.DownloadConfig

func ParserConfig(data string) {
	err := json.Unmarshal([]byte(data), &Config)
	if err != nil {
		panic(err)
	}

	tm1, _ := time.ParseInLocation("2006-01-02", Config.StartDateTime, time.Local)
	startDateTime = tm1.Unix()

	tm2, _ := time.ParseInLocation("2006-01-02", Config.EndDateTime, time.Local)
	endDateTime = tm2.Unix()
}

func Download(getUrl, saveFile string) error {
	client := &http.Client{}
	rand.Seed(time.Now().Unix())
	s := strconv.Itoa(rand.Intn(100000))
	req, err := http.NewRequest("GET", getUrl, nil)
	if err != nil {
		return err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 12_0 like Mac OS X) AppleWebKit/"+s+".1.38 (KHTML, like Gecko) Version/11.0 Mobile/15A372 Safari/"+s+".1")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("upgrade-insecure-requests", "1")
	req.Header.Add("pragma", "no-cache")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.Header.Get("content-length") == "0" {
		return errors.New("ip")
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(saveFile, b, 0666)
	if err != nil {
		return err
	}
	return nil
}

func downloadHttpFile(videoUrl string, localVideo string) error {
	timeout := 0
	for {
		err := Download(videoUrl, localVideo)
		if err == nil {
			break
		}
		utils.Log(err)
		timeout++
		if timeout > 5 {
			return errors.New("超时三次失败")
		}
	}
	return nil
}

func FilterEmoji(content string) string {
	newContent := ""
	for _, value := range content {
		if unicode.Is(unicode.Han, value) || (value >= 'a' && value <= 'z') || (value >= 'A' && value <= 'Z') || unicode.IsDigit(value) || unicode.IsSpace(value) {
			newContent += string(value)
		}
	}
	return newContent
}

func getVideoDuration(itemID string) (error, model.VideoData) {
	client := &http.Client{}
	url := "https://www.iesdouyin.com/web/api/v2/aweme/iteminfo/?item_ids=" + itemID

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err, model.VideoData{}
	}
	rand.Seed(time.Now().Unix())
	s := strconv.Itoa(rand.Intn(1000))

	req.Header.Add("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3")
	req.Header.Add("accept-language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Add("cache-control", "max-age=0")
	req.Header.Add("cookie", "_ga=GA1.2.938284732.1578806304; _gid=GA1.2.1428838914.1578806305")
	req.Header.Add("upgrade-insecure-requests", "1")
	req.Header.Add("user-agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/" + s + ".36 (KHTML, like Gecko) Chrome/75.0.3770.100 Mobile Safari/537.36")
	req.Header.Add("Host", "www.iesdouyin.com")
	res, err := client.Do(req)
	if err != nil {
		return err, model.VideoData{}
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err, model.VideoData{}
	}
	var data model.VideoData
	err = json.Unmarshal(b, &data)
	if err != nil {
		return err, model.VideoData{}
	}
	return nil, data
}


func HandleJson(data model.Data, uid string, count *int, flag *bool) {
	for _, item := range data.AwemeList {

		item.Desc = strings.ReplaceAll(item.Desc, ":", "")
		item.Desc = strings.ReplaceAll(item.Desc, "?", "")
		item.Desc = strings.ReplaceAll(item.Desc, "\\", "")
		item.Desc = strings.ReplaceAll(item.Desc, "/", "")
		item.Desc = strings.ReplaceAll(item.Desc, "\"", "")
		item.Desc = strings.ReplaceAll(item.Desc, "*", "")
		item.Desc = strings.ReplaceAll(item.Desc, "<", "")
		item.Desc = strings.ReplaceAll(item.Desc, ">", "")
		item.Desc = strings.ReplaceAll(item.Desc, "|", "")
		item.Desc = strings.ReplaceAll(item.Desc, "\r", "")
		item.Desc = strings.ReplaceAll(item.Desc, "\n", "")
		item.Desc = FilterEmoji(item.Desc)
		localVideo := "download/" + uid + "/" + item.Desc + item.AwemeId[0:13] + ".mp4"

		err, videoData := getVideoDuration(item.AwemeId)
		if err != nil {
			continue
		}

		videoTime := int64(0)

		if len(videoData.AwemeList) > 0 {
			if videoData.AwemeList[0].Duration < Config.MinDuration * 1000 || videoData.AwemeList[0].Duration > Config.MaxDuration * 1000 {
				continue
			}
			videoTime = videoData.AwemeList[0].CreateTime
		} else {
			continue
		}

		downFunc := func() int {
			if utils.IsExist(localVideo) == false {
				utils.Log("开始处理数据:", item.Desc, item.AwemeId)
				//utils.log(item.Video.PlayAddr.UrlList[0])
				err := downloadHttpFile("https://aweme.snssdk.com/aweme/v1/play/?video_id="+item.Video.Vid+"&media_type=4&vr_type=0&improve_bitrate=0&is_play_url=1&is_support_h265=0&source=PackSourceEnum_PUBLISH", localVideo)
				if len(item.Video.Cover.UrlList) > 0 {
					err := downloadHttpFile(item.Video.Cover.UrlList[0], "download/" + uid + "/" + item.AwemeId[0:13] + ".jpg")
					if err != nil {
						utils.Log("下载封面失败:", err)
					}
				}

				//err := downloadHttpFile(item.Video.PlayAddr.UrlList[0], localVideo)
				if err != nil {
					utils.Log(uid + "下载视频失败:", err)
				} else {
					utils.Log("下载视频成功:", localVideo)
					*count++
				}
			} else {
				utils.Log(item.Desc + " " + localVideo + "文件已存在，跳过")
			}
			return *count
		}

		if *count >= Config.CollectCount {
			return
		}

		if Config.Flag {
			if videoTime > startDateTime && videoTime < endDateTime {
				downFunc()
			} else if videoTime < startDateTime {
				*flag = true
				return
			}
		} else {
			downFunc()
		}
	}
}

func GetData(url string) (tac, dytk string, err error) {
	client := &http.Client{}

	rand.Seed(time.Now().Unix())
	s := strconv.Itoa(rand.Intn(100000))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 12_0 like Mac OS X) AppleWebKit/"+s+".1.38 (KHTML, like Gecko) Version/11.0 Mobile/15A372 Safari/"+s+".1")

	req.Header.Add("Accept", "*/*")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Postman-Token", "07056837-9ac4-4d7a-bc47-3af9ffb58e40")
	req.Header.Add("Cookie", "odin_tt=cbdfbdf9bc6a050b5eb1847318a9632061e9f327b8a6ef1eda145d6c25a8bf4e863bffc7d109be0efe8188bf1a59755b1b804ef6bb62bc6d9eefdb57a8640553")
	req.Header.Add("Referer", url)
	req.Header.Add("Connection", "keep-alive")
	res, err := client.Do(req)
	if err != nil {
		return
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	result := string(b)

	var tacRegexp = regexp.MustCompile(`<script>(.*?)</script>`)
	tacs := tacRegexp.FindStringSubmatch(result)
	if len(tacs) > 1 {
		tac = tacs[1]
	} else {
		return tac, dytk, errors.New("查找tac失败")
	}

	var dytkRegexp = regexp.MustCompile(`dytk: '(.*?)'`)
	dytks := dytkRegexp.FindStringSubmatch(result)
	if len(dytks) > 1 {
		dytk = dytks[1]
	} else {
		return tac, dytk, errors.New("查找dytk失败")
	}
	return
}

func GetVideo(user_id, signature, dytk string, max_cursor int64) (error, model.Data) {
	client := &http.Client{}
	url := "https://www.iesdouyin.com/web/api/v2/aweme/post/?user_id=" + user_id + "&sec_uid=&count=20&max_cursor=" + strconv.FormatInt(max_cursor, 10) + "&aid=1128&_signature=" + signature + "&dytk=" + dytk

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err, model.Data{}
	}
	req.Header.Add("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3")
	req.Header.Add("accept-language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Add("cache-control", "max-age=0")
	req.Header.Add("cookie", "_ga=GA1.2.938284732.1578806304; _gid=GA1.2.1428838914.1578806305")
	req.Header.Add("upgrade-insecure-requests", "1")
	req.Header.Add("user-agent", Config.UA)
	req.Header.Add("Host", "www.iesdouyin.com")
	res, err := client.Do(req)
	if err != nil {
		return err, model.Data{}
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err, model.Data{}
	}
	var data model.Data
	err = json.Unmarshal(b, &data)
	if err != nil {
		return err, model.Data{}
	}
	return nil, data
}
