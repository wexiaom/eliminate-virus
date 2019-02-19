package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

var secret = "8fbd540d0b23197df1d5095f0d6ee46d"
var appId = "wxa2c324b63b2a9e5e"
var gameUrl = "https://wxwyjh.chiji-h5.com"
var m int64 = 1000000

type RespData struct {
	Data map[string]interface{} `json:"data"`
	Code int                    `json:"code"`
}

func main() {

	coin := flag.Int64("coin", 0, "金币数")
	zuanShi := flag.Int64("zuanShi", 0, "钻石数量")
	tiLi := flag.Int64("tiLi", 0, "体力值")
	openID := flag.String("open_id", "oc6rl5U6hT5qv-ItIO6sTl_rTIj8", "open_id")
	flag.Parse()

	if "" == *openID || (0 == *zuanShi && 0 == *coin && 0 == *tiLi) {
		log.Fatal("用户id为空")
	}
	getResultMap := new(RespData)

	getReqMap := make(map[string]interface{})
	getReqMap["plat"] = "wx"
	getReqMap["time"] = time.Now().UnixNano() / 1e6
	getReqMap["openid"] = *openID
	getReqMap["wx_appid"] = appId
	getReqMap["wx_secret"] = secret
	getReqMap["sign"] = SignMap(getReqMap)
	delete(getReqMap, "wx_appid")
	delete(getReqMap, "wx_secret")

	getReqData, _ := json.Marshal(getReqMap)

	if err := PostWxGame("/api/archive/get", getReqData, getResultMap); err != nil {
		log.Fatal(err)
	}
	if getResultMap.Code != 0 {
		log.Fatal("获取用户信息失败")
	}

	recordStr := getResultMap.Data["record"]
	recordMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(fmt.Sprintf("%v", recordStr)), &recordMap); err != nil {
		log.Fatal(err)
	}
	if *coin > 0 {
		recordMap["money"] = fmt.Sprintf("%d", *coin*m)
	}
	if *zuanShi > 0 {
		recordMap["zuanShi"] = *zuanShi * m
	}
	if *tiLi > 0 {
		recordMap["tiLi"] = 999 * m
	}
	recordMap["sign"] = SignDataMap(recordMap)
	recordJsonStr, _ := json.Marshal(recordMap)

	fmt.Println(string(recordJsonStr))

	reqMap := make(map[string]interface{})
	reqMap["plat"] = "wx"
	reqMap["record"] = string(recordJsonStr)
	reqMap["time"] = time.Now().UnixNano() / 1e6
	reqMap["openid"] = *openID
	reqMap["wx_appid"] = appId
	reqMap["wx_secret"] = secret
	reqMap["sign"] = SignMap(reqMap)

	delete(reqMap, "wx_appid")
	delete(reqMap, "wx_secret")
	reqData, _ := json.Marshal(reqMap)

	fmt.Println(string(reqData))

	uploadResult := new(RespData)
	if err := PostWxGame("/api/archive/upload", reqData, uploadResult); err != nil {
		log.Fatal(err)
	}
	if uploadResult.Code == 0 {
		fmt.Println("更新用户信息成功")
	} else {
		fmt.Printf("刷新数据失败,%v", uploadResult)
	}
}

func PostWxGame(uri string, req []byte, respModel interface{}) error {
	resp, err := http.Post(gameUrl+uri, "application/json", bytes.NewReader(req))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if respModel != nil {
		if err := json.Unmarshal(body, respModel); err != nil {
			return err
		}
	}
	return nil
}

func SignMap(dataMap map[string]interface{}) string {
	reqKeys := make([]string, 0, len(dataMap))
	for key := range dataMap {
		reqKeys = append(reqKeys, key)
	}
	sort.Strings(reqKeys)

	reqSignEle := make([]string, 0)
	for i, j := 0, len(reqKeys); i < j; i++ {
		reqSignEle = append(reqSignEle, reqKeys[i]+"="+fmt.Sprintf("%v", dataMap[reqKeys[i]]))
	}
	reqBS := md5.Sum(bytes.NewBufferString(strings.Join(reqSignEle, "&")).Bytes())
	return strings.ToLower(hex.EncodeToString(reqBS[:]))
}

func SignDataMap(dataMap map[string]interface{}) string {
	// 取出所有的键，并排序
	keys := make([]string, 0, len(dataMap))
	for key := range dataMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// 拼接参数
	ignS := ""
	for i, j := 0, len(keys); i < j; i++ {
		ignS += fmt.Sprintf("%v", dataMap[keys[i]])
	}
	bs := md5.Sum(bytes.NewBufferString(ignS).Bytes())
	return strings.ToLower(hex.EncodeToString(bs[:]))
}
