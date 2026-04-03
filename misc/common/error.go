package common

import "github.com/go-sql-driver/mysql"

// IsDuplicateError 检查是否是重复键错误
func IsDuplicateError(err error) bool {
	if err == nil {
		return false
	}

	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		// MySQL 重复键错误码
		duplicateErrorCodes := map[uint16]bool{
			1062: true, // ER_DUP_ENTRY
			1169: true, // ER_DUP_UNIQUE
			1586: true, // ER_DUP_ENTRY_WITH_KEY_NAME
		}
		return duplicateErrorCodes[mysqlErr.Number]
	}
	return false
}
