package service

import (
	"fmt"
	"glmemo/config"
	"glmemo/helper/database"
	"glmemo/helper/syslog"
	"glmemo/model"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/google/uuid"

	"github.com/labstack/echo"
)

func login(c echo.Context) (err error) {
	name := c.QueryParam("name") //获取用户名
	pwd := c.QueryParam("pwd")   //获取密码
	if name == "" || pwd == "" { //校验用户名和密码
		return c.String(http.StatusUnauthorized, "用户名和密码不许为空")
	}
	user := &model.User{}
	stmt, err := database.Mysql.Prepare("select * from user where name = ?")
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	defer stmt.Close()
	result, err := stmt.Query(name)
	defer result.Close()
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	if result.Next() { //检测是否已注册
		err = result.Scan(&user.UUID, &user.Date, &user.Name, &user.Pwd)
		if err != nil {
			syslog.Clog.Errorln(true, err)
			return err
		}
	} else {
		return c.String(http.StatusUnauthorized, "该用户未注册")
	}
	if !(name == user.Name && pwd == user.Pwd) {
		return c.String(http.StatusUnauthorized, "密码错误，请重新输入")
	}
	return c.String(http.StatusOK, user.UUID)
}

func regist(c echo.Context) (err error) {
	name := c.QueryParam("name") //获取用户名
	pwd := c.QueryParam("pwd")   //获取密码
	if name == "" || pwd == "" { //校验用户名和密码
		return c.String(http.StatusUnauthorized, "用户名和密码不许为空")
	}

	stmt, err := database.Mysql.Prepare("select * from user where name = ?")
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	defer stmt.Close()
	result, err := stmt.Query(name)
	defer result.Close()
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	if result.Next() {
		// 该用户已存在
		return c.String(http.StatusOK, "repeat name")
	}
	uuid := uuid.New().String()
	tx, err := database.Mysql.Begin() // 开启事务
	defer func(uuid string) {
		if err != nil {
			tx.Rollback() // 有错误则直接回滚，不向后进行操作
		} else {
			if tx.Commit() != nil { // 提交失败则回滚
				syslog.Clog.Errorln(true, err)
				tx.Rollback()
			} else { //注册成功后初始化标签表中的该用户相关数据
				tx, err = database.Mysql.Begin()
				defer func() {
					if tx.Commit() != nil {
						syslog.Clog.Errorln(true, err)
						tx.Rollback()
					}
				}()
				stmt, err = tx.Prepare(`insert into tag(user_id,tag_name) values(?,"我的收藏"),(?,"其它"),(?,"生活"),(?,"学习"),(?,"工作"),(?,"日常"),(?,"随笔"),(?,"家人"),(?,"朋友"),(?,"同学"),(?,"同事"),(?,"娱乐"),(?,"游戏"),(?,"网友"),(?,"书本"),(?,"电影"),(?,"电视剧")`)
				if err != nil {
					syslog.Clog.Errorln(true, err)
					return
				}
				defer stmt.Close()
				_, err = stmt.Exec(uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid, uuid)
				if err != nil {
					syslog.Clog.Errorln(true, err)
					return
				}
			}
		}
	}(uuid)
	stmt, err = tx.Prepare("insert into user values(?,?,?,?)")
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(uuid, time.Now().Unix(), name, pwd) // 在用户表中新建数据
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	return
}

