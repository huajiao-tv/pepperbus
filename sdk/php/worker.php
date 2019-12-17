<?php
ini_set('max_execution_time', 0);
ini_set('default_socket_timeout', -1);
ini_set('memory_limit',"512M");
set_time_limit(0);
// tick use required
declare(ticks = 1);


class PepperBusWorker
{
    const REDIS_TIMEOUT = 3;

    private $stopping = false;
    private $queue = "";
    private $handler = [];
    private $maxTask = 1;

    // connection conf
    private $host = "127.0.0.1";
    private $port = 0;
    private $password = "";

    public function __construct()
    {
        // register signal
        pcntl_signal(SIGINT, [&$this, "signalHandler"]);
    }

    public function signalHandler($sig)
    {
        if ($sig == SIGINT) {
            $this->stopping = true;
        }
    }

    public function run()
    {
        $this->_parseArgvs();
        $this->_runWorker();
    }

    private function getRedis()
    {
        $redis = new redis();
        $redis->connect($this->host, $this->port, self::REDIS_TIMEOUT, null, 200);
        if (!empty($this->password)) {
            $redis->auth($this->password);
        }
        return $redis;
    }

    private function _parseArgvs()
    {
        $opts = getopt("q:h:p:a:t:H:");
        foreach ($opts as $opt => $value) {
            switch ($opt) {
                case "q":
                    $this->queue = $value;
                    break;
                case "h":
                    $this->host = $value;
                    break;
                case "p":
                    $this->port = $value;
                    break;
                case "a":
                    $this->password = $value;
                    break;
                case "t":
                    $this->maxTask = $value;
                    break;
                case "H":
                    $this->handler = explode(",", $value);
                    break;
            }
        }
        $this->password = $this->queue.":".$this->password;
    }

    private function _runWorker()
    {
        $i = 0;
        while ($i++ < $this->maxTask) {
            if ($this->stopping) {
                return;
            }

            try {
                $job = $this->getJob();
                if (!$job) {
                    sleep(rand(3, 10));
                    continue;
                }
                $this->_runQueueTask($this->handler, $job);
            } catch (\Exception $e) {
                if (!empty($job)) {
                    $this->addRescueQueue($job);
                }
                sleep(rand(3, 10));//出错时消费得慢一点
                continue;
            }
        }
    }

    private function getJob()
    {
        $redis = $this->getRedis();
        $data = $redis->rpop($this->queue);
        return json_decode($data, true);
    }

    private function addRescueQueue($job)
    {
        $job["retrying"] = true;
        $redis = $this->getRedis();
        return $redis->lPush($this->queue, json_encode($job));
    }

    private function _runQueueTask($handler, $value)
    {
        $worker = array();
        if ($handler) {
            $worker = $handler;
            $method = new \ReflectionMethod($handler[0], $handler[1]);
            if(!$method->isStatic()){
                $class=new \ReflectionClass($handler[0]);
                $worker[0] = $class->newInstance();
            }
        }

        if (!is_callable($worker)) {
            throw new \Exception(json_encode($worker) . " is invalid");//may be 'class not found'
        }

        $start_time = microtime(true);
        call_user_func($handler, $value);

        $consume = round((microtime(true) - $start_time) * 1000);

//        $this->_task["status"] = 'run worker function done';
        $this->_runTaskDone($value, $consume);
    }

    private function _runTaskDone($value, $consume = 0)
    {
//        self::addLog(array("consume" => $consume));
//
//        if ($this->_logDays > 0) {
//            $this->_log(json_encode($value));
//        }
//
//        if ($this->_debug) {
//            $this->_print(json_encode($value));
//        }
//
//        $this->_task["status"] = 'consumed';
    }
}
