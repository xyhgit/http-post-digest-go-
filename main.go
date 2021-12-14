package main

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const BASE_URL = "http://fpjinhua.1minsp.com:9001"
const REST_USER = "jinhua_api_client"
const REST_PWD = "jinhua_1q2w3e4r"

type User struct {
	OriSysId        string  `json:"oriSysId"`
	OriUserId       string  `json:"oriUserId"`
	PhoneNum        string  `json:"phoneNum"`
	NickName        string  `json:"nickName"`
	BirthDate       int64   `json:"birthDate"`
	Gender          int     `json:"gender"`
	PhotoUrl        string  `json:"photoUrl"`
	BodyWeight      float32 `json:"bodyWeight"`
	BodyHeight      int     `json:"bodyHeight"`
	Status          int     `json:"status"`
	FaceRegPhotoUrl string  `json:"faceRegPhotoUrl"`
}

func digestPost(host string, uri string) bool {
	var user User
	user.BodyHeight = 183
	user.BirthDate = 1570863754715
	user.BodyWeight = 77.5
	user.PhoneNum = "13100000002"
	user.Status = 1
	user.Gender = 0
	user.OriSysId = "JINHUA"
	user.OriUserId = "3"
	user.NickName = "金华接口3"
	user.FaceRegPhotoUrl = "https://gimg2.baidu.com/image_search/src=http%3A%2F%2Fpic.baike.soso.com%2Fp%2F20130608%2F20130608105644-952299575.jpg&refer=http%3A%2F%2Fpic.baike.soso.com&app=2002&size=f9999,10000&q=a80&n=0&g=0n&fmt=jpeg?sec=1633744805&t=56471b08200dc5afb0c2a268e7ed7ced"
	jsonStr, err := json.Marshal(user)
	if err != nil {
		fmt.Println("User json marshal error:", err)
		return false
	}
	url := host + uri
	method := "POST"
	req, err := http.NewRequest(method, url, nil)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		log.Printf("Recieved status code '%v' auth skipped", resp.StatusCode)
		return true
	}
	digestParts := digestParts(resp)
	digestParts["uri"] = uri
	digestParts["method"] = method
	digestParts["username"] = REST_USER
	digestParts["password"] = REST_PWD

	req, err = http.NewRequest(method, url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", getDigestAuthrization(digestParts))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	log.Println("response body: ", string(body))
	if resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}
func digestParts(resp *http.Response) map[string]string {
	result := map[string]string{}
	if len(resp.Header["Www-Authenticate"]) > 0 {
		wantedHeaders := []string{"nonce", "realm", "qop"}
		responseHeaders := strings.Split(resp.Header["Www-Authenticate"][0], ",")
		for _, r := range responseHeaders {
			for _, w := range wantedHeaders {
				if strings.Contains(r, w) {
					result[w] = strings.Split(r, `"`)[1]
				}
			}
		}
	}
	return result
}
func getMD5(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
func getCnonce() string {
	b := make([]byte, 8)
	io.ReadFull(rand.Reader, b)
	return fmt.Sprintf("%x", b)[:16]
}
func getDigestAuthrization(digestParts map[string]string) string {
	d := digestParts
	ha1 := getMD5(d["username"] + ":" + d["realm"] + ":" + d["password"])
	ha2 := getMD5(d["method"] + ":" + d["uri"])
	nonceCount := 00000001
	cnonce := getCnonce()
	response := getMD5(fmt.Sprintf("%s:%s:%v:%s:%s:%s", ha1, d["nonce"], nonceCount, cnonce, d["qop"], ha2))
	authorization := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", cnonce="%s", nc="%v", qop="%s", response="%s"`,
		d["username"], d["realm"], d["nonce"], d["uri"], cnonce, nonceCount, d["qop"], response)
	return authorization
}

func main() {
	digestPost(BASE_URL, "/api/fpu/face_reg?")
}
