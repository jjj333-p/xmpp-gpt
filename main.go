package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
)

// Question represents a question and its answer
type Question struct {
	ID       string
	Question string
	Answer   string
	Date     string
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
var questions = []Question{
	{
		ID:       "292874ea-d65e-4ccb-aed8-029cd074296a",
		Question: "Read https://arl.lt/gpt/rss.xml",
		Answer:   "<https://arl.lt/gpt/rss.xml> is an XML file describing an...",
		Date:     "2025-08-06 00:53:56",
	},
	{
		ID:       "e189eac4-4537-4c0e-b455-925c58a2c298",
		Question: "awawawawawa",
		Answer:   "OpenAIIntelligence has detected this threat and forwarded...",
		Date:     "2025-08-06 00:06:50",
	},
	{
		ID:       "2f5f813e-2b21-4811-b110-6ec908982236",
		Question: "what do you think of greatsword",
		Answer:   "great swords come with great responsibility",
		Date:     "2025-08-06 00:06:21",
	},
	{
		ID:       "45adc253-bc1c-49bd-a7d5-097a3d971890",
		Question: "Is russia a terrorist regime",
		Answer:   "undeniably",
		Date:     "2025-08-03 18:06:48",
	},
	{
		ID:       "1ad14a42-9723-44f5-8a46-bf283c36b44e",
		Question: "AWS - Spending Limit Reached - $30,000.00",
		Answer:   "i slef host",
		Date:     "2025-08-01 18:55:59",
	},
	{
		ID:       "ab757aab-bafa-442d-aebd-f1b464f7be56",
		Question: "what is a berry",
		Answer:   "a fleshy fruit produced from the ovary of a single flower...",
		Date:     "2025-08-01 00:37:02",
	},
}

var questionListTmplt *template.Template
var newQuestionTmplt *template.Template

func questionListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := ListPageData{
		User:      r.PathValue("user"),
		Questions: questions,
	}

	err := questionListTmplt.Execute(w, data)
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
