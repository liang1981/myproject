package repository

import (
	"GoServer-v1.0/databases"
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"strconv"
)

//查询redis数据中的url和流的访问熟练num
func SearchStream(cameraIp string, streamType string) (string, string) {
	c := databases.Rpool.Get() //从连接池，取一个链接
	defer c.Close()            //函数运行结束 ，把连接放回连接池

	err := c.Send("auth", "123456")
	if err != nil {
		fmt.Println("redis数据库密码错误", err)
	}

	keyWord := cameraIp + "_streamtype"

	//查询流的访问条次数
	n, err := redis.String(c.Do("HGet", keyWord, "num"))
	if err != nil {
		fmt.Println("查询num出错，错误信息：", err)
		return "", ""
	}
	if n == "" {
		fmt.Println("redis数据库中没有" + keyWord + "的" + "类型的流地址")
		return "", ""
	}
	//根据IP获取已存在库中的流地址
	streamUrl, err := redis.String(c.Do("HGet", keyWord, streamType))
	if err != nil {
		fmt.Println("查询流地址出错，错误信息：", err)
		return "", ""
	}
	if streamUrl == "" {
		fmt.Println("数据库中没有" + keyWord + "的" + streamType + "类型的流地址")
		return "", ""
	}
	fmt.Println("从redis获取到的数据为：", streamUrl)
	return streamUrl, n
}

//维护count的值
func UpdateCount(cameraIp string, streamCount string) {
	c := databases.Rpool.Get() //从连接池，取一个链接
	defer c.Close()            //函数运行结束 ，把连接放回连接池
	err := c.Send("auth", "123456")
	if err != nil {
		fmt.Println("redis数据库密码错误", err)
	}

	keyWord := cameraIp + "_streamtype"

	//根据IP获取已存在库中的流地址
	_, err = c.Do("HMSet", keyWord, "num", streamCount)
	if err != nil {
		fmt.Println("数据更新失败,错误信息：", err)
		return
	}
	fmt.Println("数据更新成功")
	return
}

//如果cout的值为0时，删除该流信息
func DelStream(cameraIp string) {
	c := databases.Rpool.Get() //从连接池，取一个链接
	defer c.Close()            //函数运行结束 ，把连接放回连接池
	err := c.Send("auth", "123456")
	if err != nil {
		fmt.Println("redis数据库密码错误", err)
	}

	keyWord := cameraIp + "_streamtype"

	//根据IP获取已存在库中的流地址
	_, err = c.Do("Del", keyWord)
	if err != nil {
		fmt.Println("删除失败，错误信息：", err)
		return
	}
	fmt.Println("数据更新成功")
	return
}

//查询redis数据中的流的num
func SearchStreamNum(cameraIp string) (uint, error) {
	c := databases.Rpool.Get() //从连接池，取一个链接
	defer c.Close()            //函数运行结束 ，把连接放回连接池

	err := c.Send("auth", "123456")
	if err != nil {
		fmt.Println("redis数据库密码错误", err)
	}

	keyWord := cameraIp + "_streamtype"

	//查询流的访问条次数
	n, err := redis.String(c.Do("HGet", keyWord, "num"))
	if err != nil {
		tips := fmt.Sprintf("查询num出错，错误信息：", err)
		fmt.Println(tips)
		return 0, errors.New(tips)
	}
	num, err := strconv.Atoi(n)
	if err != nil {
		return 0, err
	}
	return uint(num), nil
}
