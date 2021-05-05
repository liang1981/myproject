//@title routers.go
//@Description 完成web服务的请求路由功能，完成对服务器的数据响应功能
package routers

import (
	"GoServer-v1.0/jwt"
	"GoServer-v1.0/repository"
	"GoServer-v1.0/tcpClient"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

//定义一个16位字符串，作为加密密钥
//key的长度必须是16、24或者32字节，分别用于选择AES-128, AES-192, or AES-256
var AESKey = []byte("Wp2MKTaxt77p1Ep7")

// 定义登录信息结构体
type LoginData struct {
	UserName string `form:"userAccount" json:"userAccount" binding:"required"`
	Password string `form:"secretCode" json:"secretCode" binding:"required"`
}

//通用用户查询结构体
type CommonUserNamePRequest struct {
	UserName string `form:"userAccount" json:"userAccount" binding:"required"`
}

type CommonCameraIPRequest struct {
	CameraIP string `form:"cameraIp" json:"cameraIp" binding:"required"`
}

// 定义录像回放的结构体
type PlayBackVideoRequest struct {
	UserName  string `form:"userAccount" json:"userAccount" binding:"required"`
	StartTime string `form:"startTime" json:"startTime"`
	EndTime   string `form:"endTime" json:"endTime"`
}

//定义删除录像结构体
type RemovePlayBackVideoIDRequest struct {
	IDs string `form:"ids" json:"ids" binding:"required"`
}

//定义删除录像结构体
type CameraAliasNameRequest struct {
	ID        string `form:"id" json:"id" binding:"required"`
	AliasName string `form:"aliasName" json:"aliasName" binding:"required"`
}

// 定义接收前端发送的json的结构体
type WebInfo struct {
	UserName string `form:"userAccount" json:"userAccount"`
	CameraIP string `form:"cameraIp" json:"cameraIp" binding:"required"`
	Type     string `form:"streamType" json:"streamType" binding:"required"`
}

//拉流/关流请求SIP服务器结构体
type RequestStream struct {
	CameraIP   string `form:"cameraIP" json:"cameraIP"`
	StreamType string `form:"streamType" json:"streamType"`
	Method     string `form:"method" json:"method"`
}

//定义请求结构体，用于打包给sip请求流信息
type RequestControlStream struct {
	CameraIP   string `form:"cameraIP" json:"cameraIP"`
	StreamType string `form:"streamType" json:"streamType"`
	Method     string `form:"method" json:"method"`
	Value      string `form:"value" json:"value"`
}

//定义流信息结构体用于接收SIP服务回复的信息
type Stream struct {
	StreamUrl  string `form:"streamUrl" json:"streamUrl" binding:"required"`
	StreamType string `form:"streamType" json:"streamType" binding:"required"`
}

//判断某个字符串是否在指定数组中
func IsContain(items []string, item string) bool {
	for _, eachItem := range items {
		if eachItem == item {
			return true
		}
	}
	return false
}

//对sip回复的responseStatus进行切割
func GetValues(str string) []string {
	statusValue := strings.Split(str, "\"")
	return statusValue
}

//@todo 临时对登录用户判断，仅限于不连数据库
func validateDemoUserName(username string) error {
	//有非加密的用户名和加密的用户名
	if username != "admin" && username != "suoni" {
		return errors.New(fmt.Sprintf("用户%s不存在", username))
	}
	return nil
}

//@todo 针对传输数据解密, 传入必须是指针
func ReflectAesDecryptRequest(req interface{}) error {
	//有部分字段不需要解密，设定一个数组
	noDecryptFields := []string{"StartTime", "EndTime", "AliasName"}

	mType := reflect.TypeOf(req)
	mVal := reflect.ValueOf(req)
	if mType.Kind() == reflect.Ptr {
		// 传入的req是指针，需要.Elem()取得指针指向的value
		mType = mType.Elem()
		mVal = mVal.Elem()
	} else {
		return errors.New("request must be ptr to struct")
	}
OuterLoop:
	for i := 0; i < mType.NumField(); i++ {
		t := mType.Field(i) //字段名类型
		f := mVal.Field(i)  //值类型
		//判断是否是字符串，目前仅仅支持string
		if t.Type.Name() != "string" {
			return errors.New("invalid data")
		}
		//如果结构体中value为空，不做处理
		if f.String() == "" {
			continue
		}
		//验证是否是不需要解密的字段
		for _, ndf := range noDecryptFields {
			if ndf == t.Name {
				//跳出到外层循环
				continue OuterLoop
			}
		}
		//首先需要base64解密
		byteVal, err := base64.StdEncoding.DecodeString(f.String())
		if err != nil {
			return err
		}
		decryptVal, err := jwt.AesDecrypt(byteVal, AESKey)
		if err != nil {
			return err
		}
		f.Set(reflect.ValueOf(string(decryptVal)))
	}
	fmt.Println()
	return nil
}

//操作摄像头
func OperateCamera(c *gin.Context) (*WebInfo, error) {
	// 声明接收的变量
	var webJson WebInfo
	err := c.ShouldBind(&webJson)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("webJosn数据解析出错, 错误信息：%s", err.Error()))
	}
	fmt.Println("webJson", webJson)

	//针对加密参数解密
	err = ReflectAesDecryptRequest(&webJson)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("reflect error:%s", err.Error()))
	}
	fmt.Println("decrypt webJson", webJson)

	var arrayStreamType = []string{"rtsp", "rtmp", "hls", "flv"}
	if IsContain(arrayStreamType, webJson.Type) == false {
		return nil, errors.New("请求类型不对")
	}

	return &webJson, nil
}

