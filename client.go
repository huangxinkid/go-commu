package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIP   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int // 当前client的模式
}

func NewClient(serverIP string, serverPort int) *Client {
	// 创建客户端对象
	client := &Client{
		ServerIP:   serverIP,
		ServerPort: serverPort,
		flag:       999,
	}
	// 连接 server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIP, serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}
	client.conn = conn
	// 返回对象
	return client

}

func (client *Client) DealResponse() {
	// 一旦 client.conn 有数据，就直接 copy 到标准输出，永久阻塞监听
	io.Copy(os.Stdout, client.conn)
}

func (client *Client) menu() bool {
	var flag int
	fmt.Println("1,公聊模式")
	fmt.Println("2,私聊模式")
	fmt.Println("3,更新用户名")
	fmt.Println("0,退出")
	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>>请输入合法范围内的数字<<<<")
		return false
	}
}

func (client *Client) UpdateName() bool {
	fmt.Println("请输入用户名：")
	fmt.Scanln(&client.Name)
	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}

func (client *Client) PublicChat() {
	var chatMsg string
	fmt.Println(">>>>请输入聊天内容，exit退出")
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn write err:", err)
				break
			}
		}
		chatMsg = ""
		fmt.Println(">>>>请输入聊天内容，exit退出")
		fmt.Scanln(&chatMsg)
	}
}

func (client *Client) SearchUsers() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write err:", err)
		return
	}
}

func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	client.SearchUsers()
	fmt.Println(">>>>请输入聊天对象[用户名]，exit退出")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>>请输入消息内容，exit退出")
		fmt.Scanln(&chatMsg)
		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn write err:", err)
					break
				}
			}
			chatMsg = ""
			fmt.Println(">>>>请输入消息内容，exit退出")
			fmt.Scanln(&chatMsg)
		}
		client.SearchUsers()
		fmt.Println(">>>>请输入聊天对象[用户名]，exit退出")
		fmt.Scanln(&remoteName)
	}
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {
		}
		switch client.flag {
		case 1:
			fmt.Println("公聊模式选择")
			client.PublicChat()
			break
		case 2:
			fmt.Println("私聊模式选择")
			client.PrivateChat()
			break
		case 3:
			fmt.Println("修改用户名")
			client.UpdateName()
			break
		}
	}
}

var serverIP string
var serverPort int

func init() {
	flag.StringVar(&serverIP, "ip", "127.0.0.1", "设置服务器IP地址（默认是127.0.0.1）")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口（默认是8888）")
}

func main() {
	//命令行解析
	flag.Parse()

	client := NewClient(serverIP, serverPort)
	if client == nil {
		fmt.Println(">>>>连接服务器失败...")
		return
	}
	// 单独开启一个 goroutine 去处理 server 的回执消息（接收消息）
	go client.DealResponse()

	// 启动客户端的业务（发送消息）
	fmt.Println(">>>>连接服务器成功...")
	client.Run()
}
