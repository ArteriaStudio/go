package postgres

import (
	"database/sql"
	"fmt"
	"time"
)

// CreateSessionKey 発行したセッションキーを記録
func CreateSessionKey(pDatabase *sql.DB, pSessionKey string, pExpireStamp time.Time) error {
	pExpireTimestamp := pExpireStamp.Format(time.RFC1123Z)
	pSQL := "INSERT INTO TSessions (SessionKey, ExpireStamp) VALUES ($1, $2) ON CONFLICT ON CONSTRAINT tsessions_pkey DO UPDATE SET SessionKey = $1, ExpireStamp = $2"
	pResult, pError := pDatabase.Exec(pSQL, pSessionKey, pExpireTimestamp)
	if pError != nil {
		return pError
	}
	nRows, pError := pResult.RowsAffected()
	if pError != nil {
		return pError
	}
	if nRows == 0 {
		return fmt.Errorf("CreateSessionKey(): FAILED: UPSERT count=%d", nRows)
	}

	return nil
}

// UpdateSessionKey セッション情報にユニークキーを記録
func UpdateSessionKey(pDatabase *sql.DB, pSessionKey string, pUniqueId string) error {
	pSQL := "UPDATE TSessions SET UniqueId = $1 WHERE SessionKey = $2"
	pResult, pError := pDatabase.Exec(pSQL, pUniqueId, pSessionKey)
	if pError != nil {
		return pError
	}
	nRows, pError := pResult.RowsAffected()
	if pError != nil {
		return pError
	}
	if nRows == 0 {
		return fmt.Errorf("UpdateSessionKey(): FAILED: UPDATE count=%d", nRows)
	}

	return nil
}

// DeleteSessionKey セッション管理テーブルからセッション情報を削除
func DeleteSessionKey(pDatabase *sql.DB, pSessionKey string) error {
	pSQL := "DELETE FROM TSessions WHERE SessionKey = $1"
	pResult, pError := pDatabase.Exec(pSQL, pSessionKey)
	if pError != nil {
		return pError
	}
	nRows, pError := pResult.RowsAffected()
	if pError != nil {
		return pError
	}
	if nRows == 0 {
		return fmt.Errorf("DeleteSessionKey(): FAILED: DELETE count=%d", nRows)
	}

	return nil
}
