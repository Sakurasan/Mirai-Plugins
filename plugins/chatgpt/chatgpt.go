package chatgpt

import (
	"Mirai-Plugins/pkg/config"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis"
	openai "github.com/sashabaranov/go-openai"
)

var (
	KeyNotExist = errors.New("KeyNotExist")
)

type ChatGPT struct {
	client openai.Client
	ctx    context.Context
	// Prompt  string
	// Message []openai.ChatCompletionMessage
	Redis *redis.Client
	hasdb bool
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

	_chatgpt := &ChatGPT{
		client: *openai.NewClientWithConfig(opConfig),
		ctx:    ctx,
	}
	rdb, err := NewRedis(config.PluginConfig.GetString("plugins.chatgpt.redisaddr"), config.PluginConfig.GetString("plugins.chatgpt.redispassword"))
	if err != nil {
		_chatgpt.hasdb = false
	} else {
		_chatgpt.Redis = rdb
	}
	return _chatgpt

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

func (c *ChatGPT) ChatWithMessage(msg []openai.ChatCompletionMessage, opt ...chatOpt) (answer string, err error) {
	chatreq := openai.ChatCompletionRequest{
		Model:    openai.GPT3Dot5Turbo,
		Messages: msg,
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
	if len(msg) == 10 {
		msg = msg[1:]
	}
	msg = append(msg, openai.ChatCompletionMessage{Role: rsp.Choices[0].Message.Role, Content: rsp.Choices[0].Message.Content})
	if err := c.SetMessage(chatreq.User, msg); err != nil {
		log.Println(err)
	}

	return rsp.Choices[0].Message.Content, nil
}

func (c *ChatGPT) SetMessage(user string, msg []openai.ChatCompletionMessage) error {
	var buf = bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}
	if _, err := c.Redis.Set(user, buf.String(), 10*time.Minute).Result(); err != nil {
		return err
	}
	return nil
}

func (c *ChatGPT) GetMessage(user string) ([]openai.ChatCompletionMessage, error) {
	result, err := c.Redis.Get(user).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		log.Println(err)
		return nil, err
	}

	var str = strings.NewReader(result)
	var chatmsg []openai.ChatCompletionMessage
	if err := json.NewDecoder(str).Decode(&chatmsg); err != nil {
		log.Println(err)
		return nil, err
	}
	return chatmsg, nil
}
