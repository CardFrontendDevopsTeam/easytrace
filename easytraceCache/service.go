package easytraceCache

import (
	"github.com/weAutomateEverything/go2hal/remoteTelegramCommands"
	"time"
	"golang.org/x/net/context"
	"log"
	"os"
	"encoding/json"
	"net/http"
	"bytes"
	"io/ioutil"
	"gopkg.in/mgo.v2/bson"
)

type Service interface {
}

type service struct {
	client remoteTelegramCommands.RemoteCommandClient
}
type Recipe struct {
	ID           bson.ObjectId `bson:"_id,omitempty"`
	Recipe       string
	FriendlyName string
	ChatID       uint32
}

func NewService(client remoteTelegramCommands.RemoteCommandClient) Service {
	s := &service{client}
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
		request := remoteTelegramCommands.RemoteCommandRequest{Description: "Clear EasytraceCache", Name: "ReloadCache",Group:418124524}
		regcommandstream, err := s.client.RegisterCommand(context.Background(), &request)
		if err != nil {
			log.Println(err)
		} else {
			s.monitorForStreamResponse(regcommandstream)
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
		err = s.SetState(in.From, in.Chat, "SEARCH_CHEF")
		if err != nil {
			log.Println("The HTTP request failed with error %s\n", err)
		} else {
			res,err :=s.GetRecipes("418124524")
			if err!=nil{
				log.Println("The HTTP request failed with error %s\n", err)
			}
			chatid, _ := s.GetRoom("418124524")
			s.SendKeyboard(res,"Please select the application",chatid)
		}
	}
}

func (s *service) monitorRecipeReply(client remoteTelegramCommands.RemoteCommand_RegisterCommandLetClient) {
	for {
		_, err := client.Recv()
		if err != nil {
			log.Println(err)
			return
		}
		res,err :=s.GetEnvironments("418124524")
		if err != nil {
			log.Println("The HTTP request failed with error %s\n", err)
		}

		chatid, _ := s.GetRoom("418124524")
		s.SendKeyboard(res,"Please select the environment",chatid)

		}
}
func (s *service) monitorEnvironmentReply(client remoteTelegramCommands.RemoteCommand_RegisterCommandLetClient) {
	for {
		in, err := client.Recv()
		if err != nil {
			log.Println(err)
			return
		}

		chatid, _ := s.GetRoom("418124524")
		res,err :=s.GetNodes(in.Fields[0],in.Message,418124524)
		if err!=nil{
			log.Println("The HTTP request failed with error %s\n", err)
		}
		s.SendKeyboard(res,"Please select the node",chatid)
	}
}

func (s *service) monitorNodeReply(client remoteTelegramCommands.RemoteCommand_RegisterCommandLetClient) {
	for {
		in, err := client.Recv()
		if err != nil {
			log.Println(err)
			return
		}
		response, errresp := http.Get(getProtocol() + "://" + in.Message + "." + getDomain() + ":" + getPort() + getRestPath())
		if errresp != nil {
			log.Println(errresp)
		}
		responseData, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Println("The HTTP request failed with error %s\n", err)
		}
		log.Println(string(responseData) + " from " + in.Message)

	}
}
func (s *service) SetState(user string, chat string, state string) error {
	stateReq := setStateRequest{User: user, Chat: chat, State: "SEARCH_CHEF"}
	jsonValue, _ := json.Marshal(stateReq)
	request, _ := http.NewRequest("POST", "http://localhost:8000/api/telegram/state", bytes.NewBuffer(jsonValue))
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	_, err := client.Do(request)
	if err != nil {
		return err
	}
	return nil
}
func (s *service) GetRecipes(chat string) ([]string,error) {
	response, err := http.Get("http://localhost:8000/api/chef/recipes/group/" + chat)
	if err != nil {
		return nil,err
	}
	responseData, err := ioutil.ReadAll(response.Body)
	res, err := getRecipes(responseData)
	if err!=nil {
		return nil,err
	}
	return res,nil
}
func (s *service) GetRoom(chat string) (int64,error) {
	response, err := http.Get("http://localhost:8000/api/telegram/room/" + chat)
	if err != nil {
		return 0,err
	}
	responseData, err := ioutil.ReadAll(response.Body)
	log.Println(string(responseData))
	res,err:=getRoom(responseData)
	if err!=nil {
		return 0,err
	}
	return res,nil
}
func (s *service) GetEnvironments(chat string) ([]string,error) {
	response, err := http.Get("http://localhost:8000/api/chef/environments/group/" + chat)
	if err != nil {
		return nil,err
	}
	responseData, err := ioutil.ReadAll(response.Body)
	res, err := getEnvironments(responseData)
	if err!=nil {
		return nil,err
	}
	return res,nil
}
func (s *service) GetNodes(recipe string,environment string,chat uint64) ([]string,error) {
	jsonData2 := chefNodeRequest{Recipe: recipe, Environment: environment, Chat: uint32(chat)}
	jsonValue2, _ := json.Marshal(jsonData2)
	request2, _ := http.NewRequest("POST", "http://localhost:8000/api/chef/nodes", bytes.NewBuffer(jsonValue2))
	request2.Header.Set("Content-Type", "application/json")
	client2 := &http.Client{}
	response2, err := client2.Do(request2)
	if err!=nil{
		return nil,err
	}
	bodyText, err := ioutil.ReadAll(response2.Body)
	res, err := getNodes(bodyText)
	return res,nil
}
func (s *service) SendKeyboard(options []string,message string,groupid int64) error {
	jsonData1 := sendKeyBoardRequest{Options: options, Message: "Please select the application", GroupId: groupid}
	jsonValue1, _ := json.Marshal(jsonData1)
	request1, _ := http.NewRequest("POST", "http://localhost:8000/api/telegram/keyboard", bytes.NewBuffer(jsonValue1))
	request1.Header.Set("Content-Type", "application/json")
	client1 := &http.Client{}
	_, err := client1.Do(request1)
	return err
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

type recipeResponse struct {
	Recipes []string
}
type environmentResponse struct {
	Environments []string
}
type nodeResponse struct {
	Nodes []string
}
type sendKeyBoardRequest struct {
	Options []string
	Message string
	GroupId int64
}
type chefNodeRequest struct {
	Recipe      string
	Environment string
	Chat        uint32
}
type setStateRequest struct {
	User  string
	Chat  string
	State string
}

func getRecipes(body []byte) ([]string, error) {
	var s = new(recipeResponse)
	err := json.Unmarshal(body, &s)
	if err != nil {
		log.Println(err)
	}

	return s.Recipes, err
}
func getEnvironments(body []byte) ([]string, error) {
	var s = new(environmentResponse)
	err := json.Unmarshal(body, &s)
	if err != nil {
		log.Println(err)
	}

	return s.Environments, err
}
func getNodes(body []byte) ([]string, error) {
	var s = new(nodeResponse)
	err := json.Unmarshal(body, &s)
	if err != nil {
		log.Println(err)
	}

	return s.Nodes, err
}
func getRoom(body []byte) (int64, error) {
	var s = new(roomResponse)
	err := json.Unmarshal(body, &s)
	if err != nil {
		log.Println(err)
	}
    log.Println(string(body))
	return s.Id, err
}
type roomResponse struct {
	Id int64
}