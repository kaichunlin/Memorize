package notify

import (
	"db"
	dict "dictionary"
	"fmt"
	"log"
	"shared"
	"time"
)

type NotifyResult struct {
	Notified     bool
	SuccessCount int
	FailCount    int
}

type key struct {
	Type     dict.DictType
	Headword string
}

func Notify() (NotifyResult, error) {
	return NotifyForTime(shared.CurrentNotifyTime())
}

func NotifyForTime(t string) (result NotifyResult, err error) {
	nr := db.NotificationRecord{}
	start := time.Now()
	nr.Time = shared.CreationTime()
	nr.NotifyTime = t

	notified, err := db.HasNotifiedAt(t)
	if err != nil || notified {
		return
	}
	notifList, err := db.NotificationList(nr.NotifyTime)
	if err != nil {
		return
	}

	words := map[key][]db.UserWord{}
	for _, uw := range notifList {
		k := key{Type: uw.Type, Headword: uw.Headword}
		words[k] = append(words[k], uw)
	}
	defs := map[key]*dict.Def{}
	for k := range words {
		defs[k] = dict.Search(k.Headword, k.Type)
	}
	for k, uwList := range words {
		for _, uw := range uwList {
			userEmail := db.UserEmail(uw.UserId)
			if userEmail == "" {
				log.Println(fmt.Sprintf("[warning] NotifyForTime: user_id %v not found", uw.UserId.Hex()))
			} else {
				//TODO add end state
				log.Println("NotifyForTime: UserWord ID =", uw.Id.Hex())
				err = SendWordReminder(userEmail, uw.CurrentNotify(), uw.NextNotify(), defs[k])
				if err == nil {
					result.SuccessCount++
					nr.SuccessCount++
				} else {
					result.FailCount++
					nr.FailCount++
				}
				if err = db.UpdateNextNotifyTime(&uw); err != nil {
					log.Println(fmt.Sprintf("[warning] NotifyForTime: user word update failed: %v", uw.Id.Hex()))
				}
			}
		}
	}

	nr.Duration = int(time.Now().Unix() - start.Unix())
	nr.UniqueWords = len(defs)
	if err = db.AddNotificationRecord(&nr); err != nil {
		log.Println(fmt.Sprintf("[warning] NotifyForTime: notify result not saved: %v, err = %v", nr, err))
	}

	result.Notified = true
	return
}
