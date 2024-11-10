/*
 * @Author: FunctionSir
 * @License: AGPLv3
 * @Date: 2024-11-10 00:02:53
 * @LastEditTime: 2024-11-10 00:10:43
 * @LastEditors: FunctionSir
 * @Description: -
 * @FilePath: /HttpPollingTunnel/tool/main.go
 */

package main

import (
	"crypto/sha512"
	"fmt"
)

func calcAuthHash(salt, key string) string {
	return fmt.Sprintf("%X", sha512.Sum512([]byte(salt+key)))
}

func main() {
	var salt string
	var key string
	fmt.Println("Input salt, key, separate by enter:")
	fmt.Scanf("%s", &salt)
	fmt.Scanf("%s", &key)
	fmt.Println("Hash for auth is:")
	fmt.Println(calcAuthHash(salt, key))
}
