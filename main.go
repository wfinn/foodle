//go:generate go run static/genstatic.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Result struct {
	Name     string
	Votes    map[string]string
	MostUsed string
	Group    string
}

func newCookie(name, value string) *http.Cookie {
	return &http.Cookie{Name: name, Value: value, HttpOnly: true, SameSite: http.SameSiteStrictMode, MaxAge: 99999999}
}

func extractFood(in string) (out string) {
	out = in
	pat := regexp.MustCompile(` \(.+\)$`)
	if pat.MatchString(in) {
		match := pat.FindString(in)
		out = match[2 : len(match)-1]
	}
	return
}

func getMostUsedValue(m map[string]string) (mostUsed string) {
	counts := make(map[string]int)
	for _, v := range m {
		food := extractFood(v) // getMostUsedValue should only get the most used value, not parse comments
		counts[food] = counts[food] + 1
	}
	max := 0
	for k := range counts {
		if counts[k] > max {
			max = counts[k]
			mostUsed = k
		}
	}
	return
}

func randInt(min, max int) int {
	return min + rand.Intn(max-min)
}

func getVotesFilename(group string) string {
	return "votes-" + group + "-" + strings.Split(time.Now().String(), " ")[0] + ".json"
}

func randomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(randInt('A', 'Z'+1))
	}
	return string(bytes)
}

func handleVote(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	group := getGroupname(r.URL.Query().Get("group"))
	food := r.URL.Query().Get("food")
	secret := ""
	votes, err := readJsonMap(getVotesFilename(group))
	if err != nil {
		fmt.Println(err)
		return
	}
	users, err := readJsonMap("users.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(name) < 1 || len(food) < 1 {
		http.Error(w, "name or food was missing", http.StatusForbidden)
		return
	}
	if len(users[name]) > 0 {
		secretCookie, err := r.Cookie("secret")
		if err != nil || len(secretCookie.Value) < 1 || secretCookie.Value != users[name] {
			http.Error(w, "secret is incorrect / name already taken", http.StatusForbidden)
			return
		}
		secret = secretCookie.Value
	} else {
		secret = randomString(32)
		users[name] = secret
	}
	http.SetCookie(w, newCookie("name", name))
	http.SetCookie(w, newCookie("secret", secret))
	votes[name] = food
	writeJsonMap("users.json", users)
	writeJsonMap(getVotesFilename(group), votes)
	http.Redirect(w, r, group, http.StatusFound) 
}

func getGroupname(group string) string {
	reg, err := regexp.Compile("[^a-zA-Z0-9-]+")
	if err != nil {
		return ""
		//can this even happen?
	}
	return reg.ReplaceAllString(group, "")
}

func readJsonMap(filename string) (map[string]string, error) {
	data := make(map[string]string)
	dataJson, err := ioutil.ReadFile(filename)
	if err != nil {
		return data, nil
	}
	err = json.Unmarshal(dataJson, &data)
	return data, nil
}

func writeJsonMap(filename string, data map[string]string) {
	jsonString, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
		return
	}
	ioutil.WriteFile(filename, []byte(jsonString), 0644)

}

func getCookieValue(r *http.Request, cookiename string) (value string) {
	if cookie, err := r.Cookie(cookiename); err == nil && len(cookie.Value) > 0 {
		value = cookie.Value
	}
	return
}

func handleAll() func(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("foodle").Parse(Files["static/index.html"])
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
	return func(w http.ResponseWriter, r *http.Request) {
		group := getGroupname(r.URL.EscapedPath())
		if !strings.Contains(r.Header.Get("Accept"), "text/html") {
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Frame-Options", "DENY")
		votes, err := readJsonMap(getVotesFilename(group))
		if err != nil {
			fmt.Println(err)
			return
		}
		name := strings.TrimSpace(getCookieValue(r, "name"))
		res := Result{Name: name, Votes: votes, MostUsed: getMostUsedValue(votes), Group: group}
		if err := t.ExecuteTemplate(w, "T", res); err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}
	}
}

func main() {
	port := flag.Uint("p", 8080, "port to listen to")
	flag.Parse()
	rand.Seed(time.Now().UTC().UnixNano())
	http.HandleFunc("/", handleAll())
	http.HandleFunc("/vote", handleVote)
	http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
}
