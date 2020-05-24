package main

import (
	"encoding/json"
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
)
func Login() (ApiResponse,error)  {
	var ret ApiResponse
	client := http.DefaultClient
	url := fmt.Sprintf("%s/api/auth/login",LeanoteUrl)
	req , err := http.NewRequest("GET",url,nil)
	query := req.URL.Query()
	query.Add("email",UserName)
	query.Add("pwd",Password)
	req.URL.RawQuery = query.Encode()
	if err != nil{
		return ApiResponse{},err
	}
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	if body,ok:=ioutil.ReadAll(resp.Body);ok ==nil{
		if err := json.Unmarshal(body,&ret);err != nil{
			panic(err)
		}
	}
	return ret,nil
}

func Logout(token string) bool  {
	var ret ApiResponse
	client := http.DefaultClient
	url := fmt.Sprintf("%s/api/auth/logout",LeanoteUrl)
	req , err := http.NewRequest("GET",url,nil)
	if err != nil{
		return false
	}
	query := req.URL.Query()
	query.Add("token",token)
	req.URL.RawQuery = query.Encode()
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	if body,ok:=ioutil.ReadAll(resp.Body);ok ==nil{
		if err := json.Unmarshal(body,&ret);err != nil{
			panic(err)
		}
	}
	return ret.Ok
}

func AuthGetRequest(url,userid,token string) *http.Request {

	req,ReqErr := http.NewRequest("GET",url,nil)
	if ReqErr != nil{
		panic(ReqErr)
	}
	query := req.URL.Query()
	query.Add("userId",userid)
	query.Add("token",token)
	req.URL.RawQuery = query.Encode()
	return req
}

func GetAllNoteBook(userid,token string) []Notebook {
	client := http.DefaultClient
	var (
		errResp ApiResponse
		noteBooks []Notebook
	)
	api := fmt.Sprintf("%s/api/notebook/getNotebooks",LeanoteUrl)
	req := AuthGetRequest(api,userid,token)

	resp,err:=client.Do(req)
	if err != nil{
		panic(err)
	}
	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil{
		panic(readErr)
	}
	if err:=json.Unmarshal(body,&noteBooks);err != nil{
		json.Unmarshal(body,&errResp)
	}
	return noteBooks
}

func ()  {

}

func BookList(books []Notebook) {
	var root []Book
	// root book
	for _,v := range books{
		flag := func(b Book,l []Book) bool {
			for _,v := range l{
				if b.ParentId == v.BookId
			}
		}
		switch {
		case v.ParentNotebookId == "":
		case
		}

		if v.ParentNotebookId == ""{
			root = append(root,Book{
				BookId:   v.NotebookId,
				Title:    v.Title,
				ParentId: v.ParentNotebookId,
				next:     nil,
			})
		}
	}
	// middle book
	for
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
	books := GetAllNoteBook(UserInfo.UserId,UserInfo.Token)
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