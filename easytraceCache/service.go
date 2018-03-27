package easytraceCache

import (
	"github.com/weAutomateEverything/go2hal/remoteTelegramCommands"
	"time"
	"golang.org/x/net/context"
	"log"
	"github.com/weAutomateEverything/go2hal/alert"
)

type Service interface {
	reloadCache() (string)
}

type service struct {
	client       remoteTelegramCommands.RemoteCommandClient
	alertService alert.Service
}

func NewService(client remoteTelegramCommands.RemoteCommandClient, alert alert.Service) Service {
	s := &service{client, alert}
	go func() {
		s.registerRemoteStream()
	}()
	return s
}

func (s *service) reloadCache() (string) {
	return "hello"
}
func (s *service) registerRemoteStream() {
	for {
		request := remoteTelegramCommands.RemoteCommandRequest{Description: "Clear EasytraceCache", Name: "ReloadCache"}
		stream, err := s.client.RegisterCommand(context.Background(), &request)
		if err != nil {
			log.Println(err)
		} else {
			s.monitorForStreamResponse(stream)
		}
		time.Sleep(30 * time.Second)
	}
}
func (s *service) monitorForStreamResponse(client remoteTelegramCommands.RemoteCommand_RegisterCommandClient) {
	for {
		in, err := client.Recv()
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(in.From)
		s.alertService.SendAlertKeyboard(context.TODO(), "Please select the application")
	}
}
