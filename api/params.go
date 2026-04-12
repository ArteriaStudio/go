package api

import "time"

// Params システムパラメータ
type Params struct {
	SubscribeExtentDuration time.Duration // サブスクリプションの有効期限拡張間隔
	AccessTknExtentDuration time.Duration // アクセストークンの有効期限拡張間隔
}

func (pParams *Params) Initialize() {
	pParams.SubscribeExtentDuration = time.Hour * 24
	pParams.AccessTknExtentDuration = time.Hour * 24
}