//操作录像
func OperateVideoTape(c *gin.Context) (*CommonCameraIPRequest, error) {
	// 声明接收的变量
	var cipJson CommonCameraIPRequest
	err := c.ShouldBind(&cipJson)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("webJosn数据解析出错, 错误信息：%s", err.Error()))
	}
	fmt.Println("webJson", cipJson)

	//针对加密参数解密
	err = ReflectAesDecryptRequest(&cipJson)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("reflect error:%s", err.Error()))
	}
	fmt.Println("decrypt webJson", cipJson)

	return &cipJson, nil
}

func SIPSocket(action string, info *WebInfo) (bool, error) {
	var requestStreamJson RequestStream
	requestStreamJson.CameraIP = info.CameraIP
	requestStreamJson.StreamType = info.Type
	requestStreamJson.Method = action
	streamInfoJson, err := json.Marshal(requestStreamJson)
	if err != nil {
		return false, err
	}

	responseStatus, err := tcpClient.ClientSocket(streamInfoJson)
	if err != nil {
		return false, errors.New("连接SIP服务器失败")
	}

	strStatus := GetValues(responseStatus)
	if strStatus[3] == "true" {
		return true, nil
	}

	return false, nil
}

func SIPControlSocket(action string, info *WebInfo) (bool, error) {
	var requestStreamJson RequestControlStream
	requestStreamJson.CameraIP = info.CameraIP
	requestStreamJson.StreamType = info.Type
	requestStreamJson.Method = "ptzControl"
	requestStreamJson.Value = action
	streamInfoJson, err := json.Marshal(requestStreamJson)
	if err != nil {
		return false, err
	}

	responseStatus, err := tcpClient.ClientSocket(streamInfoJson)
	if err != nil {
		return false, errors.New("连接SIP服务器失败")
	}

	strStatus := GetValues(responseStatus)
	if strStatus[3] == "true" {
		return true, nil
	}

	return false, nil
}

