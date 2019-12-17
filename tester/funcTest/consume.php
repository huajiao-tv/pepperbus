<?php
ini_set('date.timezone','Asia/Shanghai');
// @require server sdk
require_once((dirname(__FILE__)).'/sdk/server.php');
$treeServer = new PepperBusServer();
$treeServer->register("testSuccessQueue/topic", "ProcessJob", "ExecuteSuccess");
$treeServer->register("testSuccessQueue/topicForAddDel", "ProcessJob", "ExecuteAddDel");
$treeServer->register("testUnsubscribedQueue/topic", "ProcessJob", "ExecuteNotFound");
$treeServer->register("test500Queue/topic", "ProcessJob", "ExecuteFail");

$treeServer->parseRequest();
if ($treeServer->request->isQueueRequest()) {
    try {
        $treeServer->process();
    } catch (Exception $e) {
	if($e->getCode() == 4004) {
            $treeServer->setResponse(Response::CODE_NOT_FOUND, $e->getCode(), $e->getMessage());
	} else {
            $treeServer->setResponse(Response::CODE_FAIL, $e->getCode(), $e->getMessage());
	}
    }
} else {
    $treeServer->setResponse(Response::CODE_NOT_FOUND, Response::CODE_NOT_FOUND, "cgi process not subscribe to this queue");
}
$treeServer->commit();

/**
 * Created by zhoukunta@qq.com
 * User: johntech
 * Date: 02/08/2018
 * Time: 4:06 PM
 */
class ProcessJob
{
    const MY_ERROR_CODE = 1000;
    const MY_NOTFOUND_CODE = 4004;
    const MY_ERROR_MSG = "ProcessJob fail flag";

    public function SetRedis($redis_addr, $key, $val)
    {
        $arr = explode(":", $redis_addr);
        $redis = new \Redis();
        $redis->connect($arr[0], $arr[1], 3);
        
        if (count($arr) == 3) {
            $redis->auth($arr[2]);
        }

	// 测试程序不处理异常，如果所有都正常情况下，自己排查无法连接redis和写入的原因
        $redis->set($key, $val);
    }

    public function ExecuteSuccess($request)
    {
        $jobs = $request->getJobs();
        
        $arr = explode("-", $jobs[0]["content"]);
        if(count($arr) != 2) {
            // 出错不处理，调用测试consume时，向testSuccessQueue中加job时，需要给出用于存储执行结果的redis地址
            return 0;
        }

	$key = $arr[0];
	$redis_addr = $arr[1];

        $now = time();

        self::SetRedis($redis_addr, $key, $now);

        return 0;
    }

    public function ExecuteAddDel($request)
    {
        $jobs = $request->getJobs();

        $arr = explode("-", $jobs[0]["content"]);
	$key = $arr[0];
	$redis_addr = $arr[1];

        $now = time();

        self::SetRedis($redis_addr, $key."AddDel", $now);
        return 0;
    }

    public function ExecuteFail($request)
    {
        $jobs = $request->getJobs();
        throw new Exception(self::MY_ERROR_MSG, self::MY_ERROR_CODE);
        return;
    }


    public function ExecuteNotFound($request)
    {
        throw new Exception("", self::MY_NOTFOUND_CODE);
        return;
    }
}
