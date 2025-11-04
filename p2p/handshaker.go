package p2p

// HandshakeFunc 相当于func的一个接口, 方便具体的某个func接受这一系列的func实现
// 或者作为普通interface, 某个struct中的成员变量
type HandshakeFunc func(any) error

func NopHandshakeFunc(any) error {
	return nil
}
