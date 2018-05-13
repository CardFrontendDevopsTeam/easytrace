package easytraceCache

import (
	"github.com/weAutomateEverything/go2hal/remoteTelegramCommands"
	"time"
	"golang.org/x/net/context"
	"log"
	"github.com/weAutomateEverything/go2hal/alert"
	"github.com/weAutomateEverything/go2hal/chef"
	"github.com/weAutomateEverything/go2hal/telegram"
	"strconv"
	"net/http"
	"io/ioutil"
	"os"
)

type Service interface {
}

type service struct {
	client        remoteTelegramCommands.RemoteCommandClient
	alertService  alert.Service
	chefService   chef.Service
	chefStore     chef.Store
	telegramStore telegram.Store
}

func NewService(client remoteTelegramCommands.RemoteCommandClient, alert alert.Service, chefService chef.Service, chefStore chef.Store, telegramStore telegram.Store) Service {
	s := &service{client, alert, chefService, chefStore, telegramStore}
	go func() {
		s.registerRemoteStream()
	}()
	go func() {
		s.RemoteRecipeReply()

	}()
	go func() {
		s.RemoteEnvironmentReply()

	}()
	go func() {
		s.RemoteNodeReply()

	}()
	return s
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

func (s *service) RemoteRecipeReply() {
	for {
		request := remoteTelegramCommands.Request{Nextstate: "SEARCH_CHEF_ENVIRONMENT", State: "SEARCH_CHEF"}
		stream1, err := s.client.RegisterCommandLet(context.Background(), &request)
		if err != nil {
			log.Println(err)
		} else {
			s.monitorRecipeReply(stream1)
		}
		time.Sleep(30 * time.Second)
	}
}
func (s *service) RemoteEnvironmentReply() {
	for {
		request := remoteTelegramCommands.Request{Nextstate: "SEARCH_CHEF_NODE", State: "SEARCH_CHEF_ENVIRONMENT"}
		stream2, err := s.client.RegisterCommandLet(context.Background(), &request)
		if err != nil {
			log.Println(err)
		} else {
			s.monitorEnvironmentReply(stream2)
		}
		time.Sleep(30 * time.Second)
	}
}
func (s *service) RemoteNodeReply() {
	for {
		request := remoteTelegramCommands.Request{Nextstate: "EXECUTE", State: "SEARCH_CHEF_NODE"}
		stream3, err := s.client.RegisterCommandLet(context.Background(), &request)
		if err != nil {
			log.Println(err)
		} else {
			s.monitorNodeReply(stream3)
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
		i, err := strconv.Atoi(in.From)
		s.telegramStore.SetState(i, "SEARCH_CHEF", nil)
		s.chefService.SendKeyboardRecipe(context.TODO(), "Please select the application")

	}
}

func (s *service) monitorRecipeReply(client remoteTelegramCommands.RemoteCommand_RegisterCommandLetClient) {
	for {
		in, err := client.Recv()
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(in.Message)
		s.chefService.SendKeyboardEnvironment(context.TODO(), "Please select the environment")

	}
}
func (s *service) monitorEnvironmentReply(client remoteTelegramCommands.RemoteCommand_RegisterCommandLetClient) {
	for {
		in, err := client.Recv()
		log.Println(in.Message)
		if err != nil {
			log.Println(err)
			return
		}
		s.chefService.SendKeyboardNodes(context.TODO(), in.Fields[0], in.Message, "Please select the node")
	}
}
func (s *service) monitorNodeReply(client remoteTelegramCommands.RemoteCommand_RegisterCommandLetClient) {
	for {
		in, err := client.Recv()
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(getProtocol() + "://" + in.Message + "." + getDomain() + ":" + getPort() + getRestPath())
		response, errresp := http.Get(getProtocol() + "://" + in.Message + "." + getDomain() + ":" + getPort() + getRestPath())
		if errresp != nil {
			s.alertService.SendError(context.TODO(), errresp)
		}
		responseData, err := ioutil.ReadAll(response.Body)
		if err != nil {
			s.alertService.SendError(context.TODO(), err)
		}
		s.alertService.SendAlert(context.TODO(), string(responseData)+" from "+in.Message)
	}
}
func getDomain() string {
	return os.Getenv("DOMAIN")
}
func getPort() string {
	return os.Getenv("PORT")
}
func getRestPath() string {
	return os.Getenv("REST_PATH")
}
func getProtocol() string {
	return os.Getenv("PROTOCOL")
}
