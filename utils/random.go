package utils

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// Uuid 生成全球唯一标识符
func Uuid() (string, error) {
	// 生成四个随机数用于 UUID 格式
	part1, err := RandInt(0, 0xffff)
	if err != nil {
		return "", err
	}
	part2, err := RandInt(0, 0xffff)
	if err != nil {
		return "", err
	}
	part3, err := RandInt(0, 0xffff)
	if err != nil {
		return "", err
	}
	part4, err := RandInt(0x4000, 0x4fff)
	if err != nil {
		return "", err
	}
	part5, err := RandInt(0x8000, 0xbfff)
	if err != nil {
		return "", err
	}

	// 生成唯一的 MD5 片段
	md5Part := md5.Sum([]byte(fmt.Sprintf("%d", part1)))
	md5Str := hex.EncodeToString(md5Part[:])[:12]

	// 组合成 UUID
	return fmt.Sprintf("%04x%04x-%04x-%04x-%04x-%s", part1, part2, part3, part4, part5, md5Str), nil
}

// Build 生成随机字符串
func Build(strType string, length int) (string, error) {
	var pool string
	switch strType {
	case "alpha":
		pool = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	case "alnum":
		pool = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	case "numeric":
		pool = "0123456789"
	case "noZero":
		pool = "123456789"
	case "unique", "md5":
		return GenerateMD5(Uniqid(string(RandInt64()))), nil
		//return GenerateMD5(), nil
	case "encrypt", "sha1":
		return GenerateSHA1(Uniqid(string(RandInt64()))), nil
		//return GenerateSHA1(), nil
	default:
		return "", fmt.Errorf("unsupported type: %s", strType)
	}

	return GenerateRandomString(pool, length)
}

// GenerateRandomString 根据字符池生成指定长度的随机字符串
func GenerateRandomString(pool string, length int) (string, error) {
	result := strings.Builder{}
	poolLen := big.NewInt(int64(len(pool)))

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, poolLen)
		if err != nil {
			return "", err
		}
		result.WriteByte(pool[num.Int64()])
	}

	return result.String(), nil
}

func Uniqid(prefix string) string {
	now := time.Now()
	sec := now.Unix()
	usec := now.UnixNano() % 0x100000
	return fmt.Sprintf("%s%08x%05x", prefix, sec, usec)
}

// GenerateMD5 生成 MD5 唯一字符串
func GenerateMD5(str string) string {
	if str == "" {
		str = fmt.Sprintf("%d", RandInt64())
	}
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}

// GenerateSHA1 生成 SHA1 唯一字符串
func GenerateSHA1(str string) string {
	if str == "" {
		str = fmt.Sprintf("%d", RandInt64())
	}
	hash := sha1.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}

// RandInt 生成 min 到 max 之间的随机整数
func RandInt(min, max int64) (int64, error) {
	num, err := rand.Int(rand.Reader, big.NewInt(max-min+1))
	if err != nil {
		return 0, err
	}
	return num.Int64() + min, nil
}

// RandInt64 生成一个随机 int64 数
func RandInt64() int64 {
	n, _ := rand.Int(rand.Reader, big.NewInt(1<<63-1))
	return n.Int64()
}
