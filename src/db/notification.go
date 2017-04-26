package db

import (
	"gopkg.in/mgo.v2/bson"
)

type NotificationRecord struct {
	Time         string
	NotifyTime   string `bson:"notify_time"`
	SuccessCount int    `bson:"success_cnt"`
	FailCount    int    `bson:"fail_cnt"`
	UniqueWords  int    `bson:"uniques"`
	Duration     int
}

//TODO Doesn't check for missed notification
func HasNotifiedAt(t string) (bool, error) {
	sess, c := collection(dbNotification)
	defer sess.Close()

	key := bson.M{"notified_time": t}
	count, err := c.Find(key).Count()
	return count > 0, err
}

func AddNotificationRecord(n *NotificationRecord) error {
	sess, c := collection(dbNotification)
	defer sess.Close()

	err := c.Insert(n)
	return err
}

func NotificationList(t string) (uw []UserWord, err error) {
	sess, c := collection(dbUserWord)
	defer sess.Close()

	//Doesn't check for missed notification
	key := bson.M{"next_notify": t}
	err = c.Find(key).All(&uw)
	return
}
