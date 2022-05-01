# README

## 代码结构

```
.
├── build                  # 构建目录
│   ├── Makefile               # Makefile
│   ├── client                 # client程序
│   ├── config                 # 静态配置文件, 如脏词列表
│   └── server                 # server程序
├── cmd                    # 可执行程序入口
│   ├── client             
│   ├── config                 # 公共配置
│   └── server            
├── doc                    # 文档
├── go.mod                 
└── pkg                    # 逻辑依赖包
    ├── depends                # 业务无关依赖/公共库/算法/协议定义
    ├── errors                 # 错误类型
    ├── models                 # 业务数据定义
    ├── modules                # 业务逻辑实现
    └── tasks                  # 任务调度

```

## 完成功能

1. 用户注册

命令: `/reg [username]`

2. 用户登陆

命令: `/login [username]`

3. 房间列表

命令: `/rooms `

4. 进入或切换房间

命令: `/room [room_id]`

5. 热词统计

命令 `/popular [room_id]`

6. 脏词替换

## 主要功能模块

1. `pkg/modules/rooms` 房间管理
2. `pkg/modules/users` 用户管理
3. `pkg/modules/profanity_words` 脏词替换
4. `pkg/modules/frequence_stat` 词频统计

## 依赖

[qlib](https://github.com/saitofun/qlib)
作者早前实现的一部分基础功能lib, chat项目主要用到qsock封装库和一些线程安全的数据结构等一些杂项.

## 关键算法简单说明

1. 脏词替换

trie字典树
算法代码: `pkg/depends/alg/trie`
功能代码: `pkg/modules/profanity_words`

接口:

```golang
func MaskWordsBy(sentence string, replacer rune) string
func AddWords(word ... string)
func LoadDictFromFile(path string)
func LoadDictByWords(words... string)
```




2. 热词统计

方案是用户输入后动态记录单词出现的时间和单词当前出现的次数.

维护一个KV记录单词和出现次数的关联关系:set
维护一个时间序列表:sequence
维护一个有序次数列表:ordered

接口

```golang
type OrderedSet struct {}

func (OrderedSet)AddWords(words... string)
func (OrderedSet)TopN(n int) []KeyCountElement
```

## 部署运行

```shell
$ cd build
$ make # 构建
$ ./server # 运行服务
$ ./client # 运行客户端
```

> Author: birdyfj@gmail.com
