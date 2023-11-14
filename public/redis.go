package public

import (
	"FGateWay/golang_common/lib"
	"github.com/garyburd/redigo/redis"
)

// RedisConfPipline 是一个接受多个 Redis 操作的函数，它使用 "default" 配置来创建一个 Redis 连接。
// 它会在这个连接上执行所有传入的操作，最后关闭这个连接。
func RedisConfPipline(pip ...func(c redis.Conn)) error {
	c, err := lib.RedisConnFactory("default") // 使用 "default" 配置创建一个 Redis 连接。
	if err != nil {
		return err // 如果创建连接时出现错误，则返回错误。
	}
	defer c.Close() // 确保在函数返回时关闭连接。
	for _, f := range pip {
		f(c) // 对于传入的每一个操作，都在创建的连接上执行。
	}
	c.Flush()  // 清空所有之前执行的命令的缓冲区。
	return nil // 所有操作成功执行后返回 nil。
}

// RedisConfDo 是一个接受一个 Redis 命令名和其参数的函数，它使用 "default" 配置来创建一个 Redis 连接。
// 它会在这个连接上执行传入的命令，然后返回命令的结果，最后关闭这个连接。
func RedisConfDo(commandName string, args ...interface{}) (interface{}, error) {
	c, err := lib.RedisConnFactory("default") // 使用 "default" 配置创建一个 Redis 连接。
	if err != nil {
		return nil, err // 如果创建连接时出现错误，则返回 nil 和错误。
	}
	defer c.Close()                   // 确保在函数返回时关闭连接。
	return c.Do(commandName, args...) // 在创建的连接上执行传入的命令，并返回结果。
}
