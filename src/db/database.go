package db

import (
	"gopkg.in/mgo.v2"
	"os"
)

const (
	dbUser         string = "user"
	dbUserWord     string = "user-headword"
	dbWord         string = "headword"
	dbNotification string = "notification"
)

func Init() error {
	uri := os.Getenv("MONGODB_URI")
	sess, err := mgo.Dial(uri)
	if err != nil {
		panic(err)
	}
	names, err := sess.DB(os.Getenv("DB_NAME")).CollectionNames()
	if err != nil {
		return err
	}

	inited := map[string]bool{}
	for _, c := range names {
		inited[c] = true
	}

	if !inited[dbUser] {
		sess, c := collection(dbUser)
		defer sess.Close()
		err = c.EnsureIndex(mgo.Index{
			Key:    []string{"email"},
			Unique: true,
		})
		if err != nil {
			return err
		}
	}
	if !inited[dbUserWord] {
		sess, c := collection(dbUserWord)
		defer sess.Close()
		err = c.EnsureIndex(mgo.Index{
			Key:    []string{"type", "headword", "user_id"},
			Unique: true,
		})
		if err != nil {
			return err
		}
	}
	if !inited[dbWord] {
		sess, c := collection(dbWord)
		defer sess.Close()
		err = c.EnsureIndex(mgo.Index{
			Key:    []string{"type", "headword"},
			Unique: true,
		})
		if err != nil {
			return err
		}
	}
	if !inited[dbNotification] {
		sess, c := collection(dbNotification)
		defer sess.Close()
		err = c.EnsureIndex(mgo.Index{
			Key:    []string{"time"},
			Unique: true,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func collection(name string) (*mgo.Session, *mgo.Collection) {
	uri := os.Getenv("MONGODB_URI")
	sess, err := mgo.Dial(uri)
	if err != nil {
		panic(err)
	}
	return sess, sess.DB(os.Getenv("DB_NAME")).C(name)
}
