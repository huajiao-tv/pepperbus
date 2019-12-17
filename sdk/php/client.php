<?php
/**
 * Created by PhpStorm.
 * User: johntech
 * Date: 02/08/2017
 * Time: 5:31 PM
 */


/**
 * Created by zhoukunta@qq.com.
 * User: johntech
 * Date: 02/08/2017
 * Time: 3:21 PM
 */
// @todo all functions add params check
class PepperBusClient
{
    private static $QUEUE_CONF = [
        "queue1" => [
            "host" => "127.0.0.1", "port" => 12018, "timeout" => 3, "password" => "783ab0ce"
        ],
    ];

    private $queue;
    private $redis;

    private function __construct($queue)
    {
        $config = self::$QUEUE_CONF[$queue];
        if (!$config) {
            throw new Exception("queue not found");
        }
        $this->queue = $queue;
        $this->redis = new redis();
        $this->redis->connect($config['host'], $config['port'], $config['timeout'], null, 200);
        if (!empty($config['password'])) {
            try {
                $ret = $this->redis->auth($queue.":".$config['password']);
                var_dump($ret);
            } catch (Exception $e) {
                var_dump($e->getMessage());
            }
        }
    }

    public static function getInstance($queue)
    {
        return new self($queue);
    }

    // @todo lpush give auth params in key or in value field
    public function addJob($content)
    {
        return $this->redis->rawCommand("lpush", $this->queue."/default", $content);
    }
}
