package chatgpt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type ChatGPT struct {
	client  openai.Client
	ctx     context.Context
	Prompt  string
	Message []openai.ChatCompletionMessage
}

type chatOption func(*openai.ChatCompletionRequest)

func newChatCompletionRequest(msg string, opt ...chatOption) *openai.ChatCompletionRequest {
	chatreq := new(openai.ChatCompletionRequest)
	chatreq.Model = openai.GPT3Dot5Turbo
	for _, o := range opt {
		o(chatreq)
	}

	return chatreq
}

func SetChatMessage(msg openai.ChatCompletionMessage) chatOption {
	return func(c *openai.ChatCompletionRequest) {
		c.Messages = append(c.Messages, msg)
	}
}

func SetUser(user string) chatOption {
	return func(c *openai.ChatCompletionRequest) {
		c.User = user
	}
}

type openaiOption func(*openai.ClientConfig)

func WithHttpClient(httpc *http.Client) openaiOption {
	return func(c *openai.ClientConfig) {
		c.HTTPClient = httpc
	}
}

func New(ctx context.Context, authToken string, opt ...openaiOption) *ChatGPT {
	opConfig := openai.DefaultConfig(authToken)
	for _, o := range opt {
		o(&opConfig)
	}
	return &ChatGPT{
		client: *openai.NewClientWithConfig(opConfig),
		ctx:    ctx,
	}
}

func (c *ChatGPT) Close() {
	c.ctx.Done()
}

type chatOpt func(chatreq *openai.ChatCompletionRequest)

func WithUser(user string) chatOpt {
	return func(chatreq *openai.ChatCompletionRequest) {
		chatreq.User = user
	}
}

func (c *ChatGPT) Chat(msg string, opt ...chatOpt) (answer string, err error) {
	chatreq := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: msg,
			},
		},
	}
	for _, o := range opt {
		o(&chatreq)
	}
	rsp, err := c.client.CreateChatCompletion(c.ctx, chatreq)
	if err != nil {
		if strings.Contains(err.Error(), "You exceeded your current quota") {
			log.Printf("当前apiKey[]配额已用完, 将删除并切换到下一个")
			// db.Orm.Table("apikey").Where("key = ?", apiKeys[0].Key).Delete(&ApiKey{})
			return "", errors.New("OpenAi配额已用完，请联系管理员")

		}
		if strings.Contains(err.Error(), "The server had an error while processing your request") {
			return "", errors.New("OpenAi服务出现问题，请重试")

		}
		if strings.Contains(err.Error(), "Client.Timeout exceeded while awaiting headers") {
			return "", errors.New("OpenAi服务请求超时，请重试")

		}
		if strings.Contains(err.Error(), "Please reduce your prompt") {
			return "", errors.New("OpenAi免费上下文长度限制为4096个词组，您的上下文长度已超出限制，请发送\"清空会话\"以清空上下文")
		}
		if strings.Contains(err.Error(), "Incorrect API key") {
			return "", errors.New("OpenAi ApiKey错误，请联系管理员")
		}
		return "", err
	}
	b := bytes.NewBuffer(nil)
	json.NewEncoder(b).Encode(rsp)
	log.Println(b)
	return rsp.Choices[0].Message.Content, nil
}