func getRecordList(c echo.Context) (err error) {
	uuid := c.QueryParam("uuid") // 获取用户的id
	if uuid == "" {
		return c.String(http.StatusUnauthorized, "用户uuid不许为空")
	}
	records := make([]*model.Record, 0)
	stmt, err := database.Mysql.Prepare("select id,user_id,title,text,tag_name,filepath,update_time from record where user_id = ? order by `update_time` desc")
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	defer stmt.Close()
	result, err := stmt.Query(uuid) // 查找该用户所有的记录并根据时间倒序排列
	defer result.Close()
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	var dataTemp int64
	for result.Next() {
		record := &model.Record{}
		err = result.Scan(&record.ID, &record.UUID, &record.Title, &record.Text, &record.TagName, &record.FilePath, &dataTemp)
		if err != nil {
			syslog.Clog.Errorln(true, err)
			return err
		}
		record.Date = time.Unix(dataTemp, 0).Format("2006-01-02 15:04:05")
		if record.FilePath != "" {
			if !strings.Contains(record.FilePath, "mp4") {
				record.FileType = "img"
			} else {
				record.FileType = "mp4"
			}
		}

		records = append(records, record)
	}
	return c.JSON(http.StatusOK, records)
}

func showRecord(c echo.Context) (err error) {
	recordid := c.QueryParam("recordid") // 获取查看的文案ID
	if recordid == "" {
		return c.String(http.StatusUnauthorized, "文案的id不许为空")
	}
	record := &model.Record{}
	var dataTemp int64

	info := c.QueryParam("info") // info用来判断是查看暂存的文案，还是查看提交的文案
	if info == "" {              // 当info为空时，从临时表查找上次暂存的数据进行返回
		stmt, err := database.Mysql.Prepare("select record_id,user_id,title,text,tag_name,filepath,update_time from temp_record where record_id = ?")
		syslog.Clog.Infoln(true, "mark query temp")
		if err != nil {
			syslog.Clog.Errorln(true, err)
			return err
		}
		defer stmt.Close()
		result, err := stmt.Query(recordid)
		defer result.Close()
		if err != nil {
			syslog.Clog.Errorln(true, err)
			return err
		}
		if result.Next() {
			err = result.Scan(&record.ID, &record.UUID, &record.Title, &record.Text, &record.TagName, &record.FilePath, &dataTemp)
			if err != nil {
				syslog.Clog.Errorln(true, err)
				return err
			}
			record.Date = time.Unix(dataTemp, 0).Format("2006-01-02 15:04:05")
			if record.FilePath != "" {
				if !strings.Contains(record.FilePath, "mp4") {
					record.FileType = "img"
				} else {
					record.FileType = "mp4"
				}
			}
			idx := strings.LastIndex(record.FilePath, "/")
			if idx != -1 {
				record.FileName = record.FilePath[idx+1:]
				syslog.Clog.Infoln(true, record.FileName)
			}
			return c.JSON(http.StatusOK, record)
		}
	}
	stmt, err := database.Mysql.Prepare("select id,user_id,title,text,tag_name,filepath,update_time from record where id = ?")
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	defer stmt.Close()
	result, err := stmt.Query(recordid)
	defer result.Close()
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	if result.Next() {
		err = result.Scan(&record.ID, &record.UUID, &record.Title, &record.Text, &record.TagName, &record.FilePath, &dataTemp)
		if err != nil {
			syslog.Clog.Errorln(true, err)
			return err
		}
		record.Date = time.Unix(dataTemp, 0).Format("2006-01-02 15:04:05")
		if record.FilePath != "" {
			if !strings.Contains(record.FilePath, "mp4") {
				record.FileType = "img"
			} else {
				record.FileType = "mp4"
			}
		}
	}
	idx := strings.LastIndex(record.FilePath, "/")
	if idx != -1 {
		record.FileName = record.FilePath[idx+1:]
	}
	return c.JSON(http.StatusOK, record)
}

