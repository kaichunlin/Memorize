package db

import (
	dict "dictionary"
	"gopkg.in/mgo.v2/bson"
	"log"
	"shared"
)

type wordEntry struct {
	Type        dict.DictType
	Headword    string `bson:"headword"`
	Created     string
	SearchCount int `bson:"search_cnt"`
	NotifyCount int `bson:"notify_cnt"`
}

func AddWord(e *dict.WordEntry) error {
	sess, c := collection(dbWord)
	defer sess.Close()
	w := wordEntry{}
	key := bson.M{"$and": []bson.M{{"type": e.Type}, {"headword": e.Headword}}}
	c.Find(key).One(&w)
	update := true
	var err error
	if w.Headword == "" {
		w.Type = e.Type
		w.Headword = e.Headword
		w.Created = shared.CreationTime()
		w.SearchCount = 1
		if err = c.Insert(w); err != nil {
			c.Find(key).One(&w)
			log.Println(err)
			update = true
		} else {
			update = false
		}
	}
	if update {
		err = c.Update(key, bson.M{"$inc": bson.M{"search_cnt": 1}})
	}
	return err
}
