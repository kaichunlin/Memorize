package db

import (
	dict "dictionary"
	"gopkg.in/mgo.v2/bson"
	"shared"
)

type User struct {
	Id        bson.ObjectId `bson:"_id,omitempty"`
	Email     string
	Created   string
	Verified  bool
	WordCount int `bson:"word_count"`
}

type UserWord struct {
	Id             bson.ObjectId `bson:"_id,omitempty"`
	Type           dict.DictType
	State          int
	Headword       string `bson:"headword"`
	Created        string
	NextNotifyDate string        `bson:"next_notify"`
	UserId         bson.ObjectId `bson:"user_id"`
}

func (u *UserWord) PrevNotify() int {
	return shared.CalculateDay(u.State - 1)
}

func (u *UserWord) NextNotify() int {
	return shared.CalculateDay(u.State + 1)
}

func (u *UserWord) CurrentNotify() int {
	return shared.CalculateDay(u.State)
}

func userById(uid bson.ObjectId) *User {
	sess, c := collection(dbUser)
	defer sess.Close()

	u := User{}
	key := bson.M{"_id": uid}
	c.Find(key).One(&u)
	return &u
}

func AddUser(email string) (User, error) {
	sess, c := collection(dbUser)
	defer sess.Close()
	u := User{}
	key := bson.M{"email": email}
	c.Find(key).One(&u)
	if u.Email == "" {
		u.Email = email
		u.Created = shared.CreationTime()
		c.Insert(u)
		c.Find(key).One(&u)
	}
	return u, nil
}

func HasUserWord(email string, headword string, dictType dict.DictType) (bool, error) {
	sess, c := collection(dbUser)
	defer sess.Close()
	u := User{}
	key := bson.M{"email": email}
	r := c.Find(key)
	count, err := r.Count()
	if err != nil {
		return false, err
	}
	if count == 0 {
		return false, nil
	}
	if err := c.Find(key).One(&u); err != nil {
		return false, err
	}
	sess2, c2 := collection(dbUserWord)
	defer sess2.Close()
	if count, err := c2.Find(bson.M{"$and": []bson.M{{"user_id": u.Id}, {"headword": headword}, {"type": dictType}}}).Count(); err != nil {
		return false, err
	} else {
		return count > 0, nil
	}
}

func UserEmail(uid bson.ObjectId) string {
	sess, c := collection(dbUser)
	defer sess.Close()

	u := User{}
	c.Find(bson.M{"_id": uid}).One(&u)
	return u.Email
}

func UpdateNextNotifyTime(uw *UserWord) error {
	sess, c := collection(dbUserWord)
	defer sess.Close()

	uw.State++
	if uw.State <= shared.NotifyOffsetCount() {
		uw.NextNotifyDate = shared.NextNotifyTime(uw.State, uw.Created)
		return c.Update(bson.M{"_id": uw.Id}, bson.M{"$set": bson.M{"state": uw.State, "next_notify": uw.NextNotifyDate}})
	} else {
		return c.Update(bson.M{"_id": uw.Id}, bson.M{"$set": bson.M{"state": uw.State}})
	}
}

func findUserWord(e *dict.WordEntry) (*UserWord, error) {
	sess, c := collection(dbUserWord)
	defer sess.Close()

	q := c.Find(bson.M{"$and": []bson.M{{"user_id": e.UserId}, {"type": e.Type}, {"headword": e.Headword}}})
	count, err := q.Count()
	if count == 0 {
		return nil, err
	} else {
		uw := UserWord{}
		err = q.One(&uw)
		return &uw, err
	}
}

func AddUserWord(e *dict.WordEntry) (bool, *UserWord, error) {
	var err error

	u := userById(e.UserId)
	if u.Email == "" {
		return false, nil, nil
	}

	if err = AddWord(e); err != nil {
		return false, nil, err
	}

	uw, err := findUserWord(e)
	if err != nil {
		return false, nil, err
	}
	if uw != nil {
		return false, uw, nil
	}

	t := shared.CreationTime()
	uw = &UserWord{
		Headword: e.Headword,
		Type:     e.Type,
		Created:  t,
		State:    1,
		UserId:   u.Id,
	}
	uw.NextNotifyDate = shared.NextNotifyTime(uw.State, t)
	sess, c := collection(dbUserWord)
	defer sess.Close()
	if err = c.Insert(uw); err == nil {
		err = updateWordCount(u)
	} else {
		return false, nil, err
	}

	return true, uw, nil
}

func updateWordCount(u *User) error {
	sess, c := collection(dbUser)
	defer sess.Close()

	key := bson.M{"_id": u.Id}
	return c.Update(key, bson.M{"$inc": bson.M{"word_count": 1}})
}
