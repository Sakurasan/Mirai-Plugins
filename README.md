# Mirai-Plugins

<a title="Docker Image CI" target="_blank" href="https://github.com/Sakurasan/Mirai-Plugins/actions"><img alt="GitHub Workflow Status" src="https://img.shields.io/github/actions/workflow/status/Sakurasan/Mirai-Plugins/build.yaml?label=Actions&logo=github&style=flat-square"></a>
<a title="Docker Pulls" target="_blank" href="https://hub.docker.com/r/mirrors2/qchat"><img src="https://img.shields.io/docker/pulls/mirrors2/qchat.svg?logo=docker&label=docker&style=flat-square"></a>

一个基于MiraiGo-Template的插件列表，欢迎👏🏻PR,参考ping/pong模板
### 插件列表
- ping/pong
- alarmclock
- bilibili
- [ChatGPT](./cmd/chat/README.md)

---
@机器人 提问 (此模式支持上下文)

群聊 普通问答

docker-compose.yaml
```Docker-compose.yaml
version: '3'
services:
  redis:
    image: redis:latest
    container_name: redis
    restart: always
    ports:
      - 6379:6379
    environment:
      TZ: Asia/Shanghai      
#    volumes:
#      - ./data:/data
#      - ./conf/redis.conf:/etc/redis/redis.conf
#      - ./logs:/logs
#    command: ["redis-server","/etc/redis/redis.conf"]  

  qchat:
    image: mirrors2/qchat:latest
    container_name: qchat
    depends_on:
      - redis
    restart: unless-stopped
    volumes:
      - ./config:/app/config
      - ./logs:/app/logs
    
```
### 配置文件
authToken获取地址 https://platform.openai.com/account/api-keys

./config/plugins.yaml
```
plugins:
  chatgpt:
    authToken: sk-nidemiyuezsbdzsbdzsbdzsbd
    proxyUrl: 
    redisaddr: redis:6379
    redispassword: 
```
使用`docker-compose up -d`启动容器，然后`docker logs qchat`查看日志并扫码登录

