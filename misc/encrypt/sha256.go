/**
 * @Author: Liu Hong Wei
 * @File: Sha256.go
 * @Date: 2024/11/26
 * @Description:
 */
package encrypt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// GenSha256Sign 生成HMAC sha256 签名
func GenSha256Sign(data []byte, apiSecret string) string {
	mac := hmac.New(sha256.New, []byte(apiSecret))
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}
