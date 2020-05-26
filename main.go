package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mewbak/gopass"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

var (
	UserName   string
	Password   string
	LeanoteUrl = "https://leanote.com"
	UserId     string
	Token      string
	DirRoot    = "."
)

const UserAgent = `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36`

type querystring struct {
	key   string
	value string
}

func httpClientDo(req *http.Request) ([]byte, error) {
	//fmt.Printf("%s\n",req.URL.String())
	client := http.DefaultClient
	client.Timeout = time.Second * 10
	resp, ReqErr := client.Do(req)
	if ReqErr != nil {
		return nil, ReqErr
	}
	if resp.StatusCode != 200 {
		log.Println(resp.StatusCode)
		return nil, errors.New("Respones is not 200")
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func Login() (ApiResponse, error) {
	var ret ApiResponse
	url := fmt.Sprintf("%s/api/auth/login", LeanoteUrl)
	req, err := http.NewRequest("GET", url, nil)
	query := req.URL.Query()
	query.Add("email", UserName)
	query.Add("pwd", Password)
	req.URL.RawQuery = query.Encode()
	req.Header.Set("User-Agent", UserAgent)
	if err != nil {
		return ApiResponse{}, err
	}

	if body, ok := httpClientDo(req); ok == nil {
		//fmt.Println(string(body))
		if err := json.Unmarshal(body, &ret); err != nil {
			panic(err)
		}
		return ret, nil
	} else {
		return ApiResponse{}, ok
	}
}

func Logout(user *ApiResponse) bool {
	var ret ApiResponse
	url := fmt.Sprintf("%s/api/auth/logout", LeanoteUrl)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}
	query := req.URL.Query()
	query.Add("token", user.Token)
	req.URL.RawQuery = query.Encode()
	req.Header.Set("User-Agent", UserAgent)
	if body, ok := httpClientDo(req); ok == nil {
		if err := json.Unmarshal(body, &ret); err != nil {
			panic(err)
		}
	}
	if ret.Ok {
		fmt.Printf("bye bye! %s\n", user.UserName)
		return true
	}
	return false
}

func AuthGetRequest(url string, q ...querystring) *http.Request {
	req, ReqErr := http.NewRequest("GET", url, nil)
	if ReqErr != nil {
		panic(ReqErr)
	}
	query := req.URL.Query()
	query.Add("userId", UserId)
	query.Add("token", Token)
	req.Header.Set("User-Agent", UserAgent)
	for _, v := range q {
		query.Add(v.key, v.value)
	}
	req.URL.RawQuery = query.Encode()
	return req
}

func GetAllNoteBook() []Notebook {
	var (
		errResp   ApiResponse
		noteBooks []Notebook
	)
	api := fmt.Sprintf("%s/api/notebook/getNotebooks", LeanoteUrl)
	req := AuthGetRequest(api)
	body, readErr := httpClientDo(req)
	if readErr != nil {
		panic(readErr)
	}
	if err := json.Unmarshal(body, &noteBooks); err != nil {
		json.Unmarshal(body, &errResp)
	}
	return noteBooks
}

func hasNote(notebookId string) ([]Note, bool) {
	var (
		noteList []Note
		errResp  ApiResponse
	)
	api := fmt.Sprintf("%s/api/note/getNotes", LeanoteUrl)
	req := AuthGetRequest(api, querystring{
		key:   "notebookId",
		value: notebookId,
	})

	body, GetErr := httpClientDo(req)
	if GetErr != nil {
		panic(GetErr)
	}
	err := json.Unmarshal(body, &noteList)
	if err != nil {
		json.Unmarshal(body, &errResp)
		log.Println(errResp.Msg)
		return nil, false
	}
	return noteList, true
}

func GetNoteContent(noteId string) (Note, error) {
	var (
		n       Note
		errResp ApiResponse
	)

	api := fmt.Sprintf("%s/api/note/getNoteAndContent", LeanoteUrl)
	req := AuthGetRequest(api, querystring{
		key:   "noteId",
		value: noteId,
	})
	body, GetErr := httpClientDo(req)
	if GetErr != nil {
		panic(GetErr)
	}
	if err := json.Unmarshal(body, &n); err != nil {
		json.Unmarshal(body, errResp)
		log.Fatalln(errResp.Msg)
	}
	if n.IsTrash {
		return Note{}, errors.New("delete!!")
	}
	return n, nil
}

func GetImage(fileId string) ([]byte, error) {
	api := fmt.Sprintf("%s/api/file/getImage", LeanoteUrl)
	req := AuthGetRequest(api, querystring{
		key:   "fileId",
		value: fileId,
	})
	return httpClientDo(req)
}

func GetAttach(fileId string) ([]byte, error) {
	api := fmt.Sprintf("%s/api/file/getAttach", LeanoteUrl)
	req := AuthGetRequest(api, querystring{
		key:   "fileId",
		value: fileId,
	})
	return httpClientDo(req)
}

func MakeDirTrees(books []Notebook) []*Book {
	var (
		roots  []*Book
		childs []*Book
	)
	// filter list
	for _, value := range books {
		if value.ParentNotebookId == "" {
			roots = append(roots, &Book{
				Id:       value.NotebookId,
				Title:    value.Title,
				ParentId: "",
			})
		} else {
			childs = append(childs, &Book{
				Id:       value.NotebookId,
				Title:    value.Title,
				ParentId: value.ParentNotebookId,
			})
		}
	}

	for _, v := range roots {
		MakeTree(childs, v)
	}
	return roots
}

func Write(b []byte, p string) error {
	f, e := os.OpenFile(p, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0755)
	if e != nil {
		return e
	}
	_, wErr := f.Write(b)
	return wErr
}

func Writefile(p string, n Note) error {
	var (
		abs        string
		content    string
		attachPath string
	)
	if n.IsMarkdown {
		abs = path.Join(p, n.Title+".md")
	} else {
		abs = path.Join(p, n.Title+".txt")
	}
	content = n.Content
	if len(n.Files) > 0 {
		for _, v := range n.Files {
			if v.IsAttach {
				b, e := GetAttach(v.FileId)
				if e != nil {
					return e
				}
				attachdir:= path.Join(p,"attach")
				os.MkdirAll(attachdir,0755)
				filename := v.Title
				if filename == "" {
					filename = fmt.Sprintf("%s.%s",v.FileId,v.Type)
				}
				attachPath = path.Join(attachdir, filename)
				Write(b, attachPath)
			} else {
				b, e := GetImage(v.FileId)
				if e != nil {
					return e
				}
				imagedir := path.Join(p,"images")
				os.MkdirAll(imagedir,0755)
				filetype := v.Type
				if filetype == ""{
					filetype = ".png"
				}
				rp := path.Join(imagedir, v.FileId+filetype)
				Write(b, rp)
				old := fmt.Sprintf("%s/api/file/getImage?fileId=%s", LeanoteUrl, v.FileId)
				shortPath := path.Join("images",v.FileId+filetype)
				content = strings.ReplaceAll(content, old, shortPath)
			}
		}
	}

	return Write([]byte(content), abs)
}

func GetNoteAndAll(bookID string, dirname string) error {
	if notes, ok := hasNote(bookID); ok {
		for _, v := range notes {
			note, e := GetNoteContent(v.NoteId)
			if e != nil {
				return e

			}
			// if is delete !
			if note.IsTrash {
				continue
			}

			err := Writefile(dirname, note)

			if err != nil {
				//log.Println(err)
				continue
			}
		}
	}
	return nil
}

func MakeDirs(parent string, books *Book) {
	P := path.Join(parent, books.Title)
	os.MkdirAll(P, 0755)
	GetNoteAndAll(books.Id, P)
	if len(books.Child) > 0 {
		for _, v := range books.Child {
			MakeDirs(P, v)
		}
	}
}

func Exposes(notebooks []*Book) {
	for _, v := range notebooks {
		MakeDirs(DirRoot,v)
	}
}

func main() {
	var leanoteurl string
	fmt.Printf("Leanote 网址[default: %s]： ", LeanoteUrl)
	fmt.Scanln(&leanoteurl)
	if leanoteurl != "" {
		LeanoteUrl = leanoteurl
	}
	fmt.Printf("UserName: ")
	fmt.Scanln(&UserName)
	if passwd, ok := gopass.GetPass("password: "); ok == nil {
		Password = passwd
	}
	fmt.Println(UserName, Password)
	UserInfo, UserError := Login()
	if UserError != nil {
		panic(UserError)
	}
	//fmt.Printf("%#v\n", UserInfo)
	if !UserInfo.Ok {
		log.Fatalf("登陆失败：%s", UserInfo.Msg)
	} else {
		fmt.Printf("登陆成功 %s(%s)!\n", UserInfo.UserName, UserInfo.Email)
	}
	defer Logout(&UserInfo)
	UserId = UserInfo.UserId
	Token = UserInfo.Token
	books := GetAllNoteBook()
	dirTree := MakeDirTrees(books)
	if b, err := json.Marshal(&dirTree); err == nil {
		f, ferr := os.OpenFile("./dirtrees.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if ferr != nil {
			panic(ferr)
		}
		defer f.Close()
		f.Write(b)
	}
	fmt.Printf("保存到[default(.)]:  ")
	var p string
	fmt.Scanln(&p)
	if p != ""{
		DirRoot = p
	}
	Exposes(dirTree)
}
