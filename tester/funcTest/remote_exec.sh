#!/usr/bin/expect
set ip [lindex $argv 0]
set user [lindex $argv 1]
set pass [lindex $argv 2]
set cmd [lindex $argv 3]

spawn ssh "$user@$ip" $cmd
expect "*password:"
send "$pass\r"
expect "*#"
