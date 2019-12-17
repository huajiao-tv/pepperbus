<?php
/**
 * Created by zhoukunta@qq.com.
 * User: johntech
 * Date: 02/08/2018
 * Time: 4:36 PM
 */
require_once(dirname(dirname(dirname(dirname(__FILE__)))).'/sdk/php/client.php');

try {
    // add job to pepperbus
    $jobId = PepperBusClient::getInstance("queue1")->addJob("conntent from produce");
    var_dump($jobId);
} catch (Exception $e) {
    var_dump($e);
}
