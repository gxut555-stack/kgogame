/**
 * @Author: Liu Hong Wei
 * @File: test.go
 * @Date: 2024/11/25
 * @Description:
 */
package encrypt

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"testing"
)

func TestAesDecrypt(t *testing.T) {
	origData := []byte("Hello World")                 // 待加密的数据
	key := []byte("ABCDEFGHIJKLMNOPABCDEFGHIJKLMNOP") // 加密的密钥
	log.Println("原文：", string(origData))
	log.Println("------------------ CBC模式 --------------------")
	encrypted, _ := AesEncryptCBC(origData, key)
	log.Println("密文(hex)：", hex.EncodeToString(encrypted))
	log.Println("密文(base64)：", base64.StdEncoding.EncodeToString(encrypted))
	decrypted, _ := AesDecryptCBC(encrypted, key)
	log.Println("解密结果：", string(decrypted))

	log.Println("------------------ ECB模式 --------------------")
	encrypted, _ = AesEncryptECB(origData, key)
	log.Println("密文(hex)：", hex.EncodeToString(encrypted))
	log.Println("密文(base64)：", base64.StdEncoding.EncodeToString(encrypted))
	decrypted, _ = AesDecryptECB(encrypted, key)
	log.Println("解密结果：", string(decrypted))

	log.Println("------------------ CFB模式 --------------------")
	encrypted, _ = AesEncryptCFB(origData, key)
	log.Println("密文(hex)：", hex.EncodeToString(encrypted))
	log.Println("密文(base64)：", base64.StdEncoding.EncodeToString(encrypted))
	decrypted, _ = AesDecryptCFB(encrypted, key)
	log.Println("解密结果：", string(decrypted))
}

func TestAesEncryptCBC(t *testing.T) {
	apiKey := "apiKey"
	payload := `{"account": "uaserAccount"}`
	playloadSecretKey := "r3a1i7r3e0g0w1p3p9o3i8b8l9k6j2f0"
	apiSecret := "p3h5k5s4w5c3z1m2m9f4o6k7l0n2i9p6"
	timeStamp := 1732590705000
	nonce := "a4t8b5u1w4v6n3a2"

	log.Println("===== 加密 =====")
	aesEncryted, _ := AesEncryptCBC([]byte(payload), []byte(playloadSecretKey))
	encryptedPayload := base64.StdEncoding.EncodeToString(aesEncryted)
	log.Println("encryptedPayload: ", encryptedPayload)

	//构造新payload
	newPayload := map[string]string{"payload": string(encryptedPayload)}
	newPayloadBytes, _ := json.Marshal(newPayload)
	log.Println("newPayload: ", string(newPayloadBytes))

	// 拼接生成签名的字符串
	genSignStr := string(newPayloadBytes) + fmt.Sprintf("%d", timeStamp) + nonce + apiKey
	log.Println("genSignStr: ", genSignStr)

	// 生成签名
	sign := GenSha256Sign([]byte(genSignStr), apiSecret)
	log.Println("sign: ", sign)

	log.Println("===== 解密 =====")
	base64DecodeStr, _ := base64.StdEncoding.DecodeString(encryptedPayload)
	aesDecrypted, _ := AesDecryptCBC(base64DecodeStr, []byte(playloadSecretKey))
	log.Println("解密结果：", string(aesDecrypted))
}
