package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

// CreateChallenge チャレンジ識別子を作成
func CreateChallenge(pDatabase *sql.DB, pSessionKey string, pChallengeVal string) error {
	if pDatabase == nil {
		return fmt.Errorf("CreateChallenge(): SessionKey=%s, Challenge=%s", pSessionKey, pChallengeVal)
	}
	pTransaction, pError := pDatabase.Begin()
	if pError != nil {
		return pError
	}

	const pSQL = "INSERT INTO TChallenges (SessionKey, ChallengeVal) VALUES ($1, $2);"
	pResult, pError := pTransaction.Exec(pSQL, pSessionKey, pChallengeVal)
	if pError != nil {
		if pPQError, ok := pError.(*pq.Error); ok {
			pSQLError := pPQError.Code.Name()
			if strings.Compare(pSQLError, "unique_violation") == 0 {
				//　既にセッション情報が登録されている
				fmt.Println("既にセッション情報が登録されている。[" + pChallengeVal + "]")
				fmt.Println(pError)
				pTransaction.Rollback()
				return pError
			} else if strings.Compare(pSQLError, "invalid_authorization_specification") == 0 {
				fmt.Println("データベースインスタンスからアクセスを拒否されました。")
				fmt.Println(pError)
				pTransaction.Rollback()
				return pError
			} else {
				fmt.Println("pSQLError: [" + pSQLError + "]")
			}
		}
		//　チャレンジ状態記録テーブルへの登録に失敗
		fmt.Println("チャレンジ状態記録テーブルへの記録に失敗しました。[" + pChallengeVal + "]")
		fmt.Println(pError)
		pTransaction.Rollback()
		return pError
	}

	nRows, pError := pResult.RowsAffected()
	if pError != nil {
		pTransaction.Rollback()
		return pError
	}

	//　セッション情報登録に成功、セッション識別子を返却
	if nRows == 0 {
		return fmt.Errorf("FAILED: CreateChallenge(): INSERT rows=[%d]", nRows)
	}

	pTransaction.Commit()

	return nil
}

// IsExistsChallenge チャレンジ値が記録に存在するかを検査
func IsExistsChallenge(pDatabase *sql.DB, pSessionKey string, pChallengeVal string) (int, error) {
	if pDatabase == nil {
		return 0, fmt.Errorf("IsExistsChallenge(): SessionKey=%s, Challenge=%s", pSessionKey, pChallengeVal)
	}
	pTransaction, pError := Begin(pDatabase)
	if pError != nil {
		return 0, pError
	}
	defer pTransaction.Commit()

	const pSQL = "SELECT COUNT(*) FROM TChallenges WHERE SessionKey = $1 AND ChallengeVal = $2;"
	pRows, pErr := pTransaction.Query(pSQL, pSessionKey, pChallengeVal)
	if pErr != nil {
		pTransaction.Rollback()
		return 0, pError
	}
	defer pRows.Close()

	var nRows int = 0
	if pRows.Next() {
		if pErr = pRows.Scan(&nRows); pErr != nil {
			if pPQError, ok := pErr.(*pq.Error); ok {
				pSQLError := pPQError.Code.Name()
				if strings.Compare(pSQLError, "invalid_authorization_specification") == 0 {
					fmt.Println("データベースインスタンスからアクセスを拒否されました。")
					fmt.Println(pErr)
				} else {
					fmt.Println("ログイン成功履歴の入力に失敗しました。[" + pSessionKey + "] [" + pChallengeVal + "]")
					fmt.Println(pErr)
				}
				return 0, pError
			}
		}
	}

	return nRows, nil
}

// DeleteChallenge チャレンジ識別子を削除
func DeleteChallenge(pDatabase *sql.DB, pSessionKey string, pChallengeVal string) (int64, error) {
	if pDatabase == nil {
		return 0, fmt.Errorf("IsExistsChallenge(): SessionKey=%s, Challenge=%s", pSessionKey, pChallengeVal)
	}
	pTransaction, pError := pDatabase.Begin()
	if pError != nil {
		return 0, pError
	}

	const pSQL = "DELETE FROM TChallenges WHERE SessionKey = $1 AND ChallengeVal = $2;"
	pResult, pError := pTransaction.Exec(pSQL, pSessionKey, pChallengeVal)
	if pError != nil {
		if pPQError, ok := pError.(*pq.Error); ok {
			pSQLError := pPQError.Code.Name()
			if strings.Compare(pSQLError, "unique_violation") == 0 {
				//　既にセッション情報が登録されている
				fmt.Println("既にセッション情報が登録されている。[" + pChallengeVal + "]")
				fmt.Println(pError)
				pTransaction.Rollback()
				return 0, pError
			} else if strings.Compare(pSQLError, "invalid_authorization_specification") == 0 {
				fmt.Println("データベースインスタンスからアクセスを拒否されました。")
				fmt.Println(pError)
				pTransaction.Rollback()
				return 0, pError
			} else {
				fmt.Println("pSQLError: [" + pSQLError + "]")
			}
		}
		//　チャレンジ状態記録テーブルへの登録に失敗
		fmt.Println("チャレンジ状態記録テーブルからの削除に失敗しました。[" + pChallengeVal + "]")
		fmt.Println(pError)
		pTransaction.Rollback()
		return 0, pError
	}
	nRows, pError := pResult.RowsAffected()
	if pError != nil {
		return 0, pError
	}

	//　セッション情報登録に成功、セッション識別子を返却

	pTransaction.Commit()

	return nRows, nil
}