//解析录像回放时间
func ParseVideoTapeDateTime(r PlayBackVideoRequest) (string, string, error) {
	sTimeStr := ""
	eTimeStr := ""

	compareStime := time.Time{}
	if r.StartTime != "" {
		sTime, layout, err := ParseDateTime(r.StartTime)
		if err != nil {
			return sTimeStr, eTimeStr, err
		}
		sTimeStr = sTime.Format(layout)
		compareStime = sTime
	}

	if r.EndTime != "" {
		//此时startTime必须存在
		if r.StartTime == "" {
			return sTimeStr, eTimeStr, errors.New("start time must be not null")
		}

		eTime, layout, err := ParseDateTime(r.EndTime)
		if err != nil {
			return sTimeStr, eTimeStr, err
		}

		//对比两者大小
		if eTime.Before(compareStime) {
			return sTimeStr, eTimeStr, errors.New("end time must be greater start time")
		}

		eTimeStr = eTime.Format(layout)
	}

	return sTimeStr, eTimeStr, nil
}

//format time
func ParseDateTime(timeStr string) (time.Time, string, error) {
	layout := "2006-01-02"
	if strings.Contains(timeStr, " ") {
		layout = "2006-01-02 15:04:05"
	}
	fmt.Println("dateTime layout: ", layout)
	t, err := time.ParseInLocation(layout, timeStr, time.Local)
	if err != nil {
		return t, "", err
	}
	return t, layout, nil
}

//路由函数，完成其前端的url请求的路由
func Routers() {

	// 创建路由 使用默认的了2个中间件Logger(), Recovery()
	r := gin.Default()
	//定义/rnznsp/路由组
	v := r.Group("/rhznsp")
	{
		v.POST("/getCameraStreamUrlByIp", requestCameraStream)
		v.POST("/closeCameraStreamByIp", closeCameraStream)
		v.POST("/cameraOut", ZoomOUT)
		v.POST("/cameraIn", ZoomIN)
		v.POST("/cameraDown", TiltDown)
		v.POST("/cameraUp", TiltUp)
		v.POST("/cameraLeft", PanLeft)
		v.POST("/cameraRight", PanRight)
		v.POST("/stopPtzControl", ZommStop)
		v.POST("/startRecord", startRecord)
		v.POST("/stopRecord", stopRecord)

		//登录接口
		v.POST("/login", login)
		//获取用户信息
		v.POST("/getUserInfo", getUserInfo)
		//获取所有设备IP
		v.POST("/getAllCameras", getAllCameras)
		//获取录像回放数据
		v.POST("/getPlayBackVideoList", getPlayBackVideoList)
		//删除录像回放数据
		v.POST("/removePlayBackVideo", removePlayBackVideo)
		//更新设备别名
		v.POST("/cameraAliasName", cameraAliasName)

	}
	_ = r.Run(":8082")
}

//登录
func login(c *gin.Context) {
	var loginData LoginData

	// 将request的body中的数据，自动按照json格式解析到结构体
	if err := c.ShouldBind(&loginData); err != nil {
		// 返回错误信息
		c.JSON(http.StatusBadRequest, gin.H{"code": "500", "msg": "should bind json err: " + err.Error()})
		return
	}

	//针对加密参数解密
	err := ReflectAesDecryptRequest(&loginData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "500", "msg": "reflect error:" + err.Error()})
		return
	}

	//@todo 演示版本，因为后续拉取流媒体播放地址是没有绑定关联用户信息，不做密码校验和返回数据
	err = validateDemoUserName(loginData.UserName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "500", "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "200", "msg": "登录成功"})
	return
}

func getUserInfo(c *gin.Context) {
	// @todo 此处传入的是userAccount，在数据库中没有设置唯一值
	var getUInfoSON CommonUserNamePRequest
	err := c.ShouldBind(&getUInfoSON)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": "should bind err: " + err.Error()})
		return
	}
	fmt.Println("userAccount", getUInfoSON.UserName)

	//针对加密参数解密
	err = ReflectAesDecryptRequest(&getUInfoSON)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "500", "msg": "reflect error:" + err.Error()})
		return
	}

	u, err := repository.SelectUser(getUInfoSON.UserName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "500", "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "200", "msg": "获取数据成功", "info": u})
	return
}

