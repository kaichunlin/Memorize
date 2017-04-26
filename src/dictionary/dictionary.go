package dictionary

import (
	"bytes"
	"encoding/json"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	DictionaryError          = DictType(iota)
	DictionaryEnglishEnglish = DictType(iota)
	DictionaryEnglishChinese = DictType(iota)
)

type DictType int

type WordEntry struct {
	Type     DictType
	Headword string
	UserId   bson.ObjectId
}

type Def struct {
	Type              DictType   `json:"-"`
	Headword          string     `json:"headword"`
	Pron              string     `json:"pronunciation,omitempty"`
	Valid             bool       `json:"valid"`
	DefinitionEntries []DefEntry `json:"results,omitempty"`
	Err               error      `json:"-"`
}

func (d *Def) WordEntry() *WordEntry {
	return &WordEntry{Type: d.Type, Headword: d.Headword}
}
func (d *Def) UserWordEntry(uid bson.ObjectId) *WordEntry {
	return &WordEntry{UserId: uid, Type: d.Type, Headword: d.Headword}
}

func (d *Def) Html() string {
	var buffer bytes.Buffer
	buffer.WriteString("<u style='font-size:24px'>" + d.Headword + "</u><br>")
	if d.Pron != "" {
		buffer.WriteString("<i style='font-size:20px;margin-left:24px'>[")
		buffer.WriteString(d.Pron)
		buffer.WriteString("]</i>")
	}
	buffer.WriteString("<b style='margin-top:24px;font-size:18px'>Definitions</b><br>")
	buffer.WriteString("<ol style='margin-top:0'>")
	for _, de := range d.DefinitionEntries {
		buffer.WriteString("<li style='font-size:16px'>")
		buffer.WriteString(de.Definition)
		buffer.WriteString("</li>")
	}
	buffer.WriteString("</ol>")

	var hasEx bool
	for _, de := range d.DefinitionEntries {
		if de.Example != "" {
			hasEx = true
			break
		}
	}
	if hasEx {
		buffer.WriteString("<b style='margin-top:24px;font-size:16px'>Examples</b><br>")
		buffer.WriteString("<ol style='margin-top:0'>")
		for _, de := range d.DefinitionEntries {
			if de.Example == "" {
				continue
			}
			buffer.WriteString("<li style='font-size:16px'>")
			buffer.WriteString(de.Example)
			buffer.WriteString("</li>")
		}
		buffer.WriteString("</ol>")
	}

	return buffer.String()
}

type DefEntry struct {
	PartOfSpeech string   `json:"part_of_speech"`
	Definition   string   `json:"definition"`
	Synonyms     []string `json:"synonyms,omitempty"`
	Example      string   `json:"examples,omitempty"`
}

type pearsonDef struct {
	Word       string            `json:","`
	DefEntries []pearsonDefEntry `json:"results,omitempty"`
}
type pearsonDefEntry struct {
	Headword     string                 `json:","`
	PartOfSpeech string                 `json:"part_of_speech"`
	Senses       []pearsonDefEntrySense `json:"senses,omitempty"`
}
type pearsonDefEntrySense struct {
	Antonyms string            `json:"opposite"`
	Grammar  map[string]string `json:"gramatical_info"`
	Trans    string            `json:"translation"`
}
type wordsapiDef struct {
	Word       string             `json:","`
	Pron       map[string]string  `json:","`
	DefEntries []wordsapiDefEntry `json:"results,omitempty"`
}
type wordsapiDefEntry struct {
	PartOfSpeech string   `json:"partOfSpeech"`
	Def          string   `json:"definition"`
	Synonyms     []string `json:"synonyms,omitempty"`
	Exps         []string `json:"examples,omitempty"`
}

type owlDef struct {
	Type string `json:"type"`
	Def  string `json:"defenition"`
	Exp  string `json:"example,omitempty"`
}
type owlDefList []owlDef

var NullDef Def

func Search(w string, dicType DictType) *Def {
	switch dicType {
	case DictionaryEnglishEnglish:
		return SearchEnglish(w)
	case DictionaryEnglishChinese:
		return SearchEnglishToChinese(w)
	default:
		return &NullDef
	}
}

func SearchEnglishToChinese(w string) *Def {
	def := Def{Headword: w, Type: DictionaryEnglishChinese}
	if w == "" {
		return &def
	}
	r, err := searchPearson(w)
	if err != nil {
		def.Valid = false
		if _, ok := err.(*json.SyntaxError); !ok {
			def.Err = err
		}
	} else {
		if len(r.DefEntries) > 0 {
			def.Valid = true
		} else {
			def.Valid = false
		}
	}
	var min int
	min = minOf(len(r.DefEntries), 3)
	for i, d := range r.DefEntries {
		de := DefEntry{Definition: d.Senses[0].Trans, PartOfSpeech: d.PartOfSpeech}
		def.DefinitionEntries = append(def.DefinitionEntries, de)
		if i == min-1 {
			break
		}
	}
	return &def
}

func SearchEnglish(w string) *Def {
	def := Def{Headword: w, Type: DictionaryEnglishEnglish}
	if w == "" {
		return &def
	}
	r, err := searchWordsapi(w)
	extSearch := false
	if err != nil {
		def.Valid = false
		if _, ok := err.(*json.SyntaxError); !ok {
			def.Err = err
		}
		extSearch = true
	} else {
		def.Valid = true
		if len(r.DefEntries) == 0 {
			extSearch = true
		}
		def.Pron = r.Pron["all"]
	}
	var min int
	if extSearch {
		owl, err := searchOwlbot(w)
		if err == nil {
			if len(*owl) > 0 {
				def.Valid = true
			}
		} else {
			def.Err = err
			return &def
		}
		min = minOf(len(*owl), 3)
		for i, d := range *owl {
			de := DefEntry{Definition: d.Def, PartOfSpeech: d.Type}
			de.Example = d.Exp
			def.DefinitionEntries = append(def.DefinitionEntries, de)
			if i == min-1 {
				break
			}
		}
	} else {
		min = minOf(len(r.DefEntries), 3)
		for i, d := range r.DefEntries {
			de := DefEntry{Definition: d.Def, PartOfSpeech: d.PartOfSpeech}
			if len(d.Exps) > 0 {
				de.Example = d.Exps[0]
			}
			if len(d.Synonyms) > 0 {
				de.Synonyms = d.Synonyms
			}
			def.DefinitionEntries = append(def.DefinitionEntries, de)
			if i == min-1 {
				break
			}
		}
	}
	return &def
}

func minOf(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func searchOwlbot(w string) (*owlDefList, error) {
	url := "https://owlbot.info/api/v1/dictionary/" + w + "?format=json"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defs := owlDefList{}
	err = json.Unmarshal(body, &defs)
	return &defs, err
}

func searchWordsapi(w string) (*wordsapiDef, error) {
	url := "https://wordsapiv1.p.mashape.com/words/" + w
	req, err := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add("X-Mashape-Key", os.Getenv("KEY_WORDSAPI"))
	defs := wordsapiDef{}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &defs, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &defs, err
	}
	err = json.Unmarshal(body, &defs)
	return &defs, err
}

func searchPearson(w string) (*pearsonDef, error) {
	url := "http://api.pearson.com/v2/dictionaries/ldec/entries?headword=" + w + "&limit=5"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add("apikey", os.Getenv("KEY_PEARSON"))
	defs := pearsonDef{}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &defs, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &defs, err
	}
	err = json.Unmarshal(body, &defs)
	return &defs, err
}
