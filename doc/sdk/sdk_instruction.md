# SDK使用说明
本项目提供了go、php、java三种语言版本的SDK，包括了服务端和客户端，可以用于运行server与添加job，这里以go语言为例做详细说明。
## 服务端

- 服务端用于配置消费者（即队列消息的接收者或者说调用者）

- [源代码文件](../../sdk/golang/server.go)

- 结构体函数介绍：
    - type WorkerRouter struct：储存路由信息
    - func NewWorkerRouter(pattern string) *WorkerRouter：通过后缀pattern新建WorkerRouter
    - func (wr *WorkerRouter) RegisterJobHandler(queueName, topicName string, handlerFunc WorkerHandler)：
    注册一个消费函数，queueName、topicName是队列信息，handlerFunc是用于处理请求的函数，结构为type WorkerHandler func(r *http.Request) *Resp
    - func RegisterHttpHandler(pattern string, handlerFunc http.HandlerFunc):创建一个后缀为pattern，处理函数为handlerfunc的的路由
    - func Serve(address string)：运行http server监听address，路由表由以上两个函数进行添加
- 使用示例：
    ```
    func testFunc(r *http.Request) *Resp{
	println("hello world!")
	return &Resp{
		ErrorCode:0,
		Code:200,
		Data:"hello",
	    }
    }
    func main(){
    	NewWorkerRouter("/test").RegisterJobHandler("testQueue","testTopic",testFunc)
    	Serve("192.168.238.141:4443")
    }
    ```
    该代码运行后可作为queue为testQueue、topic为testTopic的队列的消费者（需要在后台管理页面配置消费类型为http，消费文件为http://192.168.238.141:4443/test ），当该队列接收到job时，就会执行testFunc，输出hello world!（需要传输的数据可从r *http.Request中获取）

> pepperbus向消费者发送的http请求中会附带type字段，该字段的值为"ping"时，表示测试连通性，需要消费者返回header中Code字段为200，body为用于消费的队列(形式:queue1/topic1,queue2/topic2)；type字段为data时表示消费任务，用户如果不使用SDK而是自己来写的话需要注意这一点

## 客户端

- 客户端作为生产者（即队列消息的产生方），用于添加任务

- [源代码文件](../../sdk/golang/client.go)
    
- 函数介绍：
    - func NewClient(addr string) *Client：通过指定地址addr新建client，addr形式为ip:port:password，ip即是pepperbus服务器的ip，port默认为12018
    - func (c *Client) AddJobs(queue string,password string, contents ...interface{})：发送job，queue下的所有topic队列都会收到该job，password是添加Queue时设置的密码，contens是要发送的数据
- 使用示例：
    ```
    func main(){
	re := NewClient("192.168.238.141:12018")
	re.AddJobs("testQueue","test",[]string("test data1","test data2"))
    }
    ```
    改代码运行后，testQueue下的所有topic队列都会收到该job，并自动进行消费