//获取所有设备信息
func getAllCameras(c *gin.Context) {
	// 声明接收的变量
	var getCameraIPJSON CommonUserNamePRequest
	err := c.ShouldBind(&getCameraIPJSON)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": "should bind err: " + err.Error()})
		return
	}
	fmt.Println("userAccount", getCameraIPJSON.UserName)

	//针对加密参数解密
	err = ReflectAesDecryptRequest(&getCameraIPJSON)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "500", "msg": "reflect error:" + err.Error()})
		return
	}

	//@todo 演示版本，因为后续拉取流媒体播放地址是没有绑定关联用户信息，不做密码校验和返回数据
	err = validateDemoUserName(getCameraIPJSON.UserName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "500", "msg": err.Error()})
		return
	}

	//查询所有设备ip
	l, err := repository.GetCameras()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "200", "msg": "获取数据成功", "list": l})
	return
}

//获取录像回放数据
func getPlayBackVideoList(c *gin.Context) {
	// 声明接收的变量
	var pbvJSON PlayBackVideoRequest
	err := c.ShouldBind(&pbvJSON)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": "should bind err: " + err.Error()})
		return
	}
	fmt.Println("pbvJSON", pbvJSON)

	//针对加密参数解密
	err = ReflectAesDecryptRequest(&pbvJSON)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "500", "msg": "reflect error:" + err.Error()})
		return
	}

	//@todo 演示版本，因为后续拉取流媒体播放地址是没有绑定关联用户信息，不做密码校验和返回数据
	err = validateDemoUserName(pbvJSON.UserName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "500", "msg": err.Error()})
		return
	}

	startTime, endTime, err := ParseVideoTapeDateTime(pbvJSON)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "500", "msg": err.Error()})
		return
	}

	//查询
	l, err := repository.GetPlayBackVideo(startTime, endTime)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": err.Error()})
		return
	}

	//处理录像为空时，前端返回的json为空数组格式
	if len(l) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": "200", "msg": "获取数据成功", "list": []string{}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "200", "msg": "获取数据成功", "list": l})
	return
}

//删除录像回放数据
func removePlayBackVideo(c *gin.Context) {
	var rpbvIDJson RemovePlayBackVideoIDRequest
	err := c.ShouldBind(&rpbvIDJson)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": "should bind err: " + err.Error()})
		return
	}
	fmt.Println("play back video IDs: ", rpbvIDJson.IDs)

	err = ReflectAesDecryptRequest(&rpbvIDJson)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "500", "msg": "reflect error:" + err.Error()})
		return
	}

	err = repository.DeletePlayBackVideoByID(rpbvIDJson.IDs)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "200", "msg": "删除数据成功"})
	return
}

//更新设备别名
func cameraAliasName(c *gin.Context) {
	// 声明接收的变量
	var canJson CameraAliasNameRequest
	err := c.ShouldBind(&canJson)
	if err != nil {
		fmt.Println("josn数据解析出错, 错误信息：", err)
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": err.Error()})
		return
	}
	fmt.Println("canJson.ID", canJson.ID)
	fmt.Println("canJson.AliasName", canJson.AliasName)

	//针对加密参数解密
	err = ReflectAesDecryptRequest(&canJson)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "500", "msg": "reflect error:" + err.Error()})
		return
	}

	//更新别名
	err = repository.UpdateCameraAliasNameByIP(canJson.ID, canJson.AliasName)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "200", "msg": "更新数据成功"})
	return
}

