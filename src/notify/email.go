package notify

import (
	dict "dictionary"
	"fmt"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type Email struct {
	To       string
	Subject  string
	Headword string
	Content  string
	Today    int
	Next     int
}

func SendWordReminder(email string, today int, next int, def *dict.Def) error {
	e := Email{To: email, Today: today, Next: next, Subject: "Memorize It: {{ .Headword }} (Day {{ .Today }})", Headword: def.Headword, Content: def.Html()}
	return sendSendgridEmail(e)
}

func sendSendgridEmail(e Email) error {
	e.Subject = strings.Replace(e.Subject, "{{ .Headword }}", e.Headword, 1)
	e.Subject = strings.Replace(e.Subject, "{{ .Today }}", strconv.Itoa(e.Today), 1)
	emailBuf, _ := ioutil.ReadFile("templates/email/email.html")
	cnt := string(emailBuf)
	cnt = strings.Replace(cnt, "{{ .Headword }}", e.Headword, 1)
	cnt = strings.Replace(cnt, "{{ .Content }}", e.Content, 1)
	cnt = strings.Replace(cnt, "{{ .Next }}", strconv.Itoa(e.Next), 1)

	from := mail.NewEmail("Memorize It", "learn@memorize-it.com")
	to := mail.NewEmail("", e.To)
	c := mail.NewContent("text/html", cnt)
	m := mail.NewV3MailInit(from, e.Subject, to, c)

	request := sendgrid.GetRequest(os.Getenv("SENDGRID_API_KEY"), "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	request.Body = mail.GetRequestBody(m)
	response, err := sendgrid.API(request)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
	return err
}
