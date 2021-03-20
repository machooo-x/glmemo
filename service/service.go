package service

import (
	"glmemo/helper/database"
	"glmemo/helper/syslog"
	"glmemo/model"
	"net/http"
	"time"

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

	tx, err := database.Mysql.Begin()
	defer func() {
		if tx.Commit() != nil {
			syslog.Clog.Errorln(true, err)
			tx.Rollback()
		}
	}()
	stmt, err := tx.Prepare("insert into user values(?,?,?,?)")
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
	stmt, err := database.Mysql.Prepare("select id,user_id,title,text,update_time from record where user_id = ? order by `update_time` desc")
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
		err = result.Scan(&record.ID, &record.UUID, &record.Title, &record.Text, &dataTemp)
		if err != nil {
			syslog.Clog.Errorln(true, err)
			return err
		}
		record.Date = time.Unix(dataTemp, 0).Format("2006-01-02 15:04:05")
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
	stmt, err := database.Mysql.Prepare("select id,user_id,title,text,update_time from record where id = ?")
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
	var dataTemp int64
	for result.Next() {
		err = result.Scan(&record.ID, &record.UUID, &record.Title, &record.Text, &dataTemp)
		if err != nil {
			syslog.Clog.Errorln(true, err)
			return err
		}
		record.Date = time.Unix(dataTemp, 0).Format("2006-01-02 15:04:05")
	}
	return c.JSON(http.StatusOK, record)
}
func addRecord(c echo.Context) (err error) {
	userID := c.QueryParam("uuid")
	if userID == "" {
		return c.String(http.StatusBadRequest, "用户uuid不许为空")
	}
	recordID := c.QueryParam("recordid")
	if recordID != "" {
		syslog.Clog.Infoln(true, recordID)
	}

	type req struct {
		Title string `json:"title"`
		Text  string `json:"text"`
	}
	reqData := &req{}
	err = c.Bind(&reqData)
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	syslog.Clog.Infoln(true, reqData)
	if reqData.Text == "" {
		return c.String(http.StatusBadRequest, "text不许为空")
	}
	tx, err := database.Mysql.Begin()
	stmt, err := tx.Prepare("insert into record(id,user_id,update_time,title,text,size) values(?,?,?,?,?,?)")
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(uuid.New().String(), userID, time.Now().Unix(), reqData.Title, reqData.Text, len(reqData.Text))
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}

	return func() error {
		if tx.Commit() != nil {
			syslog.Clog.Errorln(true, err)
			return tx.Rollback()
		}
		return nil
	}()
}

func changeRecord(c echo.Context) (err error) {
	recordid := c.QueryParam("recordid")
	if recordid == "" {
		return c.String(http.StatusUnauthorized, "记录的id不许为空")
	}

	type req struct {
		Title string `json:"title"`
		Text  string `json:"text"`
	}
	reqData := &req{}
	err = c.Bind(&reqData)
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	syslog.Clog.Infoln(true, reqData)
	if reqData.Text == "" {
		return c.String(http.StatusBadRequest, "text不许为空")
	}

	tx, err := database.Mysql.Begin()
	defer func() {
		if tx.Commit() != nil {
			syslog.Clog.Errorln(true, err)
			tx.Rollback()
		}
	}()
	stmt, err := tx.Prepare("UPDATE record SET title = ?, text = ?,update_time = ? WHERE id = ?")
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(reqData.Title, reqData.Text, time.Now().Unix(), recordid)
	if err != nil {
		syslog.Clog.Errorln(true, err)
		return err
	}
	return
}

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