//拉流请求
func requestCameraStream(c *gin.Context) {
	webJson, err := OperateCamera(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "500", "msg": err.Error()})
		return
	}

	//数据库中数据更新，和web端不同步
	//判断该IP是否已注册到数据库中
	_, err = repository.SelectCamera(webJson.CameraIP)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": webJson.CameraIP + "尚未注册"})
		return
	}

	//IP已存在MySQL的摄像头表中注册时
	//声明streaminfo结构体用于接受stream信息
	var streamInfo Stream
	//判断该流地址是否已存在redis数据库中
	//如果已经存在redis数据库中，则直接从数据库取出
	//将数据发送给前端后，更新redis中的访问次数的值
	var streamNum string
	streamInfo.StreamUrl, streamNum = repository.SearchStream(webJson.CameraIP, webJson.Type)
	if streamInfo.StreamUrl != "" {
		//对streamUrl进行加密
		byteUrl := []byte(streamInfo.StreamUrl)
		aesUrl, err := jwt.AesEncrypt(byteUrl, AESKey)
		if err != nil {
			fmt.Println(err)
			return
		}
		aesStreamUrl := base64.StdEncoding.EncodeToString(aesUrl)
		c.JSON(http.StatusOK, gin.H{"code": "200", "msg": "", "streamType": webJson.Type, "streamUrl": aesStreamUrl})

		num, err := strconv.Atoi(streamNum)
		num = num + 1
		streamNum = strconv.Itoa(num)
		repository.UpdateCount(webJson.CameraIP, streamNum)

		return
	}
	//给sip服务发起流请求
	isSuccess, err := SIPSocket("start", webJson)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": err.Error()})
		return
	}
	//判断SIP服务器返回是否成功
	if isSuccess == false {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": "SIP服务器没有将流地址存到数据库"})
		return
	}

	//查询redis数据库
	streamInfo.StreamUrl, _ = repository.SearchStream(webJson.CameraIP, webJson.Type)
	fmt.Println("StreamUrl:", streamInfo.StreamUrl)
	streamInfo.StreamType = webJson.Type
	if streamInfo.StreamUrl == "" {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": "没有获取到流地址"})
		return
	}

	//成功获取视频
	//给前端返回200，并携带加密后stream信息
	byteStreamUrl := []byte(streamInfo.StreamUrl)
	byteAesStreamUrl, _ := jwt.AesEncrypt(byteStreamUrl, AESKey)
	aesStreamUrl := base64.StdEncoding.EncodeToString(byteAesStreamUrl)

	c.JSON(http.StatusOK, gin.H{"code": "200", "msg": "", "streamType": streamInfo.StreamType, "streamUrl": aesStreamUrl})
	return
}

//关闭流请求
func closeCameraStream(c *gin.Context) {
	webInfoJson, err := OperateCamera(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "500", "msg": err.Error()})
		return
	}
	//根据IP和流类型查询，状态值count的数值
	//对count-1不等于0，将count更新后的值存入数据库中
	//如果count-1等于0，给sip服务器发起关流请求,完成后给删除redis数据库中的流信息
	_, strCount := repository.SearchStream(webInfoJson.CameraIP, webInfoJson.Type)
	if strCount == "" {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": "redis中没有" + webInfoJson.CameraIP + "流地址"})
		return
	}
	//得到的流的条数strCount为字符串需要转为整型
	intCount, _ := strconv.Atoi(strCount)
	intCount = intCount - 1
	if intCount != 0 {
		strCount = strconv.Itoa(intCount)
		repository.UpdateCount(webInfoJson.CameraIP, strCount)
		c.JSON(http.StatusOK, gin.H{"code": "200", "msg": ""})
		return
	}

	//给sip服务发起流请求
	isSuccess, err := SIPSocket("close", webInfoJson)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": err.Error()})
		return
	}
	//判断SIP服务器返回是否成功
	if isSuccess == false {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": "关流出错"})
		return
	}

	repository.DelStream(webInfoJson.CameraIP)
	c.JSON(http.StatusOK, gin.H{"code": "200", "msg": ""})
	return
}

//缩小
func ZoomOUT(c *gin.Context) {
	webInfoJson, err := OperateCamera(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "500", "msg": err.Error()})
		return
	}

	//给sip服务发起流请求
	isSuccess, err := SIPControlSocket("OUT", webInfoJson)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": err.Error()})
		return
	}
	//判断SIP服务器返回是否成功
	if isSuccess == false {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": "控制处理出错"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "200", "msg": ""})
	return
}

