package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-ldap/ldap"
	"github.com/golang/glog"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

var ErrNotFound = errors.New("request resouce not found")
var ErrAuthorizedRequired = errors.New("authorized required")

func parseRequestBody(r *http.Request, i interface{}) error {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return err
	}
	glog.Infoln("Request Body:", string(b))
	if err := json.Unmarshal(b, i); err != nil {
		return err
	}

	return nil
}

func httpAddrMaker(addr string) string {
	if strings.HasSuffix(addr, "/") {
		addr = strings.TrimRight(addr, "/")
	}

	if !strings.HasPrefix(addr, "http://") {
		return fmt.Sprintf("http://%s", addr)
	}

	return addr
}

func httpsAddrMaker(addr string) string {
	if strings.HasSuffix(addr, "/") {
		addr = strings.TrimRight(addr, "/")
	}

	if !strings.HasPrefix(addr, "https://") {
		return fmt.Sprintf("https://%s", addr)
	}

	return addr
}

func Schemastripper(addr string) string {
	schemas := []string{"https://", "http://"}

	for _, schema := range schemas {
		if strings.HasPrefix(addr, schema) {
			return strings.TrimLeft(addr, schema)
		}
	}

	return ""
}

func stripBearToken(authValue string) string {
	return strings.TrimSpace(strings.TrimLeft(authValue, "Bearer"))
}

func convertDFValidateName(name string) string {
	return strings.Replace(name, ".", "-", -1)
}

func contains(l []string, s string) bool {
	for _, str := range l {
		if str == s {
			return true
		}
	}
	return false
}

func getMd5(content []byte) string {
	md5Ctx := md5.New()
	md5Ctx.Write(content)
	cipherStr := md5Ctx.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

func base64Encode(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}

func base64Decode(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

func etcdProfilePath(username string) string {
	return fmt.Sprintf(ETCDUserProfile, username)
}

func etcdOrgPath(orgid string) string {
	return fmt.Sprintf(ETCDOrgsPrefix, orgid)
}

func etcdGeneratePath(path, key string) string {
	return fmt.Sprintf(path, key)
}

func ldapUser(user string) (ldapuser string) {
	ldapuser = fmt.Sprintf(LdapEnv.Get(LDAP_BASE_DN), user)
	glog.Infoln("user:", ldapuser)
	return ldapuser

}

func genRandomToken() (string, error) {
	c := 16
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		glog.Error("error:", err)
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}

func generatedOrgName(strlen int) (name string) {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func checkIfExistldap(err error) bool {
	if err == nil {
		return false
	}

	if e, ok := err.(*ldap.Error); ok && e.ResultCode == ldap.LDAPResultEntryAlreadyExists {
		return true
	}

	return false
}