func queryTempSave(c echo.Context) (err error) {
	uuid := c.QueryParam("uuid")
	if uuid == "" {
		return c.String(http.StatusUnauthorized, "记录的id不许为空")
	}
	record := &model.Record{}
	stmt, err := database.Mysql.Prepare("select record_id,user_id,title,text,tag_name,filepath,update_time from temp_record where user_id = ? and is_add_save = ?")
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	defer stmt.Close()
	result, err := stmt.Query(uuid, 1)
	defer result.Close()
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	var dataTemp int64
	if result.Next() {
		err = result.Scan(&record.ID, &record.UUID, &record.Title, &record.Text, &record.TagName, &record.FilePath, &dataTemp)
		if err != nil {
			syslog.Clog.Errorln(true, err)
			return err
		}
		record.Date = time.Unix(dataTemp, 0).Format("2006-01-02 15:04:05")
	} else {
		record.Date = "0"
	}
	idx := strings.LastIndex(record.FilePath, "/")
	if idx != -1 {
		record.FileName = record.FilePath[idx+1:]
		syslog.Clog.Infoln(true, record.FileName)
	}
	return c.JSON(http.StatusOK, record)
}
func addRecord(c echo.Context) (err error) {
	userID := c.QueryParam("uuid")
	if userID == "" {
		syslog.Clog.Errorln(true, "userID==\"\"")
		return c.String(http.StatusBadRequest, "操作失败，请重新登陆")
	}
	recordID := c.QueryParam("recordid")
	if recordID == "" {
		syslog.Clog.Errorln(true, "recordid==\"\"")
		return c.String(http.StatusBadRequest, "操作失败，请重新登陆")
	}
	syslog.Clog.Infoln(true, userID, recordID)
	isCommit := c.QueryParam("iscommit")
	isAddSave := c.QueryParam("isaddsave")

	type req struct { // 用于获取用户输入内容的参数
		Title    string `json:"title"`
		Text     string `json:"text"`
		TagName  string `json:"tagname"`
		FilePath string `json:"filepath"`
		FileName string `json:"filename"`
	}
	reqData := &req{}
	err = c.Bind(&reqData)
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	syslog.Clog.Infoln(true, reqData)
	if reqData.Title == "" {
		return c.String(http.StatusBadRequest, "不许为空")
	}
	if reqData.Text == "" {
		return c.String(http.StatusBadRequest, "内容不许为空")
	}
	tx, err := database.Mysql.Begin()
	defer func(userID, recordID string, isCommit string, tagName, isAddSave string) {
		if err != nil { // 如果操作存在异常则事务回滚
			tx.Rollback()
		} else {
			if tx.Commit() != nil {
				syslog.Clog.Errorln(true, err)
				tx.Rollback()
			} else {
				if isCommit == "1" {
					tx, err := database.Mysql.Begin()
					defer func() {
						if tx.Commit() != nil {
							syslog.Clog.Errorln(true, err)
							tx.Rollback()
						}
					}()
					if isAddSave != "0" { // 如果该文案是新建的，更新标签表中该用户的数据，是该标签下文案数量加一
						stmt, err := tx.Prepare("UPDATE tag SET sum = sum+1 WHERE user_id = ? and tag_name = ?")
						if err != nil {
							syslog.Clog.Errorln(true, err)
							// return
						}
						defer stmt.Close()
						_, err = stmt.Exec(userID, tagName)
						if err != nil {
							syslog.Clog.Errorln(true, err)
							// return
						}
					}
					stmt, err := tx.Prepare("delete from temp_record WHERE record_id = ?") // 文案新建成功后，删除临时文案表中备份的记录
					if err != nil {
						syslog.Clog.Errorln(true, err)
						return
					}
					defer stmt.Close()
					_, err = stmt.Exec(recordID)
					if err != nil {
						syslog.Clog.Errorln(true, err)
						return
					}
					return
				}
			}
		}
	}(userID, recordID, isCommit, reqData.TagName, isAddSave)
	var str string
	// 如果是提交的话，将数据更新到文案表中   此处ON DUPLICATE KEY标志着如果存在则更新，不存在则添加
	if isCommit == "1" {
		str = "INSERT INTO record(id,user_id,update_time,title,text,tag_name,filename,filepath) values(?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE id = ?, user_id = ?, update_time = ?, title = ?, text = ?, tag_name = ?,filename = ?, filepath = ?"
	} else if isCommit == "0" { // 如果不是新建的话，将数据插入到临时文案表中，备份修改，便于下次获取修改记录
		str = "INSERT INTO temp_record(record_id,user_id,update_time,title,text,tag_name,filepath,is_add_save) VALUE(?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE record_id=?,user_id = ?,update_time = ?, title = ?, text= ?, tag_name = ?, filepath = ?, is_add_save = ?"
		stmt, err := tx.Prepare(str)
		if err != nil {
			syslog.Clog.Errorln(true, err)
			return err
		}
		defer stmt.Close()
		_, err = stmt.Exec(recordID, userID, time.Now().Unix(), reqData.Title, reqData.Text, reqData.TagName, reqData.FilePath, isAddSave, recordID, userID, time.Now().Unix(), reqData.Title, reqData.Text, reqData.TagName, reqData.FilePath, isAddSave)
		if err != nil {
			syslog.Clog.Errorln(true, err)
			if err.Error() == "Error 1406: Data too long for column 'title' at row 1" {
				return c.String(http.StatusBadRequest, "标题过长...")
			}
			return c.String(http.StatusBadRequest, "")
		}
		return nil
	}
	stmt, err := tx.Prepare(str)
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(recordID, userID, time.Now().Unix(), reqData.Title, reqData.Text, reqData.TagName, reqData.FileName, reqData.FilePath, recordID, userID, time.Now().Unix(), reqData.Title, reqData.Text, reqData.TagName, reqData.FileName, reqData.FilePath)
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	syslog.Clog.Infoln(true, reqData)
	return nil
}

