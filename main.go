package main

import (
	"DouyinDownload/TaskQueue"
	"DouyinDownload/service"
	"DouyinDownload/utils"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

type GetSignaturePara struct {
	Uid string `json:"uid"`
	Tac string `json:"tac"`
	UA  string `json:"ua"`
}

type GetSignatureResult struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"`
}

func GetSignature(userID, ua string) (string, string) {
	tac := ""
	dytk := ""

	tac, dytk, err := service.GetData("https://www.amemv.com/share/user/" + userID)
	for err != nil {
		tac, dytk, err = service.GetData("https://www.amemv.com/share/user/" + userID)
	}

	client := &http.Client{}
	data, _ := json.Marshal(GetSignaturePara{
		Uid: userID,
		Tac: base64.StdEncoding.EncodeToString([]byte(tac)),
		UA:  ua,
	})
	req, err := http.NewRequest("GET", "http://127.0.0.1:3000", bytes.NewBuffer(data))
	if err != nil {
		return "", ""
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return "", ""
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", ""
	}
	result := &GetSignatureResult{}
	err = json.Unmarshal(b, result)
	if err != nil {
		return "", ""
	}
	return result.Result, dytk
}

func read3(path string) string {
	fi, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer fi.Close()
	fd, err := ioutil.ReadAll(fi)
	return string(fd)
}

func GetAllFile(pathname string) error {
	rd, err := ioutil.ReadDir(pathname)
	for _, fi := range rd {
		if fi.IsDir() {
			GetAllFile(pathname + "/" + fi.Name() + "/")
		} else {

			oldPath := pathname + "/" + fi.Name()
			newPath := pathname + "/" + service.FilterEmoji(strings.ReplaceAll(fi.Name(), ".mp4", "")) + ".mp4"

			os.Rename(oldPath, newPath)
			utils.Log(oldPath, newPath)
		}
	}
	return err
}


type Task struct {
	UserID string
	Maxcursor int64
}

// Process ..
func (task *Task) Process() {

	os.MkdirAll("download/"+task.UserID, os.ModePerm)
	max_cursor := task.Maxcursor
	count := 0
	flag := false
	utils.Log(task.UserID + "正在下载")
	for {
		//signature, dytk := GetSignature(task.UserID, service.UA)

		err, d := service.GetVideo(task.UserID, "", "", max_cursor)
		if err != nil {
			continue
		}
		service.HandleJson(d, task.UserID, &count, &flag)
		if count > service.Config.CollectCount || flag {
			break
		}

		if d.HasMore {
			//签名失效 重新获取
			if d.MinCursor == 0 && d.MaxCursor == 0 {
				//signature, dytk = GetSignature(task.UserID, service.UA)
				utils.Log("签名失效")
			} else {
				max_cursor = d.MaxCursor
			}
		} else {
			break
		}
	}
	utils.Log(task.UserID + "下载完毕")
}

func main() {
	utils.Log("当前版本2020-11-11")

	startTime := int64(0)
	utils.Log("请输入截止时间:")
	_, err := fmt.Scanf("%d", &startTime)
	if err != nil {
		panic(err)
	}
	utils.Log("截止时间:", startTime)

	//time.Sleep(time.Second * 3)
	service.ParserConfig(read3("./config.json"))
	user := read3("./user.txt")
	uids := strings.Split(user, "\r\n")
	//uids := []string{"60305626883"}

	taskQ := TaskQueue.NewTaskQueue(len(uids), service.Config.ThreadNum)
	taskQ.Run()

	for _, userID := range uids {
		taskQ.PushItem(&Task{userID, startTime})
		//signature, dytk := GetSignature(userID, service.UA)
		//log.Println(userID + "下载完毕")
	}

	for {
		time.Sleep(time.Second)
	}
}
