import requests
from requests_toolbelt import MultipartEncoder
from urllib.parse import urljoin
import logging
import re
import os
import sys
import getpass
import datetime
import json
import pprint

class ApiResponseError(BaseException):
    pass

class Wiz:

    AsUrl = 'https://as.wiz.cn'

    def __init__(self,username: str,password: str):
        self.logger = logging.getLogger("Wiz")
        console = logging.StreamHandler(stream=sys.stdout)
        self.logger.addHandler(console)
        self.logger.setLevel(logging.INFO)
        self.LoginResult = {}
        self.token = ""
        self.kbServer = ""
        self.kbGuid = ""
        self.username = username
        self.password = password
        self.existFolder = {}

    def __execRequest(self,method: str,url: str,body,**kwargs):
        if self.token != "":
            header = {
                'User-Agent': 'Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:50.0) Gecko/20100101 Firefox/50.0',
                "X-Wiz-Token":self.token
            }
        else:
            header = {}

        if kwargs.get("headers","") != "":
            self.logger.debug(kwargs["headers"])
            for k,v in kwargs["headers"].items():
                header[k] = v
            kwargs.pop("headers")
        # self.logger.debug(f"headers: {header},kwargs: {kwargs}")

        resp = requests.request(method=method.lower(),url=url,json=body,headers=header,**kwargs)

        try:
            ret = resp.json()
            if ret["returnCode"] != 200:
                self.logger.error(f'ERROR!! {ret["returnMessage"]}')
                raise ApiResponseError(ret["returnMessage"])

            return ret.get("result", "")
        except json.JSONDecodeError:
            ret = resp.text


        self.logger.debug(f"{url}->{ret}")
        return ret


    def Login(self):
        self.logger.info("开始登陆！！")
        url = urljoin(Wiz.AsUrl,"/as/user/login")
        body = {
            "userId": self.username,
            "password": self.password
        }
        ret = self.__execRequest("post",url,body)
        self.LoginResult = ret
        self.token = self.LoginResult.get("token","")
        self.kbServer = self.LoginResult.get("kbServer","")
        self.kbGuid = self.LoginResult.get("kbGuid","")
        self.logger.info(f"登陆成功！欢迎 {self.username}")
        return self

    def Exit(self):
        self.logger.info(f"退出 {self.username} 登陆！！")
        url = urljoin(Wiz.AsUrl,"/as/user/logout")
        ret = self.__execRequest(method="get",url=url,body={})
        self.logger.info("退出登陆成功！！")
        return ret

    def __createNote(self,title, folder, html):
        url = f"{self.kbServer}/ks/note/create/{self.kbGuid}"
        self.logger.debug(f"Create Note: {url}")
        note = {
              "kbGuid":self.kbGuid,
              "html":html,
              "title":title,
              "category": folder,
            }
        return self.__execRequest("post",url,note)

    def __updateNote(self,note: dict):
        url = f'{self.kbServer}/ks/note/save/{self.kbGuid}/{note.get("docGuid","")}'
        self.logger.debug(f"Update Note: {url}")
        return self.__execRequest("put",url,note)

    def __uploadImage(self,docGuid,imageFile):
        url = f'{self.kbServer}/ks/resource/upload/{self.kbGuid}/{docGuid}'
        self.logger.debug(f"Upload image: {url}")
        self.logger.debug(f"filepath: {imageFile}")
        multipart_encoder = MultipartEncoder(
            fields={
                "kbGuid": self.kbGuid,
                "docGuid": docGuid,
                "data":(os.path.basename(imageFile),open(imageFile,mode="rb"),"image/jpeg")
            }
        )

        headers = {
            'User-Agent': 'Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:50.0) Gecko/20100101 Firefox/50.0',
            "content-type": multipart_encoder.content_type,
            "X-Wiz-Token": self.token
        }
        # self.logger.debug(multipart_encoder.to_string())
        resp = requests.post(url,data=multipart_encoder,headers=headers)

        # self.logger.debug(resp.status_code)
        # self.logger.debug(resp.request.headers)
        self.logger.debug(resp.text)
        return resp.json() if resp is not None else resp

    # 获取当前所有目录
    def GetALlDirectory(self):
        url = f"{self.kbServer}/ks/category/all/{self.kbGuid}"
        self.logger.debug(f"Get All directory: {url}")
        return self.__execRequest("get",url,body={})

    def __CheckLine(self,line: str)-> tuple:
        if ".png" in line:
            o = re.compile(r"\!\[.*\]\(.*.png\)")
            m = o.match(line)
            if m is None:
                return None
            # self.logger.debug(m.string)
            flag = False
            s = []
            for i in m.string:
                if i == "(":
                    flag = True
                    continue
                if i == ")":
                    flag = False
                    s.append("<>")
                if flag:
                   s.append(i)
            return "".join(s)
        else:
            return None

    # 导入一个markdown 笔记
    def LoadMarkdownWithImage(self,folder,title,MarkdownPath: str):
        note = self.__createNote(title,folder,'<html><head></head><body></body></html>')
        self.logger.debug(f"New note {note}")
        docGuid = note["docGuid"]
        base = os.path.dirname(MarkdownPath)
        contents = []
        resources = []
        with open(MarkdownPath,mode="rt",encoding="utf8") as fd:
            for line in fd:
                t = line.replace("\n","<br>",-1)
                i = self.__CheckLine(line=t)
                if i is not None:
                    for r in i.split("<>"):
                        if r == "":
                            continue
                        imagefile = self.__uploadImage(docGuid,os.path.join(base,r))
                        resources.append(imagefile["name"])
                        t = t.replace(r,imagefile["url"])
                contents.append(t)

        # self.logger.debug(contents)
        if len(contents) >0:
            note["html"] = "".join(contents)
            self.logger.debug(note.get("html"))
            note["resources"] = resources
            self.logger.debug(f"Update note {note}")
            self.__updateNote(note)
        return note

    def __CreateFolder(self,parent,child):
        if self.existFolder[child] > 1:
            return
        url = f"{self.kbServer}/ks/category/create/{self.kbGuid}"
        t = datetime.datetime.now()
        body = dict(parent=parent,child=child,pos=t.timestamp())
        self.existFolder[child] += 1
        self.logger.debug(f"{parent}->{child}")
        return self.__execRequest("post",url,body)

    def __makedirs(self,path:str):
        p,*c = path.split("/")
        for i in c:
            if i not in self.existFolder.keys():
                self.existFolder[i] = 0
        if len(c) == 1:
            pt = f"/{p}/"
            child = c.pop()
            self.__CreateFolder(pt,child)
        else:
            for i,_ in enumerate(c):
                m = "/".join(c[:i])
                pt = f'/{p}/'  if m == ""  else f'/{p}/{m}/'
                child = c[i]
                self.__CreateFolder(pt,child)
        return

    # 导入一个目录树
    def UpLoadDirectory(self,root: str):

        if os.path.isabs(root):
            parent = os.path.dirname(root)
            os.chdir(parent)
            p = os.path.basename(root)
        else:
            p = root
        dirs = []
        files = []
        for r,d,f in os.walk(p):
            if "images" in r or "attach" in r:
                continue
            if len(d) == 0:
                dirs.append(r)

            for v in f:
                if not v.startswith("."):
                    files.append(os.path.join(r,v))

        # self.logger.debug(f"file -> {files}")
        # self.logger.debug(f"dirs -> {dirs}")
        for i in dirs:
            self.__makedirs(i)

        for f in files:
            self.logger.debug(f)
            folder  = "/" + os.path.dirname(f) + "/"
            filename = os.path.basename(f)
            self.LoadMarkdownWithImage(folder=folder,title=filename,MarkdownPath=f)
            
    # 删除一个目录
    def DeleteDirs(self,folder):
        url  = f"{self.kbServer}/ks/category/delete/{self.kbGuid}{folder}"
        return self.__execRequest("delete",url,body={})

    def __enter__(self):
        return self.Login()

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.Exit()
        return




if __name__ == '__main__':
    username = input("username: ")
    password = getpass.getpass("password: ")
    # folder = "/运维工作/故障处理记录/"
    # title = "mail日志记录.md"
    # mdPath = "/Users/charles/Downloads/ouput" + folder + title
    # username = "xxxx"
    # password = "xxxx"
    root = "/Users/charles/Downloads/ouput/杂项"
    with Wiz(username,password) as w:
        w.logger.setLevel(logging.DEBUG)

        # w.DeleteDirs("运维工作/")

        # root = input("root: ")
        w.UpLoadDirectory(root)
        # ret = w.GetALlDirectory()
        # pprint.pprint(ret)
        # print(mdPath)
        # note = w.LoadMarkdownWithImage(folder=folder,title=title,MarkdownPath=mdPath)