/*redis返回数据
Error: 			-Error message\r\n

Simple String: 	+OK\r\n

Integers: 		:1000\r\n(int64)

Bulk Strings: 	$6\r\nfoobar\r\n
				$0\r\n\r\n
				$-1\r\n

Arrays:			*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n
				*0\r\n
				*-1\r\n
A client sends to the Redis server a RESP Array consisting of just Bulk Strings.
A Redis server replies to clients sending any valid RESP data type as reply.
*/
package msgRedis
