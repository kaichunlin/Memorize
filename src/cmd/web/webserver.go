package main

import (
	"db"
	dict "dictionary"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"html/template"
	"log"
	"notify"
	"os"
	"regexp"
)

type result string

const (
	resultSuccess       = result("Success")
	resultInvalidInputs = result("InvalidInputs")
	resultAlreadyExist  = result("AlreadyExist")
	resultUserNotFound  = result("UserNotFound")
	resultWordNotFound  = result("WordNotFound")
	resultError         = result("Error")
)

type dictType string

func (dt dictType) toType() dict.DictType {
	switch dt {
	case "enen":
		return dict.DictionaryEnglishEnglish
	case "ench":
		return dict.DictionaryEnglishChinese
	default:
		return dict.DictionaryError
	}
}

type serverReply struct {
	httpCode int
	Result   result   `json:"result"`
	Error    string   `json:"error"`
	DictType dictType `json:"dict_type"`
	Created  string   `json:"created"`
	dict.Def
}

func index(render render.Render) {
	render.HTML(200, "index", map[string]interface{}{"Type": "enen"})
}

func showWordDef(params martini.Params, render render.Render) {
	dictType := params["dict_type"]
	headword := params["headword"]
	r := queryWordDef(headword, dictType)
	if r.Valid {
		render.HTML(200, "index", map[string]interface{}{
			"Content":        template.HTML(r.Def.Html()),
			"Type":           dictType,
			"Headword":       headword,
			"MemorizeButton": template.HTML("<button id='add' class='btn dropdown btn-default btn-primary center-col' data-toggle='modal' data-target='#memorize-modal'>Memorize</button>"),
		})
	} else {
		render.HTML(200, "index", map[string]interface{}{
			"Content": template.HTML("<h4 class='center-text'>No definition found :(</h4>"),
			"Type":           dictType,
			"Headword":       headword,
		})
	}
}

func sanitizeHeadword(headword string) string {
	reg, err := regexp.Compile("[^a-zA-Z0-9 \\-]+")
	if err != nil {
		log.Fatal(err)
	}
	return reg.ReplaceAllString(headword, "")
}

func addUserWord(params martini.Params, render render.Render) {
	r := serverReply{httpCode: 400}
	if params["user_email"] == "" || params["dict_type"] == "" || params["headword"] == "" {
		r.Result = resultInvalidInputs
		render.JSON(r.httpCode, r)
		return
	}
	email := params["user_email"]
	r.DictType = dictType(params["dict_type"])
	headword := sanitizeHeadword(params["headword"])
	dt := r.DictType.toType()
	log.Println("addUserWord params:", headword, email, r.DictType)
	if dt == dict.DictionaryError {
		log.Println("addUserWord :resultInvalidInputs")
		r.Result = resultInvalidInputs
		render.JSON(r.httpCode, r)
		return
	}

	if exists, err := db.HasUserWord(email, headword, dt); err == nil {
		if exists {
			r.httpCode = 200
			r.Headword = headword
			r.Result = resultAlreadyExist
			render.JSON(r.httpCode, r)
			return
		}
	} else {
		r.Err = err
		r.Error = err.Error()
		render.JSON(r.httpCode, r)
		return
	}

	def := dict.Search(headword, dt)
	if def.Valid {
		user, err := db.AddUser(email)
		if err != nil {
			log.Println(err)
		}
		added, uw, err := db.AddUserWord(def.UserWordEntry(user.Id))
		if added {
			r.httpCode = 200
			r.Result = resultSuccess
			r.Created = uw.Created
			r.Valid = true
			r.Headword = uw.Headword
			notify.SendWordReminder(email, uw.PrevNotify(), uw.CurrentNotify(), def)
		} else {
			if err == nil {
				r.Result = resultUserNotFound
			} else {
				r.Result = resultError
				r.Error = err.Error()
			}
		}
	} else {
		r.Result = resultWordNotFound
	}

	if r.Err != nil {
		r.Error = r.Err.Error()
	}
	render.JSON(r.httpCode, r)
}

func queryWordDef(h string, t string) (r serverReply) {
	h = sanitizeHeadword(h)
	r.httpCode = 400
	r.Headword = h
	r.DictType = dictType(t)
	dt := r.DictType.toType()
	if dt == dict.DictionaryError {
		r.Result = resultInvalidInputs
		return
	}
	def := dict.Search(r.Headword, dt)
	if def.Valid {
		r.Result = resultSuccess
		r.httpCode = 200
		r.Def = *def
	} else {
		if def.Err == nil {
			r.httpCode = 404
			r.Result = resultWordNotFound
		} else {
			r.Result = resultError
			r.Error = def.Err.Error()
		}
	}

	return
}

func getWordDef(params martini.Params, render render.Render) {
	r := queryWordDef(params["headword"], params["dict_type"])
	render.JSON(r.httpCode, r)
}

func main() {
	m := martini.Classic()
	// render html templates from templates directory
	m.Use(render.Renderer(render.Options{
		Extensions: []string{".tmpl", ".html"},
	}))

	if err := db.Init(); err != nil {
		log.Println("db.InitDb error:", err)
	}

	logger := log.New(os.Stdout, "logger: ", log.Lshortfile)
	m.Logger(logger)

	staticOptions := martini.StaticOptions{Prefix: "static"}
	m.Use(martini.Static("static", staticOptions))

	staticOptions = martini.StaticOptions{Prefix: ""}
	m.Use(martini.Static("root", staticOptions))

	m.Get("/", index)
	m.Get("/def/:dict_type/:headword", showWordDef)
	m.Get("/api/def/:dict_type/:headword", getWordDef)
	m.Post("/api/add/:user_email/:dict_type/:headword", addUserWord)

	m.Run()
}
