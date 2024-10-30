package user

import (
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"golang.org/x/exp/rand"
)

func CheckDomain(email string) bool {
	emailDomain := strings.Split(email, "@")[1]

	client := http.Client{}
	mailList, err := client.Get("https://raw.githubusercontent.com/disposable/disposable-email-domains/master/domains.txt")
	if err != nil {
		return true
	}
	defer mailList.Body.Close()

	data, err := io.ReadAll(mailList.Body)
	if err != nil {
		log.Println("Error reading mail list: ", err)
		return true
	}

	domains := strings.Split(string(data), "\n")

	index := sort.SearchStrings(domains, emailDomain)
	if index < len(domains) && domains[index] == emailDomain {
		return false
	}

	return true
}

func generateActivationCode() int {
	rand.Seed(uint64(time.Now().UnixNano()))
	return rand.Intn(99999) + 100000
}

func generateSecureToken() (string, error) {
	rand.Seed(uint64(time.Now().UnixNano()))
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
