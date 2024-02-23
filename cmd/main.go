package main

import (
	"fmt"
	"zhipu-agent/pkg/agent"
)

func main() {
	zhipuReq := agent.NewZhipuReq("",
		"https://open.bigmodel.cn/api/paas/v4/chat/completions")
	requestBody := agent.NewRequestBody("glm-4")
	requestBody.Messages = []agent.Message{
		{Role: "system", Content: "你是一个精通go语言开发的程序员"},
		{Role: "user", Content: "帮我写一个基于gin框架的web server服务"},
	}
	fmt.Println(zhipuReq)
	zhipuReq.Request(requestBody, true)
}
