package main

import (
	"net/http"
	"sync"
	"html/template"
	"path/filepath"
	"log"
	"flag"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}
	t.templ.Execute(w, data)
}

func main() {
	var addr = flag.String("addr", ":8080", "アプリケーションのアドレス")
	var secKey = flag.String("seckey", "defaultKey", "セキュリティキー")
	var googleCId = flag.String("googleCId", "dummy", "GoogleのOAuthのclientId")
	var googleSecret = flag.String("googleSecret", "dummy", "google secret")
	var facebookCId = flag.String("facebookCId", "dummy", "FacebookのOAuthのclientId")
	var facebookSecret = flag.String("facebookSecret", "dummy", "Facebook Secret")
	var githubCId = flag.String("githubCId", "dummy", "GithubのOAuthのclientId")
	var githubSecret = flag.String("githubSecret", "dummy", "Github secret")
	flag.Parse()
	gomniauth.SetSecurityKey(*secKey)
	gomniauth.WithProviders(
		facebook.New(*facebookCId, *facebookSecret, "http://localhost:8080/auth/callback/facebook"),
		github.New(*githubCId, *githubSecret, "http://localhost:8080/auth/callback/github"),
		google.New(*googleCId, *googleSecret, "http://localhost:8080/auth/callback/google"),
	)
	r := newRoom(UseGravatar)
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler)
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:   "auth",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		w.Header()["Location"] = []string{"/chat"}
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
	http.Handle("/room", r)
	go r.run()
	log.Println("Webサーバを開始します。ポート：", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
