<?php
// @require server sdk
require_once(dirname(dirname(dirname(dirname(__FILE__)))).'/sdk/php/server.php');
$treeServer = new PepperBusServer();
$treeServer->register("BenTestQueue0/BenTestTopic0", "ProcessJob", "ExecuteSuccess");
$treeServer->register("BenTestQueue0/BenTestQueue1", "ProcessJob", "ExecuteFail");

$treeServer->parseRequest();
if ($treeServer->request->isQueueRequest()) {
    try {
        $treeServer->process();
    } catch (Exception $e) {
        $treeServer->setResponse(Response::CODE_FAIL, $e->getCode(), $e->getMessage());
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
    // user error code should not equal to zero
    const MY_ERROR_CODE = 1000;
    const MY_ERROR_MSG = "ProcessJob fail flag";

    public function ExecuteSuccess($request)
    {
        $jobs = $request->getJobs();
        //var_dump($jobs);
        return 0;
    }

    public function ExecuteFail($request)
    {
        $jobs = $request->getJobs();
        //var_dump($jobs);
        throw new Exception(self::MY_ERROR_MSG, self::MY_ERROR_CODE);
        return;
    }
}
