package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mewbak/gopass"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	UserName string
	Password string
	LeanoteUrl = "https://leanote.com"
	UserId string
	Token string
)

type querystring struct {
	key string
	value string
} 

func Login() (ApiResponse,error)  {
	var ret ApiResponse
	url := fmt.Sprintf("%s/api/auth/login",LeanoteUrl)
	req , err := http.NewRequest("GET",url,nil)
	query := req.URL.Query()
	query.Add("email",UserName)
	query.Add("pwd",Password)
	req.URL.RawQuery = query.Encode()
	if err != nil{
		return ApiResponse{},err
	}
	if body,ok:=httpClientDo(req);ok ==nil{
		if err := json.Unmarshal(body,&ret);err != nil{
			panic(err)
		}
	}
	return ret,nil
}

func Logout(token string) bool  {
	var ret ApiResponse
	url := fmt.Sprintf("%s/api/auth/logout",LeanoteUrl)
	req , err := http.NewRequest("GET",url,nil)
	if err != nil{
		return false
	}
	query := req.URL.Query()
	query.Add("token",token)
	req.URL.RawQuery = query.Encode()
	if body,ok:=httpClientDo(req);ok ==nil{
		if err := json.Unmarshal(body,&ret);err != nil{
			panic(err)
		}
	}
	return ret.Ok
}

func AuthGetRequest(url string,q ...querystring) *http.Request {
	req,ReqErr := http.NewRequest("GET",url,nil)
	if ReqErr != nil{
		panic(ReqErr)
	}
	query := req.URL.Query()
	query.Add("userId",UserId)
	query.Add("token",Token)
	for _,v := range q{
		query.Add(v.key,v.value)
	}
	req.URL.RawQuery = query.Encode()
	return req
}

func httpClientDo(req *http.Request) ([]byte, error) {
	client := http.DefaultClient
	resp,ReqErr := client.Do(req)
	if ReqErr != nil{
		return nil,ReqErr
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func GetAllNoteBook() []Notebook {
	var (
		errResp ApiResponse
		noteBooks []Notebook
	)
	api := fmt.Sprintf("%s/api/notebook/getNotebooks",LeanoteUrl)
	req := AuthGetRequest(api)
	body, readErr := httpClientDo(req)
	if readErr != nil{
		panic(readErr)
	}
	if err:=json.Unmarshal(body,&noteBooks);err != nil{
		json.Unmarshal(body,&errResp)
	}
	return noteBooks
}

func hasNote(notebookId string) ([]Note,bool) {
	var (
		noteList []Note
		errResp ApiResponse
	)
	api := fmt.Sprintf("%s/api/note/getNotes",LeanoteUrl)
	req := AuthGetRequest(api,querystring{
		key:   "noteId",
		value: notebookId,
	})
	
	body,GetErr := httpClientDo(req)
	if GetErr != nil{
		panic(GetErr)
	}
	err:= json.Unmarshal(body,&noteList)
	if err !=nil{
		json.Unmarshal(body,&errResp)
		log.Println(errResp.Msg)
		return nil,false
	}
	return noteList,true
}

func GetNoteContent(noteId string) (Note,error) {
	var (
		n Note
		errResp ApiResponse
	)

	api := fmt.Sprintf("%s/api/note/getNoteAndContent",LeanoteUrl)
	req := AuthGetRequest(api,querystring{
		key:   "noteId",
		value: noteId,
	})
	body,GetErr :=httpClientDo(req)
	if GetErr != nil{
		panic(GetErr)
	}
	if err :=json.Unmarshal(body,&n); err != nil{
		json.Unmarshal(body,errResp)
		log.Fatalln(errResp.Msg)
	}
	if n.IsTrash {
		return Note{},errors.New("delete!!")
	}
	return n,nil
}

func GetImage(fileId string) ([]byte,error) {
	api := fmt.Sprintf("%s/api/file/getImage",LeanoteUrl)
	req := AuthGetRequest(api,querystring{
		key:   "fileId",
		value: fileId,
	})
	return httpClientDo(req)
}

func GetAttach(fileId string) ([]byte,error) {
	api := fmt.Sprintf("%s/api/file/getAttach",LeanoteUrl)
	req := AuthGetRequest(api,querystring{
		key:   "fileId",
		value: fileId,
	})
	return httpClientDo(req)
}

func main() {
	var leanoteurl string
	fmt.Printf("Leanote 网址[default: %s]： ",LeanoteUrl)
	fmt.Scanln(&leanoteurl)
	if leanoteurl != ""{
		LeanoteUrl = leanoteurl
	}
	fmt.Printf("UserName: ")
	fmt.Scanln(&UserName)
	if passwd , ok :=gopass.GetPass("password: ");ok == nil{
		Password = passwd
	}
	fmt.Println(UserName,Password)
	UserInfo,UserError := Login()
	if UserError != nil{
		panic(UserError)
	}
	if !UserInfo.Ok{
		log.Fatalf("登陆失败：%s",UserInfo.Msg)
	} else {
		fmt.Printf("登陆成功 %s(%s)!\n",UserInfo.UserName,UserInfo.Email)
	}
	UserId = UserInfo.UserId
	Token = UserInfo.Token
	books := GetAllNoteBook()
	for _,v := range books{
		//if v.ParentNotebookId == ""{
		//	t = append(t, NewSubBook(&v))
		//} else {
		//	AddChildBook(v.ParentNotebookId,t)
		//}
		fmt.Printf("title: %s -> BookId: %s -> parentId: %s -> seq: %d\n",v.Title,v.NotebookId,v.ParentNotebookId,v.Seq)
	}
 	//fmt.Printf("%v\n",books[0].ParentNotebookId)
	if Logout(UserInfo.Token){
		fmt.Printf("bye bye! %s\n",UserInfo.UserName)
	}
}