func delRecord(c echo.Context) (err error) {
	userID := c.QueryParam("uuid")
	if userID == "" {
		return c.String(http.StatusUnauthorized, "用户uuid不许为空")
	}
	recordid := c.QueryParam("recordid")
	if recordid == "" {
		return c.String(http.StatusUnauthorized, "删除的记录id不许为空")
	}
	tagName := c.QueryParam("tagname")
	if recordid == "" {
		return c.String(http.StatusUnauthorized, "标签名称不许为空")
	}
	tx, err := database.Mysql.Begin()
	defer func(userID, tagName string) {
		if err != nil {
			tx.Rollback()
		} else {
			if tx.Commit() != nil {
				syslog.Clog.Errorln(true, err)
				tx.Rollback()
			} else {
				tx, err = database.Mysql.Begin()
				defer func() {
					if tx.Commit() != nil {
						syslog.Clog.Errorln(true, err)
						tx.Rollback()
					}
				}()
				stmt, err := tx.Prepare("UPDATE tag SET sum = sum-1 WHERE user_id = ? and tag_name = ?")
				if err != nil {
					syslog.Clog.Errorln(true, err)
					return
				}
				defer stmt.Close()
				_, err = stmt.Exec(userID, tagName)
				if err != nil {
					syslog.Clog.Errorln(true, err)
					return
				}
			}
		}
	}(userID, tagName)
	stmt, err := tx.Prepare("delete from record WHERE id = ?")
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(recordid)
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	return
}

func delTempSave(c echo.Context) (err error) {
	recordid := c.QueryParam("recordid")
	if recordid == "" {
		return c.String(http.StatusUnauthorized, "删除的记录id不许为空")
	}
	tx, err := database.Mysql.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			if tx.Commit() != nil {
				syslog.Clog.Errorln(true, err)
				tx.Rollback()
			}
		}
	}()
	stmt, err := tx.Prepare("delete from temp_record WHERE record_id = ?")
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(recordid)
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	return
}

