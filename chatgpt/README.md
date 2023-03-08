
## 配置文件
```
plugins:
  chatgpt:
    authToken: sk-sk-nidemiyuezsbdzsbdzsbdzsbd
    proxyUrl: http://127.0.0.1:7890
    redisaddr: 127.0.0.1:6379
    redispassword: password

```
## 开发文档

https://platform.openai.com/docs/guides/chat/introduction

https://platform.openai.com/docs/api-reference/chat/create

curl https://api.openai.com/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer sk-nidemiyuezsbdzsbdzsbdzsbd' \
  -d '{
  "model": "gpt-3.5-turbo",
  "messages": [{"role": "user", "content": "谢谢,翻译成英语"},{"role":"assistant","content":"Thanks"},{"role": "user", "content": "日语"}]
}'
