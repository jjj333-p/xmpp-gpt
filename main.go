package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/google/uuid"
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

func submitQuestionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	user := r.PathValue("user")
	question := r.FormValue("question")
	captchaID := r.FormValue("captcha_id")
	captchaAnswer := r.FormValue("captcha_answer")

	fmt.Println(captchaID, captchaAnswer)
	//// Verify captcha
	//if expectedAnswer, ok := captchaAnswers.Load(captchaID); !ok || expectedAnswer != captchaAnswer {
	//	http.Error(w, "Invalid captcha answer", http.StatusBadRequest)
	//	return
	//}
	//captchaAnswers.Delete(captchaID)

	id := uuid.New().String()

	// Save question to database
	err = qdb.CreateWithID(context.TODO(), id, user, question)
	if err != nil {
		errStr := fmt.Sprintf("Error creating question %s: %v\n", question, err)
		fmt.Println(errStr)
		http.Error(w, errStr, http.StatusInternalServerError)
		return
	}

	// Redirect back to the user's question list
	http.Redirect(w, r, "/"+user+"/", http.StatusSeeOther)
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
	http.HandleFunc("/{user}/submit-question", submitQuestionHandler)
	//http.HandleFunc("/", homeHandler)
	//http.HandleFunc("/new-query", newQueryHandler)
	//http.HandleFunc("/refresh-captcha", refreshCaptchaHandler)
	//http.HandleFunc("/question/", questionHandler)

	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
