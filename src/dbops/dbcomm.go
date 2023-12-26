// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// Author: Shiqian Yan
// ///////////////////////////////////////
package dbops

import (
	"database/sql"
	"fmt"
	. "oceanpass/src/zaplog"
	_ "oceanpass/src/zaplog"

	sts "github.com/aliyun/alibaba-cloud-sdk-go/services/sts"
	_ "github.com/go-sql-driver/mysql"

	"go.uber.org/zap"
)

func GetSecretByIdFromDB(id string, ossDb *sql.DB, tableName string) (secret string, err error) {
	sqlStr := fmt.Sprintf("select accesskey_secret from %s where accesskey_id=?;", tableName)
	rowObj := ossDb.QueryRow(sqlStr, id)
	err = rowObj.Scan(&secret)
	if err != nil {
		return "", err
	}
	return secret, nil
}

func GetRoleArnByIdFromDB(id string, ossDb *sql.DB, tableName string) (rolearn string, err error) {
	sqlStr := fmt.Sprintf("select rolearn from %s where accesskey_id=?;", tableName)
	rowObj := ossDb.QueryRow(sqlStr, id)
	err = rowObj.Scan(&rolearn)
	if err != nil {
		return "", err
	}
	return rolearn, nil
}

func InsertIntoDB(cred sts.Credentials, endPoint string, ossDb *sql.DB, tableName string) error {
	sqlStr := fmt.Sprintf("insert into %s (accesskey_id,accesskey_secret,ststoken,endpoint,expire_at) values (?,?,?,?,?);", tableName)
	_, err := ossDb.Exec(sqlStr, cred.AccessKeyId, cred.AccessKeySecret, cred.SecurityToken, endPoint, cred.Expiration)
	if err != nil {
		Logger.Error("InsertIntoDB()", zap.Any("err", err))
		return err
	}
	return nil
}

func DeleteExpireItemInDB(ossDb *sql.DB, tableName string) error {
	//SELECT * FROM oss_auth WHERE expire_at < NOW() and expire_at!="0000-00-00 00:00:00";
	sqlStr := fmt.Sprintf("DELETE FROM %s WHERE date_add(expire_at, interval 1 hour) < NOW() and expire_at!=?;", tableName)
	_, err := ossDb.Exec(sqlStr, "0000-00-00 00:00:00")
	if err != nil {
		return err
	}
	return nil
}
