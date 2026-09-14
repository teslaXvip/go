package database

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/teslaXvip/go/go_frame/post/database/model"

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
