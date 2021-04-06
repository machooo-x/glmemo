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

	name := c.QueryParam("name")
	pwd := c.QueryParam("pwd")
	if name == "" || pwd == "" {
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
	if result.Next() {
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
	name := c.QueryParam("name")
	pwd := c.QueryParam("pwd")
	if name == "" || pwd == "" {
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
		return c.String(http.StatusOK, "repeat name")
	}

	tx, err := database.Mysql.Begin()
	defer func() {
		if tx.Commit() != nil {
			syslog.Clog.Errorln(true, err)
			tx.Rollback()
		}
	}()
	stmt, err = tx.Prepare("insert into user values(?,?,?,?)")
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	defer stmt.Close()
	uuid := uuid.New().String()
	_, err = stmt.Exec(uuid, time.Now().Unix(), name, pwd)
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	return
}

func getRecordList(c echo.Context) (err error) {
	uuid := c.QueryParam("uuid")
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
	result, err := stmt.Query(uuid)
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
	recordid := c.QueryParam("recordid")
	if recordid == "" {
		return c.String(http.StatusUnauthorized, "记录的id不许为空")
	}
	record := &model.Record{}
	var dataTemp int64

	info := c.QueryParam("info")
	if info == "" {
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

	type req struct {
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
	defer func(recordID string, isCommit string, tagName string) {
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

				stmt, err := tx.Prepare("UPDATE tag SET sum = sum+1 WHERE tag_name = ?")
				if err != nil {
					syslog.Clog.Errorln(true, err)
					// return
				}
				defer stmt.Close()
				_, err = stmt.Exec(tagName)
				if err != nil {
					syslog.Clog.Errorln(true, err)
					// return
				}

				stmt, err = tx.Prepare("delete from temp_record WHERE record_id = ?")
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
	}(recordID, isCommit, reqData.TagName)

	// INSERT INTO record(id,user_id,update_time,title,text,tag_name,size) VALUE("undefined","060d55a9-f611-4c82-b6b4-994006c9c9e6",1616975733,"2","2",2) ON DUPLICATE KEY UPDATE id="undefined",title="0",text="0",size=5;

	var str string
	if isCommit == "1" {
		str = "insert into record(id,user_id,update_time,title,text,tag_name,filename,filepath,size) values(?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE id = ?, user_id = ?, update_time = ?, title = ?, text = ?, tag_name = ?,filename = ?, filepath = ?,size = ?"
	} else if isCommit == "0" {
		str = "INSERT INTO temp_record(record_id,user_id,update_time,title,text,tag_name,filename,filepath,size,is_add_save) VALUE(?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE record_id=?,user_id = ?,update_time = ?, title = ?, text= ?, tag_name = ?,filename = ?, filepath = ?, size = ?, is_add_save = ?"
		stmt, err := tx.Prepare(str)
		if err != nil {
			syslog.Clog.Errorln(true, err)
			return err
		}
		defer stmt.Close()
		_, err = stmt.Exec(recordID, userID, time.Now().Unix(), reqData.Title, reqData.Text, reqData.TagName, reqData.FileName, reqData.FilePath, len(reqData.Text), isAddSave, recordID, userID, time.Now().Unix(), reqData.Title, reqData.Text, reqData.TagName, reqData.FileName, reqData.FilePath, len(reqData.Text), isAddSave)
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
	_, err = stmt.Exec(recordID, userID, time.Now().Unix(), reqData.Title, reqData.Text, reqData.TagName, reqData.FileName, reqData.FilePath, len(reqData.Text), recordID, userID, time.Now().Unix(), reqData.Title, reqData.Text, reqData.TagName, reqData.FileName, reqData.FilePath, len(reqData.Text))
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	syslog.Clog.Infoln(true, reqData)

	return nil
}

// func changeRecord(c echo.Context) (err error) {
// 	recordid := c.QueryParam("recordid")
// 	if recordid == "" {
// 		return c.String(http.StatusUnauthorized, "记录的id不许为空")
// 	}

// 	type req struct {
// 		Title string `json:"title"`
// 		Text  string `json:"text"`
// 	}
// 	reqData := &req{}
// 	err = c.Bind(&reqData)
// 	if err != nil {
// 		syslog.Clog.Errorln(true, err)
// 		return err
// 	}
// 	syslog.Clog.Infoln(true, reqData)
// 	if reqData.Text == "" {
// 		return c.String(http.StatusBadRequest, "text不许为空")
// 	}

// 	tx, err := database.Mysql.Begin()
// 	defer func() {
// 		if tx.Commit() != nil {
// 			syslog.Clog.Errorln(true, err)
// 			tx.Rollback()
// 		}
// 	}()
// 	stmt, err := tx.Prepare("UPDATE record SET title = ?, text = ?,,tag_name = ?,update_time = ? WHERE id = ?")
// 	if err != nil {
// 		syslog.Clog.Errorln(true, err)
// 		return err
// 	}
// 	defer stmt.Close()
// 	_, err = stmt.Exec(reqData.Title, reqData.Text,reqData.TagName, time.Now().Unix(), recordid)
// 	if err != nil {
// 		syslog.Clog.Errorln(true, err)
// 		return err
// 	}
// 	return
// }

func delRecord(c echo.Context) (err error) {
	recordid := c.QueryParam("recordid")
	if recordid == "" {
		return c.String(http.StatusUnauthorized, "删除的记录id不许为空")
	}
	tx, err := database.Mysql.Begin()
	defer func() {
		if tx.Commit() != nil {
			syslog.Clog.Errorln(true, err)
			tx.Rollback()
		}
	}()
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
		if tx.Commit() != nil {
			syslog.Clog.Errorln(true, err)
			tx.Rollback()
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

func getFilename(c echo.Context) error {
	type file struct {
		Name string `json:"name"`
	}
	resp := make([]*file, 0)
	resp = append(resp, &file{
		Name: "file111"})
	resp = append(resp, &file{
		Name: "file222"})
	return c.JSON(http.StatusOK, &resp)
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
	stmt, err := database.Mysql.Prepare("select id,tag_name,sum from tag")
	syslog.Clog.Infoln(true, "mark query tag")
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	defer stmt.Close()
	result, err := stmt.Query()
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
	result, err := stmt.Query(uuid, tagName)
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
	syslog.Clog.Traceln(true, paramStr)
	records := make([]*model.Record, 0)
	stmt, err := database.Mysql.Prepare("select id,user_id,title,text,tag_name,filepath,update_time from record where user_id = ? and (title like ? or text like ? or filename like ?) order by `update_time` desc")
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	defer stmt.Close()
	result, err := stmt.Query(uuid, paramStr, paramStr, paramStr)
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
