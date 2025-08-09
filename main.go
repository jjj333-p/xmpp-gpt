package main

import (
	"context"
	"crypto/rand"
	"encoding/csv"
	"fmt"
	"html/template"
	"log"
	"math/big"
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

type captcha struct {
	Question string
	Answer   string
}

var captchaAnswers sync.Map

func randomCaptcha() captcha {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(captchas))))
	if err != nil {
		log.Printf("Error generating random number: %v", err)
		return captchas[0]
	}
	return captchas[n.Int64()]
}

var questionListTmplt *template.Template
var newQuestionTmplt *template.Template

var qdb QuestionDB

var captchas []captcha

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

	selectedCaptcha := randomCaptcha()
	captchaID := uuid.New().String()
	captchaAnswers.Store(captchaID, selectedCaptcha.Answer)

	data := NewQuestionPageData{
		User:            r.PathValue("user"),
		CaptchaQuestion: selectedCaptcha.Question,
		CaptchaID:       captchaID,
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
	// Verify captcha
	if expectedAnswer, ok := captchaAnswers.Load(captchaID); !ok || expectedAnswer != captchaAnswer {
		http.Error(w, "Invalid captcha answer", http.StatusBadRequest)
		return
	}
	captchaAnswers.Delete(captchaID)

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

	fmt.Println("Loading captchas...")
	file, err := os.Open("./captcha.csv")
	if err != nil {
		log.Fatalf("error opening captchas file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("error reading captchas: %v", err)
	}

	for _, record := range records {
		if len(record) == 2 {
			captchas = append(captchas, captcha{
				Question: record[0],
				Answer:   record[1],
			})
		}
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
