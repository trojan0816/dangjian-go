package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

// 登陆后获取并管理cookie
var jar, _ = cookiejar.New(nil)

// API总入口
var APIHost = "https://capi.dangjianwang.com"

// 默认用post
var method = "POST"

// 用于存放解析json的数据，cms, bbs, exam通用
type Data struct {
	List []map[string]string
}

// 用于存放解析json的数据，cms, bbs, exam通用
type ArticleList struct {
	Data Data
}

// 获取验证码，手动输入。返回验证码和登陆token
func getCaptcha() (string, string) {
	url := APIHost + "/official/ucenter/login/preCaptcha"
	method := "GET"
	data := ""
	b := httpHandle(method, url, data)
	type PreCaptcha struct {
		Data map[string]string
	}
	var preCaptcha PreCaptcha
	json.Unmarshal(b, &preCaptcha)
	url = preCaptcha.Data["captcha_url"]
	b = httpHandle(method, url, data)
	ioutil.WriteFile("captcha.jpg", b, 0644)
	exec.Command("Powershell", "./captcha.jpg").Run()
	var captchaCode string
	fmt.Print("请输入验证码：")
	fmt.Scanln(&captchaCode)
	return captchaCode, preCaptcha.Data["captcha_token"]
}

// 登陆
func Login() {
	var name, pwd string
	// name = ""
	// pwd = ""
	fmt.Print("手机号码/用户名： ")
	fmt.Scanln(&name)
	fmt.Println()
	fmt.Print("请输入密码： ")
	fmt.Scanln(&pwd)
	fmt.Println()
	captcha, ctoken := getCaptcha()
	baseURL := APIHost + "/official/ucenter/login/index"
	data := url.Values{
		"username":      {name},
		"password":      {pwd},
		"captcha_code":  {captcha},
		"captcha_token": {ctoken},
	}
	httpHandle(method, baseURL, data.Encode())
	return
}

// 首页前10篇文章的评论
func cms() {
	log.Println("开始首页评论")
	baseURL := APIHost + "/official/cms/article/list"
	data := url.Values{
		"menu_id":    {"90d0ac07a626a7a6b39c246d6532f7bd"},
		"page_index": {"1"},
		"page_size":  {"10"},
		"system_id":  {"5d8e26a0798913e5117584c21a18d76a"},
	}
	b := httpHandle(method, baseURL, data.Encode())
	var articleList ArticleList
	json.Unmarshal(b, &articleList)
	baseURL = APIHost + "/official/cms/comment/add"
	for _, article := range articleList.Data.List {
		data := url.Values{
			"article_id": {article["id"]},
			"content":    {"关注"},
			"img":        {},
		}
		httpHandle(method, baseURL, data.Encode())
		log.Println("评论：", article["title"])
	}
	log.Println("首页评论已完成")
}

// 党员论坛回复
func bbs() {
	log.Println("开始党员论坛回复")
	baseURL := APIHost + "/official/bbs/home/listBySys"
	data := url.Values{
		"page_index": {"1"},
		"page_size":  {"2"},
		"system_id":  {"21888"},
	}
	b := httpHandle(method, baseURL, data.Encode())
	var articleList ArticleList
	json.Unmarshal(b, &articleList)
	baseURL = APIHost + "/official/bbs/comment/add"
	for _, article := range articleList.Data.List {
		data := url.Values{
			"content":   {"关注"},
			"img":       {},
			"pid":       {article["id"]},
			"system_id": {"21888"},
		}
		httpHandle(method, baseURL, data.Encode())
		log.Println("回复：", article["title"])
	}
	log.Println("党员论坛回复已完成")
}

//  党员视角
func view() {
	log.Println("开始党员视角发布")
	baseURL := APIHost + "/official/view/View/publish"
	data := url.Values{
		"auth":    {"0"},
		"content": {"好"},
		"img":     {},
	}
	httpHandle(method, baseURL, data.Encode())
	httpHandle(method, baseURL, data.Encode())
	log.Println("党员视角发布已完成")
}

