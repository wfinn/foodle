package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

const foodle = `{{define "T"}}
<!DOCTYPE html>
<html lang="en">

<head>
    <title>Foodle</title>
    	<style>
			body {
				font-size: 20px;
				background-color: #dddddd;
				font-family: sans-serif;
			}

			table {
				border-collapse: collapse;
				border: 1px solid;
				font-size: 25px;
				width: auto;
			}

			th, td {
				text-align: left;
				padding: 8px;
			}

			input {
				font-size: 20px;
			}

			tr:nth-child(even) {background-color: #cccccc;}
		</style>
</head>

<body>
    <center>
    	<div style="width:50%; height:auto;">
	        <h1>Foodle</h1>
	        {{if .Votes}}
	        <h2>Most votes for: {{.MostUsed}}</h2>
	        <table>
	            <tr>
	                <th>Name</th>
	                <th>Food</th>
	            </tr>
	            {{range $key, $value := .Votes}}
	            <tr>
	                <td>{{$key}}</td>
	                <td>{{if $.Name}}<a href='/vote?name={{$.Name}}&food={{$value}}&token={{$.Token}}'>{{$value}}</a>{{else}}{{$value}}{{end}}</td>
	            </tr>
	            {{end}}
	        </table>
	        <br>
	        {{else}}
	        <h2>No votes yet</h2>
	        {{end}}

			<form action="/vote" method="GET">
				<input name="token" type="hidden" value="{{$.Token}}">
				<input name="name" type="text"{{if $.Name}} value="{{$.Name}}"{{end}} placeholder="your name"><br>
				<input name="food" type="search" list="foods" placeholder="your fav food"><br>
				<datalist id="foods">
					<option value="Pizza" />
					<option value="Burger" />
					<option value="Thai" />
					<option value="China Nudeln" />
					<option value="Pho Co" />
					<option value="Soy" />
					<option value="Bibi Mix" />
					<option value="Falafel" />
					<option value="Döner" />
					<option value="Mikrowelle" />
				</datalist>
				<input type="submit" value="Vote">
			</form>
		</div>
    </center>
</body>

</html>
{{end}}
`

type Result struct {
	Name     string
	Votes    map[string]string
	MostUsed string
	Token    string
}

func getMostUsedValue(m map[string]string) string {
	counts := make(map[string]int)
	for _, v := range m {
		counts[v] = counts[v] + 1
	}
	max := 0
	mostUsed := ""
	for k, _ := range counts {
		if counts[k] > max {
			max = counts[k]
			mostUsed = k
		}
	}
	return mostUsed
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func getVotesFilename() string {
	return "votes-" + strings.Split(time.Now().String(), " ")[0] + ".json"
}

func randomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(randInt('A', 'Z'+1))
	}
	return string(bytes)
}

func handleVote(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	food := r.URL.Query().Get("food")
	queryToken := r.URL.Query().Get("token")
	if cookieToken, err := r.Cookie("token"); err != nil || queryToken != cookieToken.Value {
		w.WriteHeader(403)
		w.Write([]byte("CSRF"))
		return
	}
	votes, err := readJsonMap(getVotesFilename())
	if err != nil {
		fmt.Println(err)
		return
	}
	users, err := readJsonMap("users.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(name) > 0 && len(food) > 0 {
		if len(users[name]) > 0 {
			secretCookie, err := r.Cookie("secret")
			if err != nil || len(secretCookie.Value) == 0 || secretCookie.Value != users[name] {
				w.WriteHeader(403)
				w.Write([]byte("SECRET"))
				return
			}
		} else {
			secret := randomString(32)
			users[name] = secret
			http.SetCookie(w, &http.Cookie{Name: "name", Value: name, HttpOnly: true})
			http.SetCookie(w, &http.Cookie{Name: "secret", Value: secret, HttpOnly: true})
		}
		votes[name] = food
		writeJsonMap("users.json", users)
		writeJsonMap(getVotesFilename(), votes)
	}
	http.Redirect(w, r, "/", 302)
}

func readJsonMap(filename string) (map[string]string, error) {
	dataJson, err := ioutil.ReadFile(filename)
	if err != nil {
		return make(map[string]string), nil
	}
	data := make(map[string]string)
	if err = json.Unmarshal(dataJson, &data); err != nil {
		return nil, err
	}
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

func getCookieValue(r *http.Request, cookiename string) string {
	value := ""
	if cookie, err := r.Cookie(cookiename); err == nil && len(cookie.Value) > 0 {
		value = cookie.Value
	}
	return value
}

func handleAll() func(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("foodle").Parse(foodle)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept"), "text/html") {
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Frame-Options", "DENY")
		votes, err := readJsonMap(getVotesFilename())
		if err != nil {
			fmt.Println(err)
			return
		}
		name := getCookieValue(r, "name")
		token := getCookieValue(r, "token")
		if len(token) != 32 {
			token = randomString(32)
		}
		http.SetCookie(w, &http.Cookie{Name: "token", Value: token, HttpOnly: true})
		res := Result{Name: name, Votes: votes, MostUsed: getMostUsedValue(votes), Token: token}
		if err := t.ExecuteTemplate(w, "T", res); err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}
	}
}

func main() {
	addr := flag.String("addr", ":8080", "addr to listen to")
	flag.Parse()
	rand.Seed(time.Now().UTC().UnixNano())
	http.HandleFunc("/", handleAll())
	http.HandleFunc("/vote", handleVote)
	fmt.Println(http.ListenAndServe(*addr, nil))
}
