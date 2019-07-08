package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/kovetskiy/lorg"
	"golang.org/x/crypto/argon2"
)

func (server *Server) GetUserHashedPassword(login string) (string, error) {
	var hash string

	err := server.Database.QueryRow(
		`SELECT password 
			FROM users 
			WHERE login=$1`,
		login).Scan(&hash)

	if err != nil {
		return "", err
	}

	return hash, nil
}

func TestHashedPassword(hash string) bool {
	match, err := regexp.MatchString("^\\$[a-zA-Z0-9]{1,255}\\$"+
		"[a-z]=\\d{2}\\$[a-z]=([0-9]){1,},[a-z]=([0-9]){1,},"+
		"[a-z]=([0-9]){1,}\\$.{1,255}=\\$.{1,}=",
		hash,
	)
	if err != nil {
		lorg.Error(err)
		return false
	}

	return match
}

func HashPassword(password string) (string, error) {
	salt, err := GenerateSalt()
	if err != nil {
		return "", err
	}

	algo := "argon2id"
	threads := uint8(4)
	time := uint32(10)
	memory := uint32(32 * 1024)

	hash := argon2.IDKey([]byte(password), salt, time, memory, threads, 32)

	b64Hash := base64.StdEncoding.EncodeToString(hash)
	b64Salt := base64.StdEncoding.EncodeToString(salt)

	return fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		algo, argon2.Version, memory, time, threads, b64Salt, b64Hash), nil
}

func HashPasswordSettings(
	password []byte,
	salt []byte,
	algo string,
	time,
	memory uint32,
	threads uint8,
	keyLength uint32,

) string {

	var hash []byte

	switch algo {
	case "argon2id":
		hash = argon2.IDKey(password, salt, time, memory, threads, keyLength)
	case "argon2i":
		hash = argon2.Key(password, salt, time, memory, threads, keyLength)
	}

	b64Hash := base64.StdEncoding.EncodeToString(hash)
	b64Salt := base64.StdEncoding.EncodeToString(salt)

	return fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		algo, argon2.Version, memory, time, threads, b64Salt, b64Hash)
}

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 32)

	_, err := rand.Read(salt)
	if err != nil {
		return salt, err
	}

	return salt, nil
}

func ValidatePassword(password string, encodedHash string) bool {
	if ok := TestHashedPassword(encodedHash); !ok {
		return false
	}

	data, err := SplitEncodedString(encodedHash)
	if err != nil {
		lorg.Error(err)
		return false
	}

	salt, err := base64.StdEncoding.DecodeString(data["salt"].(string))
	if err != nil {
		lorg.Error(err)
		return false
	}

	savedHash, err := base64.StdEncoding.DecodeString(data["hash"].(string))
	if err != nil {
		lorg.Error(err)
		return false
	}

	encoded := HashPasswordSettings(
		[]byte(password),
		salt,
		"argon2id",
		uint32(data["times"].(int)),
		uint32(data["memory"].(int)),
		uint8(data["threads"].(int)),
		uint32(len(savedHash)))

	return subtle.ConstantTimeCompare([]byte(encoded),
		[]byte(encodedHash)) == 1
}

func SplitEncodedString(encoded string) (map[string]interface{}, error) {
	parts := make([]string, 0)
	splits := strings.SplitAfter(encoded, "$")
	splits = splits[1:]

	for _, value := range splits {
		parts = append(parts, strings.TrimSuffix(value, "$"))
	}

	versionStr := strings.Split(parts[1], "=")[1]

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		return nil, err
	}

	parameters := strings.Split(parts[2], ",")

	memStr := strings.Split(parameters[0], "=")[1]

	mem, err := strconv.Atoi(memStr)
	if err != nil {
		return nil, err
	}

	timesStr := strings.Split(parameters[1], "=")[1]

	times, err := strconv.Atoi(timesStr)
	if err != nil {
		return nil, err
	}

	threadsStr := strings.Split(parameters[2], "=")[1]

	threads, err := strconv.Atoi(threadsStr)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"type":    parts[0],
		"version": version,
		"memory":  mem,
		"times":   times,
		"threads": threads,
		"salt":    parts[3],
		"hash":    parts[4],
	}

	return data, nil
}
