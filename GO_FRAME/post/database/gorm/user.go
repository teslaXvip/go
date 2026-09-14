package database

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/teslaXvip/go/go_frame/post/database/model"
	"gorm.io/gorm"

	"github.com/go-sql-driver/mysql"
)

func RegistUser(name, password string) (int, error) {
	user := &model.User{
		Name:     name,
		PassWord: password,
	}
	err := PostDB.Create(&user).Error
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == 1062 {
				return 0, fmt.Errorf("用户名[%s]已存在", name)
			}
		}
		slog.Error("用户注册失败", "name", name, "error", err)
		return 0, errors.New("用户注册失败，请稍后重试")
	}

	return user.Id, nil
}

// 注销用户
func LogOffUser(uid int) error {
	tx := PostDB.Delete(model.User{Id: uid})
	if tx.Error != nil {
		slog.Error("注销用户失败", "uid", uid, "error", tx.Error)
		return errors.New("用户注销失败，请稍后重试")
	}

	if tx.RowsAffected == 0 {
		return fmt.Errorf("用户注销失败,uid %d 不存在", uid)
	}

	return nil
}

func UpdatePassword(uid int, oldPass, newPass string) error {
	tx := PostDB.Model(&model.User{}).Where("id=? and password=?", uid, oldPass).Update("password", newPass)
	if tx.Error != nil {
		slog.Error("UpdatePassword failed", "uid", uid, "error", tx.Error)
		return errors.New("密码修改失败，请稍后重试")
	} else {
		if tx.RowsAffected == 0 {
			return fmt.Errorf("用户id或旧密码不对")
		}
		return nil
	}
}

func GetUserById(uid int) *model.User {
	user := model.User{Id: uid}
	tx := PostDB.Select("*").First(&user)
	if tx.Error != nil {
		if !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			slog.Error("GetUserById failed", "uid", uid, "error", tx.Error)
		}
		return nil
	}
	return &user
}
