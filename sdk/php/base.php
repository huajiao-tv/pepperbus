<?php
/**
 * Created by PhpStorm.
 * User: johntech
 * Date: 02/08/2017
 * Time: 5:31 PM
 */
// @todo log and trace info when added success
// @todo Queue Name add product and service tag in extend params

class Request
{
    const TypePing = "ping";
    const TypeData = "data";

    public $type;
    public $jobs;
    public $outTraceId;
    public $queue;
    public $topic;
    public $isQueueRequest;
    public function __construct()
    {
        $this->isQueueRequest = !empty($_POST["IS_PEPPER_BUS_REQUEST"]) ?
            $_POST["IS_PEPPER_BUS_REQUEST"] === "true" : false;
        
        $this->type = !empty($_POST["type"]) ? $_POST["type"] : "";

        switch ($this->type) {
        case self::TypeData:
            $this->outTraceId = !empty($_POST["outTraceId"]) ? $_POST["outTraceId"] : "";
            $this->queue = !empty($_POST["queue"]) ? $_POST["queue"] : "";
            $this->topic = !empty($_POST["topic"]) ? $_POST["topic"] : "";
            $this->jobs = !empty($_POST["jobs"]) ? json_decode($_POST["jobs"], true) : "";
            break;
        }
    }

    // queuePath contains queueName and queue topic which joint by "/"
    public function getQueuePath()
    {
        if (empty($this->queue)) {
            return '';
        }

        if (empty($this->topic)) {
            $this->topic = 'default';
        }

        return $this->queue."/".$this->topic;
    }

    // get job from default channel of queue
    public function getJobs()
    {
        return $this->jobs;
    }

    public function isQueueRequest()
    {
        return $this->isQueueRequest;
    }
}

class Response
{
    const CODE_FAIL = 500;
    const CODE_OK = 200;
    const CODE_NOT_FOUND = 404;
    const EXEC_SUCCESS = 0;

    public $outTraceId = 0;
    public $code = 0;
    public $content = "";
    public $errCode = 0;

    public function __construct()
    {
        $this->code = self::CODE_OK;
        $this->errCode = self::EXEC_SUCCESS;
    }

    public function getString()
    {
        $result = array(
            "OutTraceId"=> $this->outTraceId,
            "code"=>$this->code,
            "errCode"=>$this->errCode,
            "content"=>$this->content,
        );
        return json_encode($result);
    }
}