//放大
func ZoomIN(c *gin.Context) {
	webInfoJson, err := OperateCamera(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "500", "msg": err.Error()})
		return
	}

	//给sip服务发起流请求
	isSuccess, err := SIPControlSocket("IN", webInfoJson)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": err.Error()})
		return
	}
	//判断SIP服务器返回是否成功
	if isSuccess == false {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": "控制处理出错"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "200", "msg": ""})
	return

}

//上
func TiltUp(c *gin.Context) {
	webInfoJson, err := OperateCamera(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "500", "msg": err.Error()})
		return
	}

	//给sip服务发起流请求
	isSuccess, err := SIPControlSocket("UP", webInfoJson)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": err.Error()})
		return
	}
	//判断SIP服务器返回是否成功
	if isSuccess == false {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": "控制处理出错"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "200", "msg": ""})
	return
}

//下
func TiltDown(c *gin.Context) {
	webInfoJson, err := OperateCamera(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "500", "msg": err.Error()})
		return
	}

	//给sip服务发起流请求
	isSuccess, err := SIPControlSocket("Down", webInfoJson)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": err.Error()})
		return
	}
	//判断SIP服务器返回是否成功
	if isSuccess == false {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": "控制处理出错"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "200", "msg": ""})
	return
}

//左
func PanLeft(c *gin.Context) {
	webInfoJson, err := OperateCamera(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "500", "msg": err.Error()})
		return
	}

	//给sip服务发起流请求
	isSuccess, err := SIPControlSocket("Left", webInfoJson)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": err.Error()})
		return
	}
	//判断SIP服务器返回是否成功
	if isSuccess == false {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": "控制处理出错"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "200", "msg": ""})
	return
}

//右
func PanRight(c *gin.Context) {
	webInfoJson, err := OperateCamera(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "500", "msg": err.Error()})
		return
	}

	//给sip服务发起流请求
	isSuccess, err := SIPControlSocket("Right", webInfoJson)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": err.Error()})
		return
	}
	//判断SIP服务器返回是否成功
	if isSuccess == false {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": "控制处理出错"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "200", "msg": ""})
	return
}

//停止
func ZommStop(c *gin.Context) {
	webInfoJson, err := OperateCamera(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "500", "msg": err.Error()})
		return
	}

	//给sip服务发起流请求
	isSuccess, err := SIPControlSocket("Stop", webInfoJson)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": err.Error()})
		return
	}
	//判断SIP服务器返回是否成功
	if isSuccess == false {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": "控制处理出错"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "200", "msg": ""})
	return
}

func startRecord(c *gin.Context) {
	vtJson, err := OperateVideoTape(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "500", "msg": err.Error()})
		return
	}

	//首先判断redis服务器中是否已经开启了流地址，判断条件是num>0
	num, err := repository.SearchStreamNum(vtJson.CameraIP)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "500", "msg": err.Error()})
		return
	}
	if num == 0 {
		fmt.Printf("开启录像失败，原因：redis数据库中没有%s_streamtype的类型的流地址", vtJson.CameraIP)
		c.JSON(http.StatusBadRequest, gin.H{"code": "500", "msg": fmt.Sprintf("开启录像失败，请先开启直播摄像头")})
		return
	}

	//给sip服务发起流请求
	webInfoJson := &WebInfo{
		CameraIP: vtJson.CameraIP,
	}
	isSuccess, err := SIPSocket("startRecord", webInfoJson)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": err.Error()})
		return
	}
	//判断SIP服务器返回是否成功
	if isSuccess == false {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": "开启录像出错"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "200", "msg": ""})
	return
}

func stopRecord(c *gin.Context) {
	vtJson, err := OperateVideoTape(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "500", "msg": err.Error()})
		return
	}

	//给sip服务发起流请求
	webInfoJson := &WebInfo{
		CameraIP: vtJson.CameraIP,
	}
	isSuccess, err := SIPSocket("stopRecord", webInfoJson)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": err.Error()})
		return
	}
	//判断SIP服务器返回是否成功
	if isSuccess == false {
		c.JSON(http.StatusOK, gin.H{"code": "500", "msg": "暂停录像出错"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "200", "msg": ""})
	return
}