//处理上传文件的控制器
func uploadfile(c echo.Context) (err error) {
	defer func() {
		if err != nil {
			if err.Error() == "http: no such file" {
				err = c.String(http.StatusOK, "")
			} else {
				syslog.Clog.Errorln(true, err)
				err = c.String(http.StatusOK, "文件上传失败！请重新上传文件...")
			}
		}
	}()
	syslog.Clog.Infoln(true, "uploadfile 请求")

	uuid := strings.Split(c.Request().Referer(), "=")[1][:36]
	syslog.Clog.Infoln(true, uuid)

	// 通过FormFile函数获取客户端上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	//打开用户上传的文件
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	// 创建目标文件，就是我们打算把用户上传的文件保存到什么地方
	// file.Filename 参数指的是我们以用户上传的文件名，作为目标文件名，也就是服务端保存的文件名跟用户上传的文件名一样
	syslog.Clog.Infoln(true, file.Filename)
	/* 创建上层文件夹 -------------------------------------------uuid的文件夹-------------------------------------------------- */

	folderPath := fmt.Sprintf("%s/%s/%s", "data", uuid, strconv.FormatInt(time.Now().Unix(), 10))
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		// 必须分成两步
		// 先创建文件夹
		err = os.MkdirAll(folderPath, 0777)
		if err != nil {
			return err
		}
		// 再修改权限
		err = os.Chmod(folderPath, 0777)
		if err != nil {
			return err
		}
	}
	filename := folderPath + "/" + file.Filename
	dst, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer dst.Close()
	// 这里将用户上传的文件复制到服务端的目标文件
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}
	return c.HTML(http.StatusOK, fmt.Sprintf("<p>文件上传成功: %s</p><iframe name=\"frame\" frameborder=\"0\" height=\"0\" width=\"0\"scrolling=\"no\">%s</iframe>", file.Filename, filename))
}

// 此分享链接一日有效
func createTempRecord(c echo.Context) (err error) {
	recordid := c.QueryParam("recordid")
	if recordid == "" {
		return c.String(http.StatusUnauthorized, "recordid为空")
	}
	syslog.Clog.Traceln(true, recordid)
	key := uuid.New().String()
	r := database.RedisPool.Get()
	r.Do("set", key, recordid, "EX", 24*3600)
	tempURL := fmt.Sprintf("http://%s/web/sharerecord.html?token=%s", config.GLMEMO.Section("netIP").Key("IP").String(), key)
	return c.String(http.StatusOK, tempURL)
}

func getTempRecord(c echo.Context) error {
	token := c.QueryParam("token")
	syslog.Clog.Traceln(true, token)
	record := &model.Record{}
	var dataTemp int64
	r := database.RedisPool.Get()
	recordID, err := redis.String(r.Do("GET", token))
	if err != nil {
		if err.Error() == "redigo: nil returned" {
			return c.JSON(http.StatusOK, record)
		}
		syslog.Clog.Errorln(true, err)
		return err
	}
	syslog.Clog.Traceln(true, recordID)

	stmt, err := database.Mysql.Prepare("select id,user_id,title,text,tag_name,filepath,update_time from record where id = ?")
	if err != nil {

		syslog.Clog.Errorln(true, err)
		return err
	}
	defer stmt.Close()
	result, err := stmt.Query(recordID)
	defer result.Close()
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	if result.Next() {
		err = result.Scan(&record.ID, &record.UUID, &record.Title, &record.Text, &record.TagName, &record.FilePath, &dataTemp)
		if err != nil {
			syslog.Clog.Errorln(true, err)
			return err
		}
		record.Date = time.Unix(dataTemp, 0).Format("2006-01-02 15:04:05")
		if record.FilePath != "" {
			if !strings.Contains(record.FilePath, "mp4") {
				record.FileType = "img"
			} else {
				record.FileType = "mp4"
			}
		}
	}
	idx := strings.LastIndex(record.FilePath, "/")
	if idx != -1 {
		record.FileName = record.FilePath[idx+1:]
	}
	return c.JSON(http.StatusOK, record)
}

