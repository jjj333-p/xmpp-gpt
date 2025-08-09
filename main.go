package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type config struct {
	PgUrl string `yaml:"pg_url"`
}

// ListPageData for the answered questions page
type ListPageData struct {
	User      string
	Questions []Question
}

// NewQuestionPageData for the new question page
type NewQuestionPageData struct {
	User            string
	CaptchaQuestion string
	CaptchaID       string
}

var captchaAnswers sync.Map

var questionListTmplt *template.Template
var newQuestionTmplt *template.Template

var qdb QuestionDB

func questionListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	user := r.PathValue("user")

	if user == "" {
		http.Error(w, "user needs to be provided", http.StatusBadRequest)
	}

	questions, err := qdb.GetByUserIDWithLimitAnswered(context.TODO(), user, 100, 0)

	data := ListPageData{
		User:      user,
		Questions: questions,
	}

	err = questionListTmplt.Execute(w, data)
	if err != nil {
		fmt.Println(err)
	}
}

func newQuestionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := NewQuestionPageData{
		User:            r.PathValue("user"),
		CaptchaQuestion: "lorem ipsum dolor sit amet,",
		CaptchaID:       "someUUIDhere",
	}

	err := newQuestionTmplt.Execute(w, data)
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	fmt.Println("Parsing templates...")
	var err error
	questionListTmplt, err = template.ParseFiles("./templates/questionsList.html")
	if err != nil {
		log.Fatal(err)
	}
	newQuestionTmplt, err = template.ParseFiles("./templates/newQuestion.html")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Reading config...")
	configData, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatal("error reading config file:", err)
	}

	var cfg config
	err = yaml.Unmarshal(configData, &cfg)
	if err != nil {
		log.Fatal("error parsing config:", err)
	}

	fmt.Println("connecting to database...")
	qdb.Init(cfg.PgUrl)

	http.HandleFunc("/{user}/", questionListHandler)
	http.HandleFunc("/{user}/new-query", newQuestionHandler)
	//http.HandleFunc("/", homeHandler)
	//http.HandleFunc("/new-query", newQueryHandler)
	//http.HandleFunc("/submit-question", submitQuestionHandler)
	//http.HandleFunc("/refresh-captcha", refreshCaptchaHandler)
	//http.HandleFunc("/question/", questionHandler)

	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
