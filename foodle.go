package main

import (
	"flag"
	"fmt"
	"html/template"
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
				font-size: 30px;
				background-color: #dddddd;
				font-family: sans-serif;
			}

			table {
				border-collapse: collapse;
				border: 1px solid;
				font-size: 35px;
				width: auto;
			}

			th, td {
				text-align: left;
				padding: 8px;
			}

			input {
				font-size: 30px;
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
				<input name="name" type="text" {{if $.Name}}value="{{$.Name}}"{{end}}placeholder="your name"><br>
				<input name="food" type="search" list="restaurants" placeholder="your fav food"><br>
				<datalist id="restaurants">
					<option value="Pizza-Scheune" />
					<option value="Thai" />
					<option value="Pho Co" />
					<option value="Soy" />
					<option value="Pop Chicken" />
					<option value="Kiez Falafel" />
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

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt('A', 'Z'+1))
	}
	return string(bytes)
}

type Context struct {
	votes map[string]string
	users map[string]string
}

func initContext() *Context {
	return &Context{
		votes: make(map[string]string),
		users: make(map[string]string),
	}
}

func (context *Context) handleVote(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	food := r.URL.Query().Get("food")
	queryToken := r.URL.Query().Get("token")
	if cookieToken, err := r.Cookie("token"); err != nil || queryToken != cookieToken.Value {
		w.WriteHeader(403)
		w.Write([]byte("CSRF"))
		return
	}

	if len(name) > 0 && len(food) > 0 {
		if len(context.users[name]) > 0 {
			secretCookie, err := r.Cookie("secret")
			if err != nil || len(secretCookie.Value) == 0 || secretCookie.Value != context.users[name] {
				w.WriteHeader(403)
				w.Write([]byte("SECRET"))
				return
			}
		} else {
			secret := randomString(32)
			context.users[name] = secret
			http.SetCookie(w, &http.Cookie{Name: "name", Value: name, HttpOnly: true})
			http.SetCookie(w, &http.Cookie{Name: "secret", Value: secret, HttpOnly: true})
		}
		context.votes[name] = food
		http.Redirect(w, r, "/", 302)
		return
	}
}

func (context *Context) handleAll(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.Header.Get("Accept"), "text/html") {
		return //prevent favicon.ico from messing with cookies
	}
	t, err := template.New("foodle").Parse(foodle)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
	lastRequest := time.Now()
	if lastRequest.Add(time.Hour * 8).Before(time.Now()) {
		context.votes = make(map[string]string)
	}
	lastRequest = time.Now()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Frame-Options", "DENY")
	name := ""
	if nameCookie, err := r.Cookie("name"); err == nil && len(nameCookie.Value) > 0 {
		name = nameCookie.Value
	}
	token := randomString(32)
	http.SetCookie(w, &http.Cookie{Name: "token", Value: token, HttpOnly: true})
	res := Result{Name: name, Votes: context.votes, MostUsed: getMostUsedValue(context.votes), Token: token}
	if err = t.ExecuteTemplate(w, "T", res); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}

func main() {
	listen := flag.String("http", ":8080", "http url to listen to")
        flag.Parse()
	rand.Seed(time.Now().UTC().UnixNano())
	foodleContext := initContext()
	http.HandleFunc("/", foodleContext.handleAll)
	http.HandleFunc("/vote", foodleContext.handleVote)
	fmt.Println(http.ListenAndServe(*listen, nil))
}
