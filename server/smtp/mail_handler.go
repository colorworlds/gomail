package smtp

import (
	"encoding/json"
	"gomail/server/db"
	"gomail/server/response"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"log"
	"net/http"
)

type MailHandle struct {
	Client *MailClient
	Db     *db.Client
}

func (mh *MailHandle) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var jsonData []byte
	var task = MailTask{}
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
			err := r.(error)
			_, _ = writer.Write(response.Fail(1, err.Error()))
		}
	}()
	writer.Header().Add("Content-Type", "application/json")
	jsonData, _ = ioutil.ReadAll(request.Body)
	err := json.Unmarshal(jsonData, &task)
	if err != nil {
		panic(err)
	}
	if task.Attachment.WithFile {
		file, err := mh.Db.Download(bson.ObjectIdHex(task.Attachment.Id))
		if err != nil {
			panic(err)
		}
		task.Attachment.ContentType = file.ContentType()
		task.Attachment.Name = file.Name()
		task.Attachment.Reader = file
	}

	MessageId, err := mh.Client.Send(task)
	if err != nil {
		panic(err)
	}
	_, _ = writer.Write(response.Success(MessageId))
}