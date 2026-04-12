package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

// RegistAccount アカウントを登録
func RegistAccount(pDatabase *sql.DB, pMailAddress string, pSecretHashV string, pDisplayName string) error {
	if pDatabase == nil {
		return fmt.Errorf("RegistAccount(): MailAddress=%s, SecretHashV=%s, DisplayName=%s", pMailAddress, pSecretHashV, pDisplayName)
	}
	pTransaction, pError := pDatabase.Begin()
	if pError != nil {
		return pError
	}

	const pSQL = "INSERT INTO MAccounts (DisplayName, MailAddress, SecretHashV) VALUES ($1, $2, $3);"
	pResult, pError := pTransaction.Exec(pSQL, pDisplayName, pMailAddress, pSecretHashV)
	if pError != nil {
		if pPQError, ok := pError.(*pq.Error); ok {
			pSQLError := pPQError.Code.Name()
			if strings.Compare(pSQLError, "unique_violation") == 0 {
				//　既にセッション情報が登録されている
				fmt.Println("既に同じメールアドレスが登録されている。[" + pMailAddress + "]")
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
		fmt.Println("アカウント管理テーブルへの記録に失敗しました。[" + pMailAddress + "]")
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

// CountOfAccount アカウント管理テーブルから識別子と共有秘密が一致するレコードの件数を検索
func CountOfAccount(pDatabase *sql.DB, pMailAddress string, pSecretHashV string, pUniqueId *string, pDisplayName *string) (int, error) {
	if pDatabase == nil {
		return 0, fmt.Errorf("CountOfAccount(): MailAddress=%s, SecretHashV=%s", pMailAddress, pSecretHashV)
	}
	pTransaction, pError := Begin(pDatabase)
	if pError != nil {
		return 0, pError
	}
	defer pTransaction.Commit()

	const pSQL = "SELECT UniqueId, DisplayName FROM MAccounts WHERE MailAddress = $1 AND SecretHashV = $2;"
	pRows, pErr := pTransaction.Query(pSQL, pMailAddress, pSecretHashV)
	if pErr != nil {
		pTransaction.Rollback()
		return 0, pError
	}
	defer pRows.Close()

	var nRows int = 0
	if pRows.Next() {
		if pErr = pRows.Scan(pUniqueId, pDisplayName); pErr != nil {
			if pPQError, ok := pErr.(*pq.Error); ok {
				pSQLError := pPQError.Code.Name()
				if strings.Compare(pSQLError, "invalid_authorization_specification") == 0 {
					fmt.Println("CountOfAccount(): データベースインスタンスからアクセスを拒否されました。")
					fmt.Println(pErr)
				} else {
					fmt.Println("CountOfAccount(): アカウント管理テーブルの入力に失敗しました。[" + pMailAddress + "] [" + pSecretHashV + "]")
					fmt.Println(pErr)
				}
				return 0, pError
			}
		}
		nRows++
	}

	return nRows, nil
}
