// +build integration

package api_tests

import (
	"fmt"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/gorilla/websocket"
	. "github.com/vivowares/octopus/Godeps/_workspace/src/github.com/smartystreets/goconvey/convey"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/verdverm/frisby"
	. "github.com/vivowares/octopus/connections"
	. "github.com/vivowares/octopus/models"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func ApiSendToDevicePath(chId, deviceId string) string {
	return fmt.Sprintf("%s/channels/%s/devices/%s/send", ApiServer, chId, deviceId)
}

func ApiRequestToDevicePath(chId, deviceId string) string {
	return fmt.Sprintf("%s/channels/%s/devices/%s/request", ApiServer, chId, deviceId)
}

func TestApiToDevice(t *testing.T) {

	InitializeDB()
	DB.LogMode(true)
	DB.SetLogger(log.New(os.Stdout, "", log.LstdFlags))
	DB.DropTableIfExists(&Channel{})
	DB.AutoMigrate(&Channel{})

	InitializeIndexClient()

	chId, ch := CreateTestChannel()

	Convey("successfully send data to device from api", t, func() {
		deviceId := "abc"
		cli := CreateWsConnection(chId, deviceId, ch)

		message := "this is a test message"
		var rcvData []byte
		var rcvMsgType int
		var rcvErr error
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			cli.SetReadDeadline(time.Now().Add(2 * time.Second))
			rcvMsgType, rcvData, rcvErr = cli.ReadMessage()
			wg.Done()
		}()

		f := frisby.Create("send message to device").Post(ApiSendToDevicePath(chId, deviceId)).
			SetHeader("AuthToken", authStr()).SetJson(map[string]string{"test": message}).Send()
		f.ExpectStatus(http.StatusOK)

		So(rcvErr, ShouldBeNil)
		So(rcvMsgType, ShouldEqual, websocket.BinaryMessage)
		strs := strings.Split(string(rcvData), "|")
		So(strs[len(strs)-1], ShouldEqual, fmt.Sprintf("{\"test\":\"%s\"}", message))
		So(strs[0], ShouldEqual, strconv.Itoa(TypeSendMessage))

		wg.Wait()
		cli.Close()
	})

	Convey("successfully request data from device", t, func() {
		deviceId := "abc"
		cli := CreateWsConnection(chId, deviceId, ch)

		reqMsg := "request message"
		respMsg := "response message"
		var rcvMessage string
		var rcvData []byte
		var rcvMsgType int
		var rcvErr error
		var sendMsgType int
		go func() {
			cli.SetReadDeadline(time.Now().Add(2 * time.Second))
			rcvMsgType, rcvData, rcvErr = cli.ReadMessage()
			msg, _ := Unmarshal(rcvData)
			rcvMessage = string(msg.Payload)
			sendMsgType = msg.MessageType
			msg = &Message{
				MessageType: TypeResponseMessage,
				MessageId:   msg.MessageId,
				Payload:     []byte(respMsg),
			}
			p, _ := msg.Marshal()
			cli.WriteMessage(websocket.BinaryMessage, p)
		}()

		f := frisby.Create("send message to device").Post(ApiRequestToDevicePath(chId, deviceId)).
			SetHeader("AuthToken", authStr()).SetJson(map[string]string{"test": reqMsg}).Send()
		f.ExpectStatus(http.StatusOK).
			AfterContent(func(F *frisby.Frisby, content []byte, err error) {
			So(string(content), ShouldEqual, respMsg)
		})

		So(rcvErr, ShouldBeNil)
		So(rcvMsgType, ShouldEqual, websocket.BinaryMessage)
		So(rcvMessage, ShouldEqual, `{"test":"request message"}`)
		So(sendMsgType, ShouldEqual, TypeRequestMessage)

		cli.Close()
	})

	frisby.Global.PrintReport()
}