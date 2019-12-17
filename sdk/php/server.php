<?php
/**
 * Created by zhoukunta@qq.com.
 * User: johntech
 * Date: 02/08/2017
 * Time: 3:21 PM
 */
ini_set('max_execution_time', 0);
ini_set('default_socket_timeout', -1);
ini_set('memory_limit', "512M");
set_time_limit(0);
require_once "base.php";

class PepperBusServer
{
    const DEFAULT_TOPIC_NAME = "default";

    public $router = array();

    public $request;

    public $response;

    public function __construct()
    {
        $this->response = new Response();
        $this->setResponse(Response::CODE_OK, 0, "");
    }

    public function register($queue, $class, $method)
    {
        // TODO: 这里理论上也是系统级错误，目前这里的异常未处理
        if (empty($queue) || array_key_exists($queue, $this->router)) {
            throw new Exception("Invalid queue or dumplicated register", Response::CODE_FAIL);
        }
        $this->router[$queue] = array("class"=>$class,"method"=>$method);
    }

    public function parseRequest()
    {
        $this->request = new Request();
    }

    public function process()
    {
        switch ($this->request->type) {
        case Request::TypePing:
            $this->ping();
            break;
        case Request::TypeData:
            $this->call();
            break;
        }
    }

    private function ping()
    {
        $this->setResponse(Response::CODE_OK, 0, join(",", array_keys($this->router)));
    }

    // @todo check if request pass by reference
    private function call()
    {
        $queue = $this->request->getQueuePath();

        if (empty($queue)) {
            $this->setResponse(
                Response::CODE_FAIL,
                Response::CODE_FAIL,
                "Queue is empty"
            );
            return;
        }

        $className = !empty($this->router[$queue]["class"]) ? $this->router[$queue]["class"] : '';
        $method = !empty($this->router[$queue]["method"]) ? $this->router[$queue]["method"] : '';

        // 系统级错误判断，检查是否注册了类与方法、类与方法是否存在
        if (empty($className) || empty($method)) {
            $this->setResponse(
                Response::CODE_NOT_FOUND,
                Response::CODE_NOT_FOUND,
                "The className or method empty. Queue: " .  $queue
            );
            return;
        }

        if (!class_exists($className)) {
            $this->setResponse(
                Response::CODE_NOT_FOUND,
                Response::CODE_NOT_FOUND,
                "Can not found Classesss or Method registered to QueueName: " . $queue
            );
            return;
        }
        $instance = new $className();
        if (!method_exists($instance, $method)) {
            $this->setResponse(
                Response::CODE_NOT_FOUND,
                Response::CODE_NOT_FOUND,
                "Invalid method: " . $method
            );
            return;
        }

        // 业务数据返回
        $data = call_user_func_array(array($instance,$method), array($this->request));
        $this->response->content = $data;
    }

    public function setResponse($code, $errCode, $msg = "")
    {
        $this->response->code = $code;
        $this->response->errCode = $errCode;
        $this->response->content = $msg;
    }

    // commit response to pepperbus
    public function commit()
    {
        header("Code: {$this->response->code}");
        header("Error-Code: {$this->response->errCode}");
        header("Body-Length: " . mb_strlen($this->response->content));
        echo $this->response->content;
    }
}