// 学习心得体会
func study() {
	log.Println("开始学习心得体会")
	baseURL := APIHost + "/official/study/comment/add"
	data := url.Values{
		"content": {"不忘初心，方得始终。中国共产党人的初心和使命，就是为中国人民谋幸福，" +
			"为中华民族谋复兴。这个初心和使命是激励中国共产党人不断前进的根本动力。" +
			"全党同志一定要永远与人民同呼吸、共命运、心连心，永远把人民对美好生活的向往作为奋斗目标，" +
			"以永不懈怠的精神状态和一往无前的奋斗姿态，继续朝着实现中华民族伟大复兴的宏伟目标奋勇前进。"},
		"mid": {"ad0bdd6d140a3aa05945f3b7d6b3a74b"},
	}
	httpHandle(method, baseURL, data.Encode())
	log.Println("学习心得体会已完成")
}

// 答题
func exam() {
	log.Println("开始答题")
	// 获取题库答案
	baseURL := APIHost + "/official/exam/competition/order"
	data := url.Values{"id": {"19"}}
	b := httpHandle(method, baseURL, data.Encode())
	var answers ArticleList
	json.Unmarshal(b, &answers)
	// 获取本次试题
	baseURL = APIHost + "/official/exam/competition/begin"
	b = httpHandle(method, baseURL, data.Encode())
	var questions ArticleList
	json.Unmarshal(b, &questions)
	// 根据本次试题中每题id遍历题库，id相同取出答案
	baseURL = APIHost + "/official/exam/ques/check"
	for _, question := range questions.Data.List {
		for _, answer := range answers.Data.List {
			if answer["id"] == question["id"] {
				data := url.Values{
					"answer_loc":  {"2"},
					"bank_id":     {"19"},
					"question_id": {question["id"]},
					"user_answer": {answer["answer"]},
				}
				httpHandle(method, baseURL, data.Encode())
			}
		}
	}
	log.Println("答题已完成")
}

// 签到
func checkin() {
	log.Println("开始签到")
	baseURL := APIHost + "/official/ucenter/ucuser/checkin"
	data := url.Values{"client": {"2"}, "version": {"0.0.1"}}
	httpHandle(method, baseURL, data.Encode())
	log.Println("签到已完成")
}

// 在线学习时长
func studyTime() {
	log.Println("开始在线学习时长")
	baseURL := APIHost + "/official/study/Common/startStudy"
	httpHandle(method, baseURL, "")
	for i := 303; i >= 0; i-- {
		fmt.Printf("还需要继续学习%03d秒。\r", i)
		time.Sleep(1 * time.Second)
	}
	baseURL = APIHost + "/official/study/Common/endStudy"
	data := url.Values{
		"mid":      {"ad0bdd6d140a3aa05945f3b7d6b3a74b"},
		"type":     {"1"},
		"web_time": {"305"},
	}
	httpHandle(method, baseURL, data.Encode())
	log.Println("在线学习时长已完成")
}

// 构建client，发送请求
func httpHandle(method, url, data string) []byte {
	var client = &http.Client{Jar: jar}
	var req *http.Request
	if data == "" {
		req, _ = http.NewRequest(method, url, nil)
	} else {
		req, _ = http.NewRequest(method, url, strings.NewReader(data))
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) "+
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.26 "+
		"Safari/537.36 Core/1.63.6735.400 QQBrowser/10.2.2614.400")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	req.Header.Add("Appid", "33beba686fd8333e") // 不加入该项无法获取cookie
	// req.Header.Add("Accept-Encoding", "gzip, deflate, br")  // 加入该项不能获取验证码图片
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	// cookies := jar.Cookies(req.URL)
	b, _ := ioutil.ReadAll(resp.Body)
	time.Sleep(2 * time.Second)
	return b //, cookies
}

func main() {
	Login()
	cms()
	bbs()
	view()
	study()
	exam()
	exam()
	checkin()
	studyTime()
}
