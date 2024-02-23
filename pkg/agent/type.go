package agent

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type ZhipuClaims struct {
	APIKey    string `json:"api_key"`
	Timestamp int64  `json:"timestamp"`
	jwt.RegisteredClaims
}

type RequestBody struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type Message struct {
	Role       string `json:"role"`
	Content    string `json:"content"`
	ToolCallID string `json:"tool_call_id"`
}

func NewRequestBody(model string) *RequestBody {
	body := &RequestBody{
		Model: model,
	}
	return body
}

type ZhipuReq struct {
	RequestToken string
	URL          string
	Client       *http.Client
}

type ResponseBody struct {
	Created   int64      `json:"created"`
	ID        string     `json:"id"`
	Model     string     `json:"model"`
	RequestID string     `json:"request_id"`
	Choices   []Choice   `json:"choices"`
	Usage     UsageStats `json:"usage"`
}

type Choice struct {
	FinishReason string   `json:"finish_reason"`
	Index        int      `json:"index"`
	Message      Messages `json:"message"`
}

type Messages struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type UsageStats struct {
	CompletionTokens int `json:"completion_tokens"`
	PromptTokens     int `json:"prompt_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func NewZhipuReq(APIKey string, URL string) *ZhipuReq {
	parts := strings.Split(APIKey, ".")
	if len(parts) != 2 {
		log.Fatal("invalid apikey")
	}
	id, secret := parts[0], parts[1]
	claims := ZhipuClaims{
		APIKey:    id,
		Timestamp: time.Now().UnixMilli(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(600) * time.Second)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["alg"] = "HS256"
	token.Header["sign_type"] = "SIGN"
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		log.Fatal(err)
	}
	req := &ZhipuReq{
		RequestToken: tokenString,
		URL:          URL,
		Client:       &http.Client{},
	}
	return req
}

func (z *ZhipuReq) Request(Body *RequestBody, stream bool) {
	Body.Stream = stream
	reqBody, err := json.Marshal(Body)
	if err != nil {
		log.Fatal("Error Marshal Body:", err)
	}
	req, err := http.NewRequest("POST", z.URL, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Fatal("Error creating request:", err)
	}
	req.Header.Set("Authorization", z.RequestToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := z.Client.Do(req)
	fmt.Println(resp.Body)
	if err != nil {
		log.Fatal("Error sending request:", err)
	}
	defer resp.Body.Close()
	if stream {
		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Fatal("Error reading response body:", err)
			}
			if strings.TrimSpace(line) == "data: [DONE]" {
				break
			}
			if strings.HasPrefix(line, "data: ") {
				line = strings.TrimPrefix(line, "data: ")

				var streamData struct {
					Choices []struct {
						Delta struct {
							Content string `json:"content"`
						} `json:"delta"`
					} `json:"choices"`
				}

				if err := json.Unmarshal([]byte(line), &streamData); err != nil {
					fmt.Println("Error unmarshaling data:", err)
					continue
				}
				for _, choices := range streamData.Choices {
					fmt.Print(choices.Delta.Content)
				}
			}
		}
	} else {
		var response ResponseBody
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("Error reading response body: ", err)
		}
		err = json.Unmarshal(body, &response)
		if len(response.Choices) > 0 {
			fmt.Println("Content:", response.Choices[0].Message.Content)
		} else {
			fmt.Println("No content found")
		}
	}
}
