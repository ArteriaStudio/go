package api

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// IP許可リスト
var allowedIPs = map[string]bool{
	"116.82.244.244": true,
	"133.203.173.96": true,
}

// DumpRequestFormsStdout aaa
func DumpRequestFormsStdout(r *http.Request) {
	fmt.Println("<p>")
	fmt.Println("--- All Form Items ---<br />")
	for key, values := range r.Form {
		for _, value := range values {
			fmt.Printf("%s: %s\n", key, value)
		}
	}
	fmt.Println("</p>")
}

// DumpRequestHeader リクエストヘッダーをダンプ
func DumpRequestHeader(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<p>")
	fmt.Fprintln(w, "--- All Headers ---<br />")
	for key, values := range r.Header {
		for _, value := range values {
			fmt.Fprintf(w, "%s: %s\n", key, value)
		}
	}
	fmt.Fprintln(w, "</p>")
}

// DumpRequestHeaderStdout リクエストヘッダーをダンプ
func DumpRequestHeaderStdout(r *http.Request) {
	fmt.Println("<p>")
	fmt.Println("--- All Headers ---<br />")
	for key, values := range r.Header {
		for _, value := range values {
			fmt.Printf("%s: %s\n", key, value)
		}
	}
	fmt.Println("</p>")
}

// DumpRequestParameters リクエストパラメータをダンプ
func DumpRequestParameters(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<p>")
	fmt.Fprintln(w, "--- All Parameters ---<br />")
	path := r.URL.Path
	rawQuery := r.URL.RawQuery
	scheme := r.URL.Scheme
	host := r.URL.Host

	fmt.Fprintf(w, "Path: %s<br />\n", path)
	fmt.Fprintf(w, "RawQuery: %s<br />\n", rawQuery)
	fmt.Fprintf(w, "Scheme: %s<br />\n", scheme)
	fmt.Fprintf(w, "Host: %s<br />\n", host)
	fmt.Fprintln(w, "</p>")
}

// DumpRequestBody リクエストボディーをダンプ
func DumpRequestBody(w http.ResponseWriter, r *http.Request) {
	pBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	fmt.Fprintln(w, "<p>")
	fmt.Fprintln(w, "--- Request Body ---<br />")
	fmt.Fprintln(w, string(pBody))
	fmt.Fprintln(w, "</p>")
}

// DumpRequestBodyStdout リクエストボディーをダンプ
func DumpRequestBodyStdout(r *http.Request) {
	pBody, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("<p>no body</p")
		return
	}
	defer r.Body.Close()

	fmt.Println("<p>")
	fmt.Println("--- Request Body ---<br />")
	fmt.Println(string(pBody))
	fmt.Println("</p>")
}

// DumpSlice スライスをダンプ
func DumpSlice(w http.ResponseWriter, pResults []string) {
	for iItem, pItem := range pResults {
		fmt.Fprintf(w, "[%d]: %s\n", iItem, pItem)
	}
}

// getClientIP はリクエストからクライアントのIPアドレスを取得します
func GetClientIP(r *http.Request) string {
	// ロードバランサやプロキシ経由の場合
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	// 直接接続の場合
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// DumpClientIP クライアント側IPアドレスをダンプ
func DumpClientIP(w http.ResponseWriter, r *http.Request) {
	pClientIP := GetClientIP(r)
	fmt.Fprintf(w, "<p>")
	fmt.Fprintf(w, "ClientIP: %s<br />", pClientIP)
	fmt.Fprintf(w, "</p>")
}

// CheckRequest リクエストをフィルタ
// true：リクエストを拒否
func CheckRequest(r *http.Request) bool {
	pClientIP := GetClientIP(r)
	fChecked := !allowedIPs[pClientIP]
	if !fChecked {
		return fChecked
	}

	// Microsoft Graphの変更通知
	fChecked, _ = IsIPInSubnet(pClientIP, "20.20.32.0/19")
	if !fChecked {
		return fChecked
	}
	fChecked, _ = IsIPInSubnet(pClientIP, "20.190.128.0/18")
	if !fChecked {
		return fChecked
	}
	fChecked, _ = IsIPInSubnet(pClientIP, "20.231.128.0/19")
	if !fChecked {
		return fChecked
	}
	fChecked, _ = IsIPInSubnet(pClientIP, "40.126.0.0/18")
	if !fChecked {
		return fChecked
	}

	return !fChecked
}

// IsIPInSubnet Generated Gemini.2025/10/07
func IsIPInSubnet(ipStr string, cidrStr string) (bool, error) {
	// 1. 検査対象のIPアドレスを解析
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, fmt.Errorf("無効なIPアドレス: %s", ipStr)
	}

	// IPv4アドレスのみを検査する場合は、.To4()でnilチェックを行う
	if ip.To4() == nil {
		return false, fmt.Errorf("IPv4アドレスではありません: %s", ipStr)
	}

	// 2. サブネット（CIDR）を解析
	// ParseCIDRはネットワークアドレスとIPNet構造体を返す
	_, ipNet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return false, fmt.Errorf("無効なCIDR表記: %s, エラー: %w", cidrStr, err)
	}

	// 3. Containsメソッドを使用して包含関係を検査
	// IPNet.Contains(IP) は、IPがネットワークに含まれる場合にtrueを返す
	return ipNet.Contains(ip), nil
}

// DumpSessionKey セッション識別子を出力
func DumpSessionKey(w http.ResponseWriter, r *http.Request, pSessionID string) {
	fmt.Fprintf(w, "SessionID: %s<br />\n", pSessionID)
}

// IsCheckWellKnownURI 既知のURIに向けたリクエストであるかを判定
func IsCheckWellKnownURI(pRequestURI string) bool {
	status := true
	switch pRequestURI {
	case "/style.css":
	case "/favicon.ico":
	case "/google-style.css":
	default:
		status = false
	}
	return status
}

// IsListenerURI リスナーイベントであるかを判定
func IsListenerURI(pRequestURI string) bool {
	status := true
	switch pRequestURI {
	case "/listener/entra":
	case "/listener/entra/lifecycle":
	case "/listener/google":
	case "/listener/google/lifecycle":
	default:
		status = false
	}
	return status
}

// GenerateSHA256Hash SHA256一方向ハッシュ関数
func GenerateSHA256Hash(data string) string {
	// 1. 新しいSHA-256ハッシュオブジェクトを生成
	h := sha256.New()

	// 2. ハッシュ化したいデータをバイト列として書き込む
	// 処理に成功すると nil が返されます。
	h.Write([]byte(data))

	// 3. 最終的なハッシュ値（バイトスライス）を取得
	hashBytes := h.Sum(nil)

	// 4. 人間が読みやすいように16進数文字列にエンコードして返す
	return hex.EncodeToString(hashBytes)
}

// GetLocalTimeStampByRFC3399 RFC3399形式で表現された日時表現をローカル時刻表現に変換
func GetLocalTimeStampByRFC3399(pTimeStamp string) string {
	pTime, pError := time.Parse(time.RFC3339, pTimeStamp)
	if pError == nil {
		pTimeStamp = pTime.Local().Format(time.RFC3339)
	}

	return pTimeStamp
}
