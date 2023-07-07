## k8s-grpc-go
实现grpc-go在k8s中动态负载均衡
### 1. `client-go` + `Custom Name Resolution`
	自定义name resolver，通过client-go动态监听指定endpoint变更并更新resolver连接状态，实现动态负载均衡

### 2. 修改默认dns resolver源代码
**变量:**

- [minDNSResRate](https://github.com/grpc/grpc-go/blob/11feb0a9afd844fd2ab1f18dca02ad6a344b21bf/internal/resolver/dns/dns_resolver.go#L87): 默认解析间隔时间30s

**dnsResolver内部方法:**

- [ResolveNow](https://github.com/grpc/grpc-go/blob/11feb0a9afd844fd2ab1f18dca02ad6a344b21bf/internal/resolver/dns/dns_resolver.go#L201): 客户端连接失败后触发立即解析

- [watcher](https://github.com/grpc/grpc-go/blob/11feb0a9afd844fd2ab1f18dca02ad6a344b21bf/internal/resolver/dns/dns_resolver.go#L214): 重新解析和更新客户端连接状态

默认情况下当客户端连接失败时执行ResolveNow控制watcher立即解析，并间隔执行。

启动客户端后当pod副本数增多时客户端无感知。调整watcher函数实现默认间隔解析、客户端连接失败后触发立即解析。
```go
func (d *dnsResolver) watcher() {
	defer d.wg.Done()
	backoffIndex := 1
	for {
		state, err := d.lookup()
		if err != nil {
			// Report error to the underlying grpc.ClientConn.
			d.cc.ReportError(err)
		} else {
			err = d.cc.UpdateState(*state)
		}

		var timer *time.Timer
		if err == nil {
			// Success resolving, wait for the next ResolveNow. However, also wait 30
			// seconds at the very least to prevent constantly re-resolving.
			backoffIndex = 1
			timer = newTimerDNSResRate(minDNSResRate)
			select {
			case <-d.ctx.Done():
				timer.Stop()
				return
			//case <-d.rn:        //
			default:              //
			}
		} else {
			// Poll on an error found in DNS Resolver or an error received from
			// ClientConn.
			timer = newTimer(backoff.DefaultExponential.Backoff(backoffIndex))
			backoffIndex++
		}
		select {
		case <-d.rn:             //
		case <-d.ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}
```
