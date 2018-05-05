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
)

type Service interface {
	reloadCache() (string)
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
		s.registerRemoteStream1()
		//s.registerCommandLetRemoteStream("SEARCH_CHEF", "SEARCH_CHEF_ENVIRONMENT")
	}()
	go func() {
		s.registerRemoteStream2()
		//s.registerCommandLetRemoteStream("SEARCH_CHEF_ENVIRONMENT", "SEARCH_CHEF_NODE")
	}()
	go func() {
		s.registerRemoteStream3()
		//s.registerCommandLetRemoteStream("SEARCH_CHEF_NODE", "EXECUTE")
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

func (s *service) registerRemoteStream1() {
	for {
		request := remoteTelegramCommands.Request{Nextstate: "SEARCH_CHEF_ENVIRONMENT", State: "SEARCH_CHEF"}
		stream1, err := s.client.RegisterCommandLet(context.Background(), &request)
		//stream1,err :=s.client.RegisterCommandLet(context.Background(),&request)
		if err != nil {
			log.Println(err)
		} else {
			s.monitorForStreamResponse1(stream1)
		}
		time.Sleep(30 * time.Second)
	}
}

/*func (s *service) registerCommandLetRemoteStream(state string, nextstate string) {
	for {
		request := remoteTelegramCommands.Request{Description: nextstate, Name: state}
		commandLetStream, err := s.client.RegisterCommandLet(context.Background(), &request)
		if err != nil {
			log.Println(err)
		} else {
			s.monitorForStreamCommandLetResponse(commandLetStream, state)
		}
		time.Sleep(30 * time.Second)
	}
}*/

func (s *service) registerRemoteStream2() {
	for {
		request := remoteTelegramCommands.Request{Nextstate: "SEARCH_CHEF_NODE", State: "SEARCH_CHEF_ENVIRONMENT"}
		stream2, err := s.client.RegisterCommandLet(context.Background(), &request)
		if err != nil {
			log.Println(err)
		} else {
			s.monitorForStreamResponse2(stream2)
		}
		time.Sleep(30 * time.Second)
	}
}
func (s *service) registerRemoteStream3() {
	for {
		request := remoteTelegramCommands.Request{Nextstate: "EXECUTE", State: "SEARCH_CHEF_NODE"}
		stream3, err := s.client.RegisterCommandLet(context.Background(), &request)
		if err != nil {
			log.Println(err)
		} else {
			s.monitorForStreamResponse3(stream3)
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
		s.alertService.SendAlertKeyboard(context.TODO(), "Please select the application")

	}
}

func (s *service) monitorForStreamResponse1(client remoteTelegramCommands.RemoteCommand_RegisterCommandLetClient) {
	for {
		in, err := client.Recv()
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(in.Message)
		e, err := s.chefStore.GetChefEnvironments()
		if err != nil {
			s.alertService.SendError(context.TODO(), err)

			return
		}
		buttons := make([]string, len(e))
		for x, i := range e {
			buttons[x] = i.FriendlyName

			//s.alertService.SendAlert(context.TODO(), i.FriendlyName)
		}

		s.alertService.SendAlertEnvironment(context.TODO(),buttons)

	}
}
func (s *service) monitorForStreamResponse2(client remoteTelegramCommands.RemoteCommand_RegisterCommandLetClient) {
	for {
		in, err := client.Recv()
		log.Println(in.Message)
		if err != nil {
			log.Println(err)
			return
		}
		var a [2]string
		a[0] = "dsbggena45v"
		a[1] = "dsbggena44v"
		for i := 0; i < 1; i++ {
			response, errresp := http.Get("http://" + a[i] + ".standardbank.co.za:8080/rest/load/branch")
			if errresp != nil {
				s.alertService.SendError(context.TODO(), errresp)
			}
			responseData, err := ioutil.ReadAll(response.Body)
			if err != nil {
				s.alertService.SendError(context.TODO(), err)
			}
			s.alertService.SendAlert(context.TODO(), string(responseData)+" from "+a[i])
			continue
		}
		/*nodes := s.chefService.FindNodesFromFriendlyNames(in.Fields[0], in.Message)
		res := make([]string, len(nodes))
		for i, x := range nodes {
			res[i] = x.Name
			response, errresp := http.Get("http://" + x.Name + ":8080/branch")
			if errresp != nil {
				s.alertService.SendError(context.TODO(), errresp)
			}
			responseData, err := ioutil.ReadAll(response.Body)
			if err != nil {
				s.alertService.SendError(context.TODO(), err)
			}
			s.alertService.SendAlert(context.TODO(), string(responseData)+" from "+x.Name)
		}
*/
		//s.alertService.SendAlertNodes(context.TODO(), res)
	}
}
func (s *service) monitorForStreamResponse3(client remoteTelegramCommands.RemoteCommand_RegisterCommandLetClient) {
	for {
		in, err := client.Recv()
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(in.Message)

		log.Println("execute reload cache")

	}
}

/*
func (s *service) monitorForStreamCommandLetResponse(client remoteTelegramCommands.RemoteCommand_RegisterCommandLetClient, state string) {
	for {
		in, err := client.Recv()
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(in.From)
		if in.From == "SEARCH_CHEF" {
			s.alertService.SendAlertKeyboard(context.TODO(), "Please select the environment")
		}
		if in.From == "SEARCH_CHEF_ENVIRONMENT" {
			s.alertService.SendAlertKeyboard(context.TODO(), "Please select the node")
		}
		if in.From == "SEARCH_CHEF_NODE" {
			log.Println("execute reload cache")
		}


	}
}
*/
