<!--
 * @Author: FunctionSir
 * @License: AGPLv3
 * @Date: 2024-11-10 20:33:19
 * @LastEditTime: 2024-11-10 20:41:41
 * @LastEditors: FunctionSir
 * @Description: -
 * @FilePath: /HttpPollingTunnel/README.md
-->

# HttpPollingTunnel

A tunnel, with http and polling... Super slow.

## 这个实在是个很慢的东西

属于是一个验证自己猜想的产物.
能用, 也的确是血统纯正的不行的HTTP.
不过速度... 最快也就是1M宽带时期的速率了.

## 看呐! 漫天飞舞的HTTP请求

纯纯的HTTP, 漫天飞舞.

## 使用方法

1. 设置服务器(配置文件样例在源码中附赠了).
2. 设置客户端(配置文件样例也附赠了).
3. 设定客户端的TUN.

对于Linux, 样例:

在A机器上:

``` bash
sudo ip addr add 10.1.0.10/24 dev ht0 && sudo ip link set dev ht0 up
```

在B机器上:

``` bash
sudo ip addr add 10.1.0.11/24 dev ht0 && sudo ip link set dev ht0 up
```

注意不要设一样的IP.

然后, 可以试试在A机器上:

``` bash
ping 10.1.0.10 -I ht0
```