func getAllTag(c echo.Context) error {
	userID := c.QueryParam("uuid")
	if userID == "" {
		return c.String(http.StatusUnauthorized, "用户uuid不许为空")
	}
	stmt, err := database.Mysql.Prepare("select id,tag_name,sum from tag where user_id = ? order by id")
	syslog.Clog.Infoln(true, "mark query tag")
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	defer stmt.Close()
	result, err := stmt.Query(userID)
	defer result.Close()
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	tags := make([]*model.Tag, 0)
	for result.Next() {
		tag := &model.Tag{}
		err = result.Scan(&tag.ID, &tag.TagName, &tag.Sum)
		if err != nil {
			syslog.Clog.Errorln(true, err)
			return err
		}
		tags = append(tags, tag)
	}
	return c.JSON(http.StatusOK, tags)
}

func queryByTag(c echo.Context) error {
	uuid := c.QueryParam("uuid")
	if uuid == "" {
		return c.String(http.StatusUnauthorized, "用户uuid不许为空")
	}
	tagName := c.QueryParam("tagname")
	if tagName == "" {
		return c.String(http.StatusUnauthorized, "标签不许为空")
	}
	records := make([]*model.Record, 0)
	stmt, err := database.Mysql.Prepare("select id,user_id,title,text,tag_name,filepath,update_time from record where user_id = ? and tag_name = ? order by `update_time` desc")
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	defer stmt.Close()
	result, err := stmt.Query(uuid, tagName) // 通过标签查找时，通过用户id与标签名称进行查找，建表时建立了索引，查找时间不会耗费太多时间
	defer result.Close()
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	var dataTemp int64
	for result.Next() {
		record := &model.Record{}
		err = result.Scan(&record.ID, &record.UUID, &record.Title, &record.Text, &record.TagName, &record.FilePath, &dataTemp)
		if err != nil {
			syslog.Clog.Errorln(true, err)
			return err
		}
		record.Date = time.Unix(dataTemp, 0).Format("2006-01-02 15:04:05")
		if record.FilePath != "" {
			if !strings.Contains(record.FilePath, "mp4") {
				record.FileType = "img"
			} else {
				record.FileType = "mp4"
			}
		}

		records = append(records, record)
	}
	return c.JSON(http.StatusOK, records)
}

// 模糊查询  select * from record where text like '%第%' or filename like '%mg%';
func queryByLike(c echo.Context) error {
	uuid := c.QueryParam("uuid")
	if uuid == "" {
		return c.String(http.StatusUnauthorized, "用户uuid不许为空")
	}
	paramStr := c.QueryParam("paramstr")
	if paramStr == "" {
		return c.String(http.StatusUnauthorized, "查找内容不许为空")
	}
	paramStr = "%" + paramStr + "%"
	records := make([]*model.Record, 0)
	stmt, err := database.Mysql.Prepare("select id,user_id,title,text,tag_name,filepath,update_time from record where user_id = ? and (title like ? or text like ? or filename like ?) order by `update_time` desc")
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	defer stmt.Close()
	result, err := stmt.Query(uuid, paramStr, paramStr, paramStr) // 通过mysql语句关键字like进行模糊查询并将结果依据更新时间倒序排列
	defer result.Close()
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	var dataTemp int64
	for result.Next() {
		record := &model.Record{}
		err = result.Scan(&record.ID, &record.UUID, &record.Title, &record.Text, &record.TagName, &record.FilePath, &dataTemp)
		if err != nil {
			syslog.Clog.Errorln(true, err)
			return err
		}
		record.Date = time.Unix(dataTemp, 0).Format("2006-01-02 15:04:05")
		if record.FilePath != "" {
			if !strings.Contains(record.FilePath, "mp4") {
				record.FileType = "img"
			} else {
				record.FileType = "mp4"
			}
		}
		records = append(records, record)
	}
	return c.JSON(http.StatusOK, records)
}

func getNetIP(c echo.Context) error {
	return c.String(http.StatusOK, config.GLMEMO.Section("netIP").Key("IP").String())
}
