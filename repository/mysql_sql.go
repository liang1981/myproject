//@Title mysql_sql.go
//@Description 服务数数据库的查询功能

package repository

import (
	"GoServer-v1.0/databases"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

//定义全局结构体，用于接收查询道德摄像头信息I
type CamearInfo struct {
	CameraIP string
	ID       string
}

type UserInfo struct {
	UserID      int    `json:"userId"`
	DeptID      int    `json:"deptId"`
	LoginName   string `json:"loginName"`
	UserName    string `json:"userName"`
	UserType    string `json:"userType"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
	Sex         string `json:"sex"`
	Avatar      string `json:"avatar"`
	Password    string `json:"password"`
	Salt        string `json:"salt"`
	Status      string `json:"status"`
	DelFlag     string `json:"delFlag"`
	LoginIP     string `json:"loginIp"`
	LoginDate   string `json:"loginDate"`
	CreateBy    string `json:"createBy"`
	CreateTime  string `json:"createTime"`
	UpdateBy    string `json:"updateBy"`
	UpdateTime  string `json:"updateTime"`
	Remark      string `json:"remark"`
}

//定义camere结构体
type RegisterCamera struct {
	ID              string `json:"id"`
	CameraIP        string `json:"cameraIp"`
	CameraAliasName string `json:"cameraAliasName"`
}

//定义回放视频结构体
type PlayBackVideo struct {
	ID         string `json:"id"`
	CameraIP   string `json:"cameraIp"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
	RecordName string `json:"recordName"`
	RecordUrl  string `json:"recordUrl"`
}

//根据IP来查询Mysql数据库的camera_register表，通过表的主键来判断IP是否已经注册到数据库中
func SelectCamera(IP string) (string, error) {

	var camera CamearInfo
	//根据IP进行单行数据查询
	err := databases.Db.QueryRow("select id from camera_register where camera_ip=?", IP).Scan(&camera.ID)
	if err != nil {
		return "", err
	}
	//关闭数据库
	//defer databases.Db.Close()
	//	println(camera.id)
	return camera.ID, nil
}

//根据IP来查询Mysql数据库的sys_user表，通过表的主键来判断用户是否已存在数据库中
func SelectUser(userName string) (*UserInfo, error) {
	queryFields := "user_id,IFNULL(dept_id,0),login_name,user_name,user_type,email,phonenumber,sex,avatar,password,salt,status,del_flag,login_ip,IFNULL(login_date,''),create_by,create_time,update_by,IFNULL(update_time,''),IFNULL(remark,'')"
	row := databases.Db.QueryRow(fmt.Sprintf("select %s from sys_user where status=0 and del_flag=0 and login_name='%s'", queryFields, userName))

	var u UserInfo

	err := row.Scan(&u.UserID, &u.DeptID, &u.LoginName, &u.UserName, &u.UserType, &u.Email, &u.PhoneNumber, &u.Sex, &u.Avatar, &u.Password, &u.Salt, &u.Status, &u.DelFlag, &u.LoginIP, &u.LoginDate, &u.CreateBy, &u.CreateTime, &u.UpdateBy, &u.UpdateTime, &u.Remark)
	if err != nil {
		if strings.Contains(err.Error(), "sql: no rows in result set") {
			return nil, errors.New("没有当前数据")
		}
		return nil, err
	}
	return &u, nil
}

//查询所有设备ip信息
func GetCameras() ([]*RegisterCamera, error) {
	rows, err := databases.Db.Query("select id,camera_ip, IFNULL(camera_alias_name, camera_ip) from camera_register where status='ON';")
	if err != nil {
		return nil, err
	}

	var cameras []*RegisterCamera

	for rows.Next() {
		var camera RegisterCamera
		err := rows.Scan(&camera.ID, &camera.CameraIP, &camera.CameraAliasName)
		if err != nil {
			return nil, err
		}
		//如果没有别名，则赋值ip
		if camera.CameraAliasName == "" {
			camera.CameraAliasName = camera.CameraIP
		}

		cameras = append(cameras, &camera)
	}

	return cameras, nil
}

func GetPlayBackVideo(startTime, endTime string) ([]*PlayBackVideo, error) {
	sqlQuery := fmt.Sprintf("select id,camera_ip,record_name,start_time,IFNULL(end_time,''),record_url from play_back_video")
	if startTime != "" && endTime == "" {
		sqlQuery = fmt.Sprintf("%s where unix_timestamp(start_time) >= unix_timestamp('%s')", sqlQuery, startTime)
	} else if endTime != "" && startTime != "" {
		sqlQuery = fmt.Sprintf("%s where unix_timestamp(start_time) >= unix_timestamp('%s') and unix_timestamp(end_time) <= unix_timestamp('%s')", sqlQuery, startTime, endTime)
	}

	fmt.Println("get play back video sql:", sqlQuery)
	rows, err := databases.Db.Query(sqlQuery)
	if err != nil {
		return nil, err
	}

	var pbvs []*PlayBackVideo

	for rows.Next() {
		var pbv PlayBackVideo
		err := rows.Scan(&pbv.ID, &pbv.CameraIP, &pbv.RecordName, &pbv.StartTime, &pbv.EndTime, &pbv.RecordUrl)
		if err != nil {
			return nil, err
		}

		pbvs = append(pbvs, &pbv)
	}

	return pbvs, nil
}

func DeletePlayBackVideoByID(id string) error {
	tx, err := databases.Db.Begin()
	if err != nil {
		return err
	}
	defer clearTransaction(tx)

	res, err := tx.Exec("delete from play_back_video where id=?", id)
	if err != nil {
		return err
	}
	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func UpdateCameraAliasNameByIP(id, aliasName string) error {
	tx, err := databases.Db.Begin()
	if err != nil {
		return err
	}
	defer clearTransaction(tx)

	res, err := tx.Exec("update camera_register set camera_alias_name=? where id=?", aliasName, id)
	if err != nil {
		return err
	}
	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func clearTransaction(tx *sql.Tx) {
	err := tx.Rollback()
	if err != nil && err != sql.ErrTxDone {
		fmt.Print(err)
	}
}